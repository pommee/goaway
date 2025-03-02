package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"goaway/internal/blacklist"
	"goaway/internal/database"
	"goaway/internal/logging"
	"goaway/internal/settings"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
)

var log = logging.GetLogger()
var rcodes = map[int]string{
	dns.RcodeSuccess:        "NoError",
	dns.RcodeFormatError:    "FormErr",
	dns.RcodeServerFailure:  "ServFail",
	dns.RcodeNameError:      "NXDomain",
	dns.RcodeNotImplemented: "NotImp",
	dns.RcodeRefused:        "Refused",
	dns.RcodeYXDomain:       "YXDomain",
	dns.RcodeYXRrset:        "YXRRSet",
	dns.RcodeNXRrset:        "NXRRSet",
	dns.RcodeNotAuth:        "NotAuth",
	dns.RcodeNotZone:        "NotZone",
	dns.RcodeBadSig:         "BADSIG",
	dns.RcodeBadKey:         "BADKEY",
	dns.RcodeBadTime:        "BADTIME",
	dns.RcodeBadMode:        "BADMODE",
	dns.RcodeBadName:        "BADNAME",
	dns.RcodeBadAlg:         "BADALG",
	dns.RcodeBadTrunc:       "BADTRUNC",
	dns.RcodeBadCookie:      "BADCOOKIE",
}
var arpTable = map[string]string{}
var dbMutex sync.Mutex
var wsMutex sync.Mutex

const batchSize = 100

type MacVendor struct {
	Vendor string `json:"company"`
}

type DNSServer struct {
	Config              settings.DNSServerConfig
	Blacklist           blacklist.Blacklist
	DB                  *sql.DB
	Counters            CounterDetails
	StatisticsRetention int
	lastLogTime         time.Time
	logIntervalSeconds  int
	cache               sync.Map
	WebServer           *gin.Engine
	logEntryChannel     chan RequestLogEntry
	WS                  *websocket.Conn
}

type QueryResponse struct {
	IPAddresses []dns.RR
	Ttl         uint32
	Status      int
}

type cachedRecord struct {
	IPAddresses []dns.RR
	ExpiresAt   time.Time
}

type CounterDetails struct {
	AllowedRequests int `json:"allowed_requests"`
	BlockedRequests int `json:"blocked_requests"`
}

type RequestLogEntry struct {
	Timestamp      time.Time     `json:"timestamp"`
	Domain         string        `json:"domain"`
	IP             []string      `json:"ip"`
	Blocked        bool          `json:"blocked"`
	Cached         bool          `json:"cached"`
	ResponseTimeNS time.Duration `json:"responseTimeNS"`
	ClientInfo     *Client       `json:"client"`
	Status         string        `json:"status"`
	QueryType      string        `json:"queryType"`
}

type Request struct {
	w         dns.ResponseWriter
	msg       *dns.Msg
	question  dns.Question
	timestamp time.Time
	client    *Client
}

type Client struct {
	IP   string
	Name string
	MAC  string
}

func NewDNSServer(config *settings.DNSServerConfig) (*DNSServer, error) {
	db, err := database.Initialize()
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	blacklist := blacklist.Blacklist{
		DB: db.Con,
		BlocklistURL: map[string]string{
			"StevenBlack": "https://raw.githubusercontent.com/StevenBlack/hosts/refs/heads/master/hosts",
		},
	}

	if err := blacklist.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize blacklist: %w", err)
	}

	if err := blacklist.InitializeBlocklist("Custom", ""); err != nil {
		return nil, fmt.Errorf("failed to initialize custom blocklist: %w", err)
	}

	blacklist.BlocklistURL, err = blacklist.GetBlocklistUrls()
	if err != nil {
		log.Error("Failed to get blocklist URLs: %v", err)
	}

	server := &DNSServer{
		Config:              *config,
		Blacklist:           blacklist,
		DB:                  db.Con,
		StatisticsRetention: config.StatisticsRetention,
		lastLogTime:         time.Now(),
		logIntervalSeconds:  1,
		cache:               sync.Map{},
		logEntryChannel:     make(chan RequestLogEntry, 1000),
	}

	return server, nil
}

