package server

import (
	"database/sql"
	"fmt"
	"goaway/internal/blacklist"
	"goaway/internal/database"
	"goaway/internal/logging"
	"net"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
)

var log = logging.GetLogger()

const batchSize = 100

type ServerConfig struct {
	Port              int
	WebsitePort       int
	LogLevel          logging.LogLevel
	LoggingDisabled   bool
	UpstreamDNS       []string
	PreferredUpstream string
	CacheTTL          time.Duration
}

type DNSServer struct {
	Config             ServerConfig
	Blacklist          blacklist.Blacklist
	DB                 *sql.DB
	Counters           CounterDetails
	lastLogTime        time.Time
	logIntervalSeconds int
	cache              sync.Map
	WebServer          *gin.Engine
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
	Blocked        bool          `json:"blocked"`
	Cached         bool          `json:"cached"`
	ResponseTimeNS time.Duration `json:"responseTimeNS"`
	ClientInfo     *Client       `json:"client"`
}

type Client struct {
	IP   string
	Name string
}

func NewDNSServer(config *ServerConfig) (DNSServer, error) {
	db, err := database.Initialize()

	if err != nil {
		return DNSServer{}, fmt.Errorf("failed to create tables: %w", err)
	}

	blacklist := blacklist.Blacklist{
		DB:           db.Con,
		BlocklistURL: []string{"https://raw.githubusercontent.com/StevenBlack/hosts/refs/heads/master/hosts"},
	}
	if err := blacklist.Initialize(); err != nil {
		return DNSServer{}, fmt.Errorf("failed to initialize blacklist: %w", err)
	}

	return DNSServer{
		Config:             *config,
		Blacklist:          blacklist,
		DB:                 db.Con,
		lastLogTime:        time.Now(),
		logIntervalSeconds: 1,
		cache:              sync.Map{},
	}, nil
}

func (s *DNSServer) Init() (int, *dns.Server) {
	server := &dns.Server{
		Addr:      fmt.Sprintf(":%d", s.Config.Port),
		Net:       "udp",
		Handler:   s,
		UDPSize:   65535,
		ReusePort: true,
	}

	if err := s.LoadCounters(); err != nil {
		log.Error("Failed to load counters from database: %v", err)
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

	var requestLogMutex sync.Mutex
	var batch []RequestLogEntry
	for entry := range results {
		requestLogMutex.Lock()
		batch = append(batch, entry)
		if len(batch) >= batchSize {
			go s.SaveRequestLogs(batch)
			batch = nil
		}
		requestLogMutex.Unlock()
	}

	if len(batch) > 0 {
		s.SaveRequestLogs(batch)
	}
}

func (s *DNSServer) getClientInfo(remoteAddr string) (string, string) {
	clientIP := strings.Split(remoteAddr, ":")[0]
	lookupNames, _ := net.LookupAddr(clientIP)
	if len(lookupNames) > 0 {
		return clientIP, strings.TrimSuffix(lookupNames[0], ".")
	}
	return clientIP, ""
}

func (s *DNSServer) processQuery(w dns.ResponseWriter, msg *dns.Msg, question dns.Question, timestamp time.Time, clientIP, clientName string) RequestLogEntry {
	isBlacklisted, _ := s.Blacklist.IsBlacklisted(strings.TrimSuffix(question.Name, "."))
	if isBlacklisted {
		return s.handleBlacklisted(w, msg, question.Name, timestamp, clientIP, clientName)
	}

	return s.handleQuery(w, msg, question, timestamp, clientIP, clientName)
}

func (s *DNSServer) handleQuery(w dns.ResponseWriter, msg *dns.Msg, question dns.Question, timestamp time.Time, clientIP, clientName string) RequestLogEntry {
	answers, cached := s.resolve(question.Name, question.Qtype)
	msg.Answer = append(msg.Answer, answers...)
	_ = w.WriteMsg(msg)

	s.Counters.AllowedRequests++

	return RequestLogEntry{
		Timestamp:      timestamp,
		Domain:         question.Name,
		Blocked:        false,
		Cached:         cached,
		ResponseTimeNS: time.Since(timestamp),
		ClientInfo:     &Client{IP: clientIP, Name: clientName},
	}
}

func (s *DNSServer) resolve(domain string, qtype uint16) ([]dns.RR, bool) {
	if cached, found := s.cache.Load(domain); found {
		cachedRecord := cached.(cachedRecord)
		if time.Now().Before(cachedRecord.ExpiresAt) {
			log.Debug("Cached response for %s", domain)
			return cachedRecord.IPAddresses, true
		}
	}

	var ipAddresses []dns.RR
	done := make(chan struct{})
	go func() {
		defer close(done)
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(domain), qtype)
		m.RecursionDesired = true

		c := new(dns.Client)
		in, _, err := c.Exchange(m, s.Config.PreferredUpstream)
		if err != nil {
			log.Error("Resolution error: %v", err)
			return
		}
		ipAddresses = in.Answer
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		log.Warning("DNS lookup for %s timed out", domain)
		return nil, false
	}

	if len(ipAddresses) > 0 {
		s.cache.Store(domain, cachedRecord{
			IPAddresses: ipAddresses,
			ExpiresAt:   time.Now().Add(s.Config.CacheTTL),
		})
	}

	return ipAddresses, false
}

