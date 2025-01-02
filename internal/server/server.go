package server

import (
	"encoding/json"
	"fmt"
	"goaway/internal/blacklist"
	"goaway/internal/logger"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
)

var log = logger.GetLogger()

type ServerConfig struct {
	Port            int
	WebsitePort     int
	LogLevel        logger.LogLevel
	LoggingDisabled bool
	UpstreamDNS     []string
	BestUpstreamDNS string
	BlacklistPath   string
	CountersFile    string
	RequestLogFile  string
	CacheTTL        time.Duration
}

type DNSServer struct {
	Config             ServerConfig
	Blacklist          blacklist.Blacklist
	Counters           CounterDetails
	lastLogTime        time.Time
	logIntervalSeconds int
	cache              sync.Map
	cacheMutex         sync.RWMutex
	RequestLog         []RequestLogEntry
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
	if !fileExists(config.CountersFile) {
		newCounters := &CounterDetails{}
		saveCounters(config.CountersFile, *newCounters)
	}
	counters, _ := LoadCounters(config.CountersFile)

	if !fileExists(config.RequestLogFile) {
		err := os.WriteFile(config.RequestLogFile, []byte("[]"), 0644)
		if err != nil {
			log.Error("Error writing file %s", err)
		}
	}

	requestLog, err := LoadRequestLog(config.RequestLogFile)
	if err != nil {
		return DNSServer{}, fmt.Errorf("failed to load request log: %w", err)
	}

	start := time.Now()
	dnsBlacklist, _ := blacklist.LoadBlacklist(config.BlacklistPath)
	log.Debug("Loading %s took %v", config.BlacklistPath, time.Since(start))
	return DNSServer{
		Config:             *config,
		Blacklist:          dnsBlacklist,
		Counters:           counters,
		lastLogTime:        time.Now(),
		logIntervalSeconds: 1,
		cache:              sync.Map{},
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
	timestamp := time.Now()
	clientIP := strings.Split(w.RemoteAddr().String(), ":")[0]

	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	var wg sync.WaitGroup
	results := make(chan RequestLogEntry, len(r.Question))

	for _, question := range r.Question {
		wg.Add(1)
		go func(question dns.Question) {
			defer wg.Done()

			if s.IsBlacklisted(question.Name) {
				s.handleBlacklisted(w, msg, question.Name)
				results <- RequestLogEntry{
					Timestamp:      timestamp,
					Domain:         question.Name,
					Blocked:        true,
					Cached:         false,
					ResponseTimeNS: time.Since(timestamp),
					ClientInfo:     &Client{IP: clientIP},
				}
			} else {
				entry := s.handleQuery(w, msg, question, timestamp, question.Name, clientIP)
				results <- entry
			}
		}(question)
	}

	wg.Wait()
	close(results)

	var requestLogMutex sync.Mutex
	for entry := range results {
		requestLogMutex.Lock()
		s.RequestLog = append(s.RequestLog, entry)
		requestLogMutex.Unlock()
	}

	go s.SaveRequestLog(s.Config.RequestLogFile)
	go s.logStats()
}
func (s *DNSServer) handleQuery(w dns.ResponseWriter, msg *dns.Msg, question dns.Question, timestamp time.Time, domain, clientIP string) RequestLogEntry {
	answers, cached := s.resolve(question.Name, question.Qtype)
	msg.Answer = append(msg.Answer, answers...)
	_ = w.WriteMsg(msg)

	responseTime := time.Since(timestamp)
	s.Counters.AllowedRequests++

	return RequestLogEntry{
		Timestamp:      timestamp,
		Domain:         domain,
		Blocked:        false,
		Cached:         cached,
		ResponseTimeNS: responseTime,
		ClientInfo:     &Client{IP: clientIP},
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
		in, _, err := c.Exchange(m, s.Config.BestUpstreamDNS)
		if err != nil {
			log.Error("Resolution error: %v", err)
			return
		}
		ipAddresses = in.Answer
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
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

func (s *DNSServer) IsBlacklisted(domain string) bool {
	return s.Blacklist.Domains[strings.TrimSuffix(domain, ".")]
}

func (s *DNSServer) handleBlacklisted(w dns.ResponseWriter, msg *dns.Msg, domain string) {
	domain = strings.TrimSpace(domain)
	log.Info("Blocked: %s", domain)

	msg.Rcode = dns.RcodeSuccess
	rr := &dns.A{
		Hdr: dns.RR_Header{
			Name:   domain,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    uint32(s.Config.CacheTTL),
		},
		A: net.ParseIP("0.0.0.0"),
	}

	msg.Answer = append(msg.Answer, rr)
	if err := w.WriteMsg(msg); err != nil {
		log.Error("Error sending DNS response: %v", err)
	}

	s.Counters.BlockedRequests++
}

func (s *DNSServer) logStats() {
	if time.Since(s.lastLogTime).Seconds() > float64(s.logIntervalSeconds) {
		s.lastLogTime = time.Now()

		err := saveCounters(s.Config.CountersFile, s.Counters)
		if err != nil {
			log.Error("Failed to save counters: %v", err)
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
