package server

import (
	"encoding/json"
	"fmt"
	"goaway/internal/blacklist"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type ServerConfig struct {
	Port           int
	WebsitePort    int
	UpstreamDNS    string
	BlacklistPath  string
	CountersFile   string
	RequestLogFile string
	CacheTTL       time.Duration
}

type DNSServer struct {
	Config             ServerConfig
	Blacklist          blacklist.Blacklist
	Counters           CounterDetails
	lastLogTime        time.Time
	logIntervalSeconds int
	cache              map[string]cachedRecord
	cacheMutex         sync.Mutex
	RequestLog         []RequestLogEntry
}

type CounterDetails struct {
	AllowedRequests int           `json:"allowed_requests"`
	BlockedRequests int           `json:"blocked_requests"`
	Details         []DomainStats `json:"details"`
}

type DomainStats struct {
	Blocked int `json:"blocked"`
}

type cachedRecord struct {
	IPAddresses []dns.RR
	ExpiresAt   time.Time
}

type RequestLogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	Domain     string    `json:"domain"`
	Blocked    bool      `json:"blocked"`
	ClientInfo *Client   `json:"client"`
}

type Client struct {
	IP   string
	Name string
}

func NewDNSServer(config ServerConfig) (DNSServer, error) {
	if !fileExists(config.CountersFile) {
		newCounters := &CounterDetails{Details: []DomainStats{}}
		saveCounters(config.CountersFile, *newCounters)
	}
	counters, _ := LoadCounters(config.CountersFile)

	if !fileExists(config.RequestLogFile) {
		err := os.WriteFile(config.RequestLogFile, []byte("[]"), 0644)
		if err != nil {
			log.Printf("Error writing file: %v\n", err)
		}
	}

	requestLog, err := LoadRequestLog(config.RequestLogFile)
	if err != nil {
		return DNSServer{}, fmt.Errorf("failed to load request log: %w", err)
	}

	dnsBlacklist, _ := blacklist.LoadBlacklist(config.BlacklistPath)
	return DNSServer{
		Config:             config,
		Blacklist:          dnsBlacklist,
		Counters:           counters,
		lastLogTime:        time.Now(),
		logIntervalSeconds: 1,
		cache:              make(map[string]cachedRecord),
		RequestLog:         requestLog,
	}, nil
}

func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (s *DNSServer) Init() (int, *dns.Server) {
	server := &dns.Server{
		Addr:      fmt.Sprintf(":%d", s.Config.Port),
		Net:       "udp",
		Handler:   s,
		UDPSize:   65535,
		ReusePort: true,
	}

	return len(s.Blacklist.Domains), server
}

func (s *DNSServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	clientIP := strings.Split(w.RemoteAddr().String(), ":")[0]
	var clientName = "None"

	lookupNames, _ := net.LookupAddr(clientIP)
	if len(lookupNames) > 0 {
		clientName = strings.TrimSuffix(lookupNames[0], ".")
	}

	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, question := range r.Question {
		timestamp := time.Now()
		domain := strings.TrimSuffix(question.Name, ".")

		// Check if the domain is blacklisted
		if s.IsBlacklisted(question.Name) {
			s.handleBlacklisted(w, msg, question.Name)
			s.RequestLog = append(s.RequestLog, RequestLogEntry{
				Timestamp:  timestamp,
				Domain:     domain,
				Blocked:    true,
				ClientInfo: &Client{IP: clientIP, Name: clientName},
			})
			go s.SaveRequestLog(s.Config.RequestLogFile)
			return
		}

		s.handleQuery(w, msg, question)
		s.RequestLog = append(s.RequestLog, RequestLogEntry{
			Timestamp:  timestamp,
			Domain:     domain,
			Blocked:    false,
			ClientInfo: &Client{IP: clientIP, Name: clientName},
		})
		go s.SaveRequestLog(s.Config.RequestLogFile)
	}

	s.logStats()
}

func (s *DNSServer) IsBlacklisted(domain string) bool {
	domain = strings.TrimSuffix(domain, ".")
	return s.Blacklist.Domains[domain]
}

func (s *DNSServer) handleBlacklisted(w dns.ResponseWriter, msg *dns.Msg, domain string) {
	domain = strings.TrimSuffix(domain, ".")
	log.Printf("Blocked query, domain: %s\n", domain)
	msg.Rcode = dns.RcodeNameError // NXDOMAIN = blacklisted domain
	w.WriteMsg(msg)

	s.Counters.BlockedRequests++
}

func (s *DNSServer) handleQuery(w dns.ResponseWriter, msg *dns.Msg, question dns.Question) {
	answers := s.resolve(question.Name, question.Qtype)
	msg.Answer = append(msg.Answer, answers...)
	w.WriteMsg(msg)

	s.Counters.AllowedRequests++
}

func (s *DNSServer) resolve(domain string, qtype uint16) []dns.RR {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	if cached, found := s.cache[domain]; found && time.Now().Before(cached.ExpiresAt) {
		return cached.IPAddresses
	}

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true

	c := new(dns.Client)
	in, _, err := c.Exchange(m, s.Config.UpstreamDNS)
	if err != nil {
		log.Printf("Resolution error: %v\n", err)
		return nil
	}

	s.cache[domain] = cachedRecord{
		IPAddresses: in.Answer,
		ExpiresAt:   time.Now().Add(s.Config.CacheTTL),
	}

	return in.Answer
}

func (s *DNSServer) logStats() {
	if time.Since(s.lastLogTime).Seconds() > float64(s.logIntervalSeconds) {
		s.lastLogTime = time.Now()

		err := saveCounters(s.Config.CountersFile, s.Counters)
		if err != nil {
			log.Printf("Failed to save counters: %v\n", err)
		}
	}
}

func LoadCounters(filename string) (CounterDetails, error) {
	counters := CounterDetails{}

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return counters, nil
		}
		return counters, err
	}

	err = json.Unmarshal(data, &counters)
	if err != nil {
		return counters, err
	}

	return counters, nil
}

func saveCounters(filename string, counters CounterDetails) error {
	data, err := json.Marshal(counters)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func LoadRequestLog(filename string) ([]RequestLogEntry, error) {
	var requestLog []RequestLogEntry

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return requestLog, nil
		}
		return requestLog, err
	}

	err = json.Unmarshal(data, &requestLog)
	if err != nil {
		return requestLog, err
	}

	return requestLog, nil
}

func (s *DNSServer) SaveRequestLog(filename string) error {
	data, err := json.Marshal(s.RequestLog)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
