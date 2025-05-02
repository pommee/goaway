package server

import (
	"database/sql"
	"fmt"
	"goaway/backend/dns/blacklist"
	"goaway/backend/dns/database"
	model "goaway/backend/dns/server/models"
	"goaway/backend/logging"
	"goaway/backend/settings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/miekg/dns"
)

var (
	log     = logging.GetLogger()
	dbMutex sync.Mutex
)

type MacVendor struct {
	Vendor string `json:"company"`
}

type Status struct {
	Paused    bool
	PausedAt  time.Time
	PauseTime int
}

type DNSServer struct {
	Config              settings.DNSServerConfig
	Blacklist           *blacklist.Blacklist
	DB                  *sql.DB
	Counters            CounterDetails
	StatisticsRetention int
	logIntervalSeconds  int
	lastLogTime         time.Time
	cache               sync.Map
	WebServer           *gin.Engine
	logEntryChannel     chan model.RequestLogEntry
	WS                  *websocket.Conn
	dnsClient           *dns.Client
	Status              Status
}

type QueryResponse struct {
	IPAddresses []dns.RR
	Ttl         uint32
	Status      int
}

type cachedRecord struct {
	IPAddresses []dns.RR
	ExpiresAt   time.Time
	CachedAt    time.Time
	OriginalTTL uint32
	Key         string
}

type CounterDetails struct {
	AllowedRequests int `json:"allowed_requests"`
	BlockedRequests int `json:"blocked_requests"`
}

type Request struct {
	w        dns.ResponseWriter
	msg      *dns.Msg
	question dns.Question
	sent     time.Time
	client   *model.Client
}

func NewDNSServer(config *settings.DNSServerConfig) (*DNSServer, error) {
	db, err := database.Initialize()
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	blacklistEntry, err := blacklist.Initialize(db.Con)
	if err != nil {
		log.Error("Failed to initialize blacklist")
	}

	dnsClient := &dns.Client{
		Timeout: 3 * time.Second,
		UDPSize: 4096,
	}

	server := &DNSServer{
		Config:              *config,
		Blacklist:           blacklistEntry,
		DB:                  db.Con,
		StatisticsRetention: config.StatisticsRetention,
		lastLogTime:         time.Now(),
		logIntervalSeconds:  1,
		cache:               sync.Map{},
		logEntryChannel:     make(chan model.RequestLogEntry, 1000),
		dnsClient:           dnsClient,
	}

	return server, nil
}

func (s *DNSServer) Init() (int, *dns.Server) {
	server := &dns.Server{
		Addr:      fmt.Sprintf(":%d", s.Config.Port),
		Net:       "udp",
		Handler:   s,
		UDPSize:   512,
		ReusePort: true,
	}

	domains, _ := s.Blacklist.CountDomains()
	return domains, server
}

func (s *DNSServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	sent := time.Now()
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.RecursionAvailable = true

	if r.IsEdns0() != nil {
		msg.SetEdns0(1024, false)
	}

	client := s.getClientInfo(w.RemoteAddr().String())

	for _, question := range r.Question {
		entry := s.processQuery(&Request{w, msg, question, sent, client})
		s.logEntryChannel <- entry
	}
}