func (s *DNSServer) Init() (int, *dns.Server) {
	server := &dns.Server{
		Addr:      fmt.Sprintf(":%d", s.Config.Port),
		Net:       "udp",
		Handler:   s,
		UDPSize:   65535,
		ReusePort: true,
	}

	domains, _ := s.Blacklist.CountDomains()
	return domains, server
}

func (s *DNSServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	timestamp := time.Now()
	clientIP, clientName, macAddress := s.getClientInfo(w.RemoteAddr().String())

	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true
	msg.RecursionAvailable = true

	var wg sync.WaitGroup
	results := make(chan RequestLogEntry, len(r.Question))

	for _, question := range r.Question {
		wg.Add(1)
		go func(question dns.Question) {
			defer wg.Done()
			client := &Client{IP: clientIP, Name: clientName, MAC: macAddress}
			entry := s.processQuery(&Request{w, msg, question, timestamp, client})
			entryWSJson, _ := json.Marshal(entry)
			if s.WS != nil {
				wsMutex.Lock()
				s.WS.WriteMessage(websocket.TextMessage, []byte(entryWSJson))
				wsMutex.Unlock()
			}
			results <- entry
		}(question)
	}

	wg.Wait()
	close(results)

	for entry := range results {
		s.logEntryChannel <- entry
	}
}

func getMacAddress(ip string) string {
	mac, exists := arpTable[ip]
	if exists {
		return mac
	}

	return "unknown"
}

func (s *DNSServer) GetVendor(mac string) (string, error) {
	var vendor string
	dbMutex.Lock()
	defer dbMutex.Unlock()

	err := s.DB.QueryRow("SELECT vendor FROM mac_addresses WHERE mac = ?", mac).Scan(&vendor)
	if err == sql.ErrNoRows {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return vendor, nil
}

func (s *DNSServer) SaveMacVendor(clientIP, mac, vendor string) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	log.Debug("Saving new MAC address: %s %s", mac, vendor)
	_, err := s.DB.Exec("INSERT INTO mac_addresses (ip, mac, vendor) VALUES (?, ?, ?)", clientIP, mac, vendor)
	return err
}

func (s *DNSServer) getClientInfo(remoteAddr string) (string, string, string) {
	clientIP := strings.Split(remoteAddr, ":")[0]
	macAddress := getMacAddress(clientIP)

	vendor, err := s.GetVendor(macAddress)
	if macAddress != "unknown" {
		if err != nil || vendor == "" {
			log.Debug("Lookup vendor for mac %s", macAddress)
			vendor, err = lookupMacVendor(macAddress)
			if err == nil {
				s.SaveMacVendor(clientIP, macAddress, vendor)
			} else {
				log.Warning("Error while lookup mac address vendor: %v", err)
			}
		}
	}

	if clientIP == "127.0.0.1" || clientIP == "::1" {
		if h, err := os.Hostname(); err == nil {
			return clientIP, h, macAddress
		}
		return clientIP, "localhost", macAddress
	}

	if hostnames, err := net.LookupAddr(clientIP); err == nil && len(hostnames) > 0 {
		return clientIP, strings.TrimSuffix(hostnames[0], "."), macAddress
	}

	return clientIP, "unknown", macAddress
}