func (s *DNSServer) handleBlacklisted(w dns.ResponseWriter, msg *dns.Msg, domain string, timestamp time.Time, clientIP, clientName string) RequestLogEntry {
	log.Info("Blocked: %s", domain)

	msg.Rcode = dns.RcodeSuccess
	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   domain,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    uint32(s.Config.CacheTTL.Seconds()),
		},
		A: net.ParseIP("0.0.0.0"),
	}
	msg.Answer = append(msg.Answer, rr)
	_ = w.WriteMsg(msg)

	s.Counters.BlockedRequests++

	return RequestLogEntry{
		Timestamp:      timestamp,
		Domain:         domain,
		Blocked:        true,
		Cached:         false,
		ResponseTimeNS: time.Since(timestamp),
		ClientInfo:     &Client{IP: clientIP, Name: clientName},
	}
}

func (s *DNSServer) LoadCounters() error {
	row := s.DB.QueryRow("SELECT allowed_requests, blocked_requests FROM counters WHERE id = 1")
	counters := CounterDetails{}
	if err := row.Scan(&counters.AllowedRequests, &counters.BlockedRequests); err != nil {
		if err == sql.ErrNoRows {
			_, err = s.DB.Exec("INSERT INTO counters (id, allowed_requests, blocked_requests) VALUES (1, 0, 0)")
			return err
		}
		return err
	}

	s.Counters = counters
	return nil
}

func (s *DNSServer) SaveCounters(counters CounterDetails) error {
	_, err := s.DB.Exec("UPDATE counters SET allowed_requests = ?, blocked_requests = ? WHERE id = 1",
		counters.AllowedRequests, counters.BlockedRequests)
	return err
}

func (s *DNSServer) LoadRequestLog() ([]RequestLogEntry, error) {
	rows, err := s.DB.Query("SELECT timestamp, domain, blocked, cached, response_time_ns, client_ip, client_name FROM request_log")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []RequestLogEntry
	for rows.Next() {
		var entry RequestLogEntry
		var clientIP, clientName sql.NullString
		if err := rows.Scan(&entry.Timestamp, &entry.Domain, &entry.Blocked, &entry.Cached, &entry.ResponseTimeNS, &clientIP, &clientName); err != nil {
			return nil, err
		}
		entry.ClientInfo = &Client{IP: clientIP.String, Name: clientName.String}
		logs = append(logs, entry)
	}
	return logs, nil
}

func (s *DNSServer) SaveRequestLogs(entries []RequestLogEntry) {
	tx, err := s.DB.Begin()
	if err != nil {
		log.Error("Could not start database transaction %v", err)
	}
	defer func() {
		if err := tx.Commit(); err != nil {
			log.Warning("DB commit error %v", err)
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO request_log (timestamp, domain, blocked, cached, response_time_ns, client_ip, client_name) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Error("Could not create a prepared statement for request logs %v", err)
	}
	defer stmt.Close()

	for _, entry := range entries {
		if _, err := stmt.Exec(
			entry.Timestamp, entry.Domain, entry.Blocked, entry.Cached, entry.ResponseTimeNS, entry.ClientInfo.IP, entry.ClientInfo.Name,
		); err != nil {
			log.Error("Could not save request log %v", err)
		}
	}

	s.saveCounters(tx)
}

func (s *DNSServer) saveCounters(tx *sql.Tx) {
	stmt, err := tx.Prepare(`
        UPDATE counters
        SET allowed_requests = ?, blocked_requests = ?
    `)
	if err != nil {
		log.Error("Could not create a prepared statement when saving counters %v", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(s.Counters.AllowedRequests, s.Counters.BlockedRequests)
	if err != nil {
		log.Error("Could not create a prepared statement when saving counters %v", err)
	}

	if rowsAffected, err := res.RowsAffected(); err != nil {
		log.Warning("Could not create counters %v", err)
	} else if rowsAffected == 0 {
		stmt, err = tx.Prepare(`
            INSERT INTO counters (allowed_requests, blocked_requests)
            VALUES (?, ?)
        `)
		if err != nil {
			log.Warning("Could not create a prepared statement for counters %v", err)
		}
		defer stmt.Close()

		_, err = stmt.Exec(s.Counters.AllowedRequests, s.Counters.BlockedRequests)
		if err != nil {
			log.Warning("Could not update counters %v", err)
		}
	}
}
