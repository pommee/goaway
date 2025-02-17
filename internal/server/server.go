package server

import (
	"database/sql"
	"fmt"
	"goaway/internal/blacklist"
	"goaway/internal/database"
	"goaway/internal/logging"
	"goaway/internal/settings"
	"net"
	"os"
	"strings"
	"sync"
	"time"

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

const batchSize = 100

var dbMutex sync.Mutex

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

type Client struct {
	IP   string
	Name string
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

	if err := blacklist.InitializeCustomBlocklist(); err != nil {
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
	clientIP, clientName := s.getClientInfo(w.RemoteAddr().String())

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
			entry := s.processQuery(w, msg, question, timestamp, clientIP, clientName)
			results <- entry
		}(question)
	}

	wg.Wait()
	close(results)

	for entry := range results {
		s.logEntryChannel <- entry
	}
}

func (s *DNSServer) getClientInfo(remoteAddr string) (string, string) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Error("%v", err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	clientIP := strings.Split(remoteAddr, ":")[0]

	if clientIP == "127.0.0.1" || clientIP == "::1" || clientIP == localAddr.IP.String() {
		hostname, err := os.Hostname()
		if err != nil {
			return clientIP, "localhost"
		}
		return clientIP, hostname
	}

	hostnames, err := net.LookupAddr(clientIP)
	if err != nil || len(hostnames) == 0 {
		return clientIP, "unknown"
	}
	return clientIP, strings.TrimRight(hostnames[0], ".")
}

func (s *DNSServer) processQuery(w dns.ResponseWriter, msg *dns.Msg, question dns.Question, timestamp time.Time, clientIP, clientName string) RequestLogEntry {
	isBlacklisted, err := s.Blacklist.IsBlacklisted(strings.TrimSuffix(question.Name, "."))
	if err != nil {
		log.Error("%v", err)
	}
	if isBlacklisted {
		return s.handleBlacklisted(w, msg, question.Name, timestamp, clientIP, clientName)
	}
	return s.handleQuery(w, msg, question, timestamp, clientIP, clientName)
}

func (s *DNSServer) handleQuery(w dns.ResponseWriter, msg *dns.Msg, question dns.Question, timestamp time.Time, clientIP, clientName string) RequestLogEntry {
	answers, cached, status := s.resolve(question.Name, question.Qtype)
	msg.Answer = append(msg.Answer, answers...)

	if len(answers) == 0 && question.Qtype == dns.TypePTR {
		authority := s.createSOARecord(question.Name)
		msg.Ns = append(msg.Ns, authority)
	}

	_ = w.WriteMsg(msg)
	s.Counters.AllowedRequests++

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

	return RequestLogEntry{
		Timestamp:      timestamp,
		Domain:         question.Name,
		IP:             resolvedAddresses,
		Blocked:        false,
		Cached:         cached,
		ResponseTimeNS: time.Since(timestamp),
		ClientInfo:     &Client{IP: clientIP, Name: clientName},
		Status:         status,
		QueryType:      dns.TypeToString[question.Qtype],
	}
}

func (s *DNSServer) createSOARecord(name string) dns.RR {
	return &dns.SOA{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    1800,
		},
		Ns:      "ns1.goaway.",
		Mbox:    "admin.goaway.",
		Serial:  2024021500,
		Refresh: 10000,
		Retry:   2400,
		Expire:  604800,
		Minttl:  1800,
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

func (s *DNSServer) handleBlacklisted(w dns.ResponseWriter, msg *dns.Msg, domain string, timestamp time.Time, clientIP, clientName string) RequestLogEntry {
	log.Info("Blocked: %s", domain)

	msg.Rcode = dns.RcodeSuccess
	var status = "Blacklisted"

	rr4 := &dns.A{
		Hdr: dns.RR_Header{
			Name:   domain,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    uint32(s.Config.CacheTTL.Seconds()),
		},
		A: net.ParseIP("0.0.0.0"),
	}

	rr6 := &dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   domain,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    uint32(s.Config.CacheTTL.Seconds()),
		},
		AAAA: net.ParseIP("::"),
	}

	msg.Answer = append(msg.Answer, rr4, rr6)
	_ = w.WriteMsg(msg)

	s.Counters.BlockedRequests++

	return RequestLogEntry{
		Timestamp:      timestamp,
		Domain:         domain,
		IP:             []string{""},
		Blocked:        true,
		Cached:         false,
		ResponseTimeNS: time.Since(timestamp),
		ClientInfo:     &Client{IP: clientIP, Name: clientName},
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
		var clientIP, clientName sql.NullString
		if err := rows.Scan(&entry.Timestamp, &entry.Domain, &entry.IP, &entry.Blocked, &entry.Cached, &entry.ResponseTimeNS, &clientIP, &clientName); err != nil {
			return nil, err
		}
		entry.ClientInfo = &Client{IP: clientIP.String, Name: clientName.String}
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