func lookupMacVendor(mac string) (string, error) {
	url := fmt.Sprintf("https://api.maclookup.app/v2/macs/%s", mac)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch MAC vendor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		Success bool   `json:"success"`
		Found   bool   `json:"found"`
		Company string `json:"company"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Found {
		return result.Company, nil
	}

	return "", fmt.Errorf("vendor not found")
}

func (s *DNSServer) processQuery(request *Request) RequestLogEntry {
	if request.question.Qtype == dns.TypePTR && strings.HasSuffix(request.question.Name, ".in-addr.arpa.") {
		return s.handlePTRQuery(request)
	}

	isBlacklisted, err := s.Blacklist.IsBlacklisted(strings.TrimSuffix(request.question.Name, "."))
	if err != nil {
		log.Error("%v", err)
	}
	if isBlacklisted {
		return s.handleBlacklisted(request)
	}
	return s.handleQuery(request)
}

func (s *DNSServer) handlePTRQuery(request *Request) RequestLogEntry {
	ptrName := request.question.Name
	ipParts := strings.TrimSuffix(ptrName, ".in-addr.arpa.")
	parts := strings.Split(ipParts, ".")

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	ipStr := strings.Join(parts, ".")

	if ipStr == "127.0.0.1" {
		ptr := &dns.PTR{
			Hdr: dns.RR_Header{
				Name:   request.question.Name,
				Rrtype: dns.TypePTR,
				Class:  dns.ClassINET,
				Ttl:    3600,
			},
			Ptr: "localhost.lan.",
		}

		request.msg.Answer = append(request.msg.Answer, ptr)
		_ = request.w.WriteMsg(request.msg)

		return RequestLogEntry{
			Timestamp:      request.timestamp,
			Domain:         request.question.Name,
			IP:             []string{"localhost.lan"},
			Blocked:        false,
			Cached:         false,
			ResponseTimeNS: time.Since(request.timestamp),
			ClientInfo:     request.client,
			Status:         "NoError",
			QueryType:      "PTR",
		}
	}

	hostname := "unknown"

	rows, err := s.DB.Query("SELECT client_name FROM request_log WHERE client_ip = ? AND client_name != 'unknown' LIMIT 1", ipStr)
	if err == nil {
		defer rows.Close()
		if rows.Next() {
			rows.Scan(&hostname)
			hostname = strings.TrimSuffix(hostname, ".")
		}
	}

	if hostname == "unknown" {
		if names, err := net.LookupAddr(ipStr); err == nil && len(names) > 0 {
			hostname = strings.TrimSuffix(names[0], ".")
		}
	}

	if hostname != "unknown" {
		ptr := &dns.PTR{
			Hdr: dns.RR_Header{
				Name:   request.question.Name,
				Rrtype: dns.TypePTR,
				Class:  dns.ClassINET,
				Ttl:    3600,
			},
			Ptr: hostname + ".",
		}

		request.msg.Answer = append(request.msg.Answer, ptr)
		_ = request.w.WriteMsg(request.msg)

		return RequestLogEntry{
			Timestamp:      request.timestamp,
			Domain:         request.question.Name,
			IP:             []string{hostname},
			Blocked:        false,
			Cached:         false,
			ResponseTimeNS: time.Since(request.timestamp),
			ClientInfo:     request.client,
			Status:         "NoError",
			QueryType:      "PTR",
		}
	}

	return s.handleQuery(request)
}

func (s *DNSServer) handleQuery(request *Request) RequestLogEntry {
	answers, cached, status := s.resolve(request.question.Name, request.question.Qclass)
	request.msg.Answer = append(request.msg.Answer, answers...)

	if status == "NXDomain" {
		request.msg.Rcode = dns.RcodeNameError
	} else if status == "ServFail" {
		request.msg.Rcode = dns.RcodeServerFailure
	}

	var resolvedAddresses []string
	if len(answers) > 0 {
		for _, answer := range answers {
			switch rec := answer.(type) {
			case *dns.A:
				resolvedAddresses = append(resolvedAddresses, rec.A.String())
			case *dns.AAAA:
				resolvedAddresses = append(resolvedAddresses, rec.AAAA.String())
			case *dns.PTR:
				resolvedAddresses = append(resolvedAddresses, rec.Ptr)
			case *dns.CNAME:
				resolvedAddresses = append(resolvedAddresses, rec.Target)
			}
		}
	}

	_ = request.w.WriteMsg(request.msg)
	s.Counters.AllowedRequests++

	return RequestLogEntry{
		Timestamp:      request.timestamp,
		Domain:         request.question.Name,
		IP:             resolvedAddresses,
		Blocked:        false,
		Cached:         cached,
		ResponseTimeNS: time.Since(request.timestamp),
		ClientInfo:     request.client,
		Status:         status,
		QueryType:      dns.TypeToString[request.question.Qtype],
	}
}

func (s *DNSServer) resolve(domain string, qtype uint16) ([]dns.RR, bool, string) {
	cacheKey := fmt.Sprintf("%s-%d", domain, qtype)
	if cached, found := s.cache.Load(cacheKey); found {
		if ipAddresses, valid := s.getCachedRecord(cached); valid {
			return ipAddresses, true, "NoError"
		}
	}

	answers, ttl, status := s.resolveCNAMEChain(domain, qtype, make(map[string]bool))
	if len(answers) > 0 {
		s.cacheRecord(cacheKey, answers, ttl)
	}

	return answers, false, status
}

func (s *DNSServer) resolveCNAMEChain(domain string, qtype uint16, visited map[string]bool) ([]dns.RR, uint32, string) {
	if visited[domain] {
		return nil, 0, "SERVFAIL"
	}
	visited[domain] = true

	answers, ttl, status := s.queryUpstream(domain, qtype)
	if len(answers) == 0 {
		cnameAnswers, cnameTTL, cnameStatus := s.queryUpstream(domain, dns.TypeCNAME)
		if len(cnameAnswers) > 0 {
			for _, answer := range cnameAnswers {
				if cname, ok := answer.(*dns.CNAME); ok {
					targetAnswers, targetTTL, targetStatus := s.resolveCNAMEChain(cname.Target, qtype, visited)
					if len(targetAnswers) > 0 {
						minTTL := cnameTTL
						if targetTTL < minTTL {
							minTTL = targetTTL
						}
						return append(cnameAnswers, targetAnswers...), minTTL, targetStatus
					}
					return cnameAnswers, cnameTTL, cnameStatus
				}
			}
		}
	}
	return answers, ttl, status
}

func (s *DNSServer) getCachedRecord(cached interface{}) ([]dns.RR, bool) {
	cachedRecord := cached.(cachedRecord)
	if time.Now().Before(cachedRecord.ExpiresAt) {
		return cachedRecord.IPAddresses, true
	}
	return nil, false
}

func (s *DNSServer) cacheRecord(domain string, ipAddresses []dns.RR, ttl uint32) {
	cacheTTL := s.Config.CacheTTL
	if ttl > 0 {
		cacheTTL = time.Duration(ttl) * time.Second
	}
	s.cache.Store(domain, cachedRecord{
		IPAddresses: ipAddresses,
		ExpiresAt:   time.Now().Add(cacheTTL),
	})
}

func (s *DNSServer) queryUpstream(domain string, qtype uint16) ([]dns.RR, uint32, string) {
	var (
		ipAddresses []dns.RR
		ttl         uint32
		status      = "SERVFAIL"
	)
	done := make(chan struct{})

	go func() {
		defer close(done)
		c := new(dns.Client)

		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(domain), qtype)
		m.RecursionDesired = true

		in, _, err := c.Exchange(m, s.Config.PreferredUpstream)
		if err != nil {
			log.Error("Resolution error for domain (%s): %v", domain, err)
			return
		}

		if statusStr, ok := rcodes[in.Rcode]; ok {
			status = statusStr
		}

		ipAddresses = append(ipAddresses, in.Answer...)
		if len(in.Answer) > 0 && ttl == 0 {
			ttl = in.Answer[0].Header().Ttl
		}

		ipAddresses = append(ipAddresses, in.Ns...)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		log.Warning("DNS lookup for %s timed out", domain)
		status = "SERVFAIL"
	}

	return ipAddresses, ttl, status
}

func (s *DNSServer) handleBlacklisted(request *Request) RequestLogEntry {
	log.Info("Blocked: %s", request.question.Name)

	request.msg.Rcode = dns.RcodeSuccess
	var status = "Blacklisted"

	rr4 := &dns.A{
		Hdr: dns.RR_Header{
			Name:   request.question.Name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    uint32(s.Config.CacheTTL.Seconds()),
		},
		A: net.ParseIP("0.0.0.0"),
	}

	rr6 := &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   request.question.Name,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    uint32(s.Config.CacheTTL.Seconds()),
		},
		AAAA: net.ParseIP("::"),
	}

	request.msg.Answer = append(request.msg.Answer, rr4, rr6)
	_ = request.w.WriteMsg(request.msg)

	s.Counters.BlockedRequests++

	return RequestLogEntry{
		Timestamp:      request.timestamp,
		Domain:         request.question.Name,
		IP:             []string{""},
		Blocked:        true,
		Cached:         false,
		ResponseTimeNS: time.Since(request.timestamp),
		ClientInfo:     request.client,
		Status:         status,
	}
}

func (s *DNSServer) LoadRequestLog() ([]RequestLogEntry, error) {
	rows, err := s.DB.Query("SELECT timestamp, domain, ip, blocked, cached, response_time_ns, client_ip, client_name, status FROM request_log")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []RequestLogEntry
	for rows.Next() {
		var entry RequestLogEntry
		var clientIP, clientName, mac sql.NullString
		if err := rows.Scan(&entry.Timestamp, &entry.Domain, &entry.IP, &entry.Blocked, &entry.Cached, &entry.ResponseTimeNS, &clientIP, &clientName); err != nil {
			return nil, err
		}
		entry.ClientInfo = &Client{IP: clientIP.String, Name: clientName.String, MAC: mac.String}
		logs = append(logs, entry)
	}
	return logs, nil
}

func (s *DNSServer) ProcessLogEntries() {
	var batch []RequestLogEntry
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case entry := <-s.logEntryChannel:
			batch = append(batch, entry)
			if len(batch) >= batchSize {
				s.saveBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.saveBatch(batch)
				batch = nil
			}
		}
	}
}

func (s *DNSServer) ProcessARPTable() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Update first time server is started
	updateARPTable()

	for {
		select {
		case <-ticker.C:
			updateARPTable()
		}
	}
}

func updateARPTable() {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		fmt.Println("Error running ARP command:", err)
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.ReplaceAll(line, "(", "")
		line = strings.ReplaceAll(line, ")", "")

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			ip := fields[1]
			mac := fields[3]
			arpTable[ip] = mac
		}
	}
}

func (s *DNSServer) saveBatch(entries []RequestLogEntry) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	tx, err := s.DB.Begin()
	if err != nil {
		log.Error("Could not start database transaction %v", err)
		return
	}
	defer func() {
		if err := tx.Commit(); err != nil {
			log.Warning("DB commit error %v", err)
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO request_log (timestamp, domain, ip, blocked, cached, response_time_ns, client_ip, client_name, status, query_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Error("Could not create a prepared statement for request logs %v", err)
		return
	}
	defer stmt.Close()

	for _, entry := range entries {
		if _, err := stmt.Exec(
			entry.Timestamp.Unix(),
			entry.Domain,
			strings.Join(entry.IP, ","),
			entry.Blocked,
			entry.Cached,
			entry.ResponseTimeNS,
			entry.ClientInfo.IP,
			entry.ClientInfo.Name,
			entry.Status,
			entry.QueryType,
		); err != nil {
			log.Error("Could not save request log. Reason: %v", err)
			return
		}
	}
}

func (s *DNSServer) ClearOldEntries() {
	const (
		maxRetries      = 10
		retryDelay      = 150 * time.Millisecond
		cleanupInterval = 1 * time.Minute
	)

	for {
		requestThreshold := ((60 * 60) * 24) * s.StatisticsRetention
		log.Debug("Running next cleanup in %s", time.Now().Add(cleanupInterval).Format("15:04:05"))
		time.Sleep(cleanupInterval)

		for retryCount := 0; retryCount < maxRetries; retryCount++ {
			result, err := s.DB.Exec(fmt.Sprintf("DELETE FROM request_log WHERE strftime('%%s', 'now') - timestamp > %d", requestThreshold))
			if err != nil {
				if err.Error() == "database is locked" {
					log.Debug("Database is locked; retrying (%d/%d)", retryCount+1, maxRetries)
					time.Sleep(retryDelay)
					continue
				}
				log.Error("Failed to clear old entries: %s", err)
				break
			}

			if affected, err := result.RowsAffected(); err == nil && affected > 0 {
				log.Debug("Cleared %d old entries", affected)
			}
			s.UpdateCounters()
			break
		}
	}
}

func (s *DNSServer) GetCounters() (int, int, error) {
	var blockedCount, allowedCount int

	err := s.DB.QueryRow("SELECT COUNT(*) FROM request_log WHERE blocked = 1").Scan(&blockedCount)
	if err != nil {
		log.Error("Failed to get blocked requests count: %s", err)
		return 0, 0, err
	}

	err = s.DB.QueryRow("SELECT COUNT(*) FROM request_log WHERE blocked = 0").Scan(&allowedCount)
	if err != nil {
		log.Error("Failed to get allowed requests count: %s", err)
		return 0, 0, err
	}

	return blockedCount, allowedCount, nil
}

func (s *DNSServer) UpdateCounters() {
	blockedCount, allowedCount, err := s.GetCounters()
	if err != nil {
		log.Error("%s %v", "failed to get counters: ", err)
	}
	s.Counters.BlockedRequests = blockedCount
	s.Counters.AllowedRequests = allowedCount
}
