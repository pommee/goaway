package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"goaway/internal/blacklist"
	"goaway/internal/database"
	"goaway/internal/logging"
	model "goaway/internal/server/models"
	"goaway/internal/settings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/miekg/dns"
)

var (
	log     = logging.GetLogger()
	dbMutex sync.Mutex
	wsMutex sync.Mutex
)

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
	logEntryChannel     chan model.RequestLogEntry
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

type Request struct {
	w         dns.ResponseWriter
	msg       *dns.Msg
	question  dns.Question
	timestamp time.Time
	client    *model.Client
}

func NewDNSServer(config *settings.DNSServerConfig) (*DNSServer, error) {
	db, err := database.Initialize()
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	blacklistEntry := blacklist.Blacklist{
		DB: db.Con,
		BlocklistURL: map[string]string{
			"StevenBlack": "https://raw.githubusercontent.com/StevenBlack/hosts/refs/heads/master/hosts",
		},
	}

	if err := blacklistEntry.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize blacklist: %w", err)
	}

	if err := blacklistEntry.InitializeBlocklist("Custom", ""); err != nil {
		return nil, fmt.Errorf("failed to initialize custom blocklist: %w", err)
	}

	blacklistEntry.BlocklistURL, err = blacklistEntry.GetBlocklistUrls()
	if err != nil {
		log.Error("Failed to get blocklist URLs: %v", err)
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
	results := make(chan model.RequestLogEntry, len(r.Question))

	for _, question := range r.Question {
		wg.Add(1)
		go func(question dns.Question) {
			defer wg.Done()
			client := &model.Client{IP: clientIP, Name: clientName, MAC: macAddress}
			entry := s.processQuery(&Request{w, msg, question, timestamp, client})
			entryWSJson, _ := json.Marshal(entry)
			if s.WS != nil {
				wsMutex.Lock()
				_ = s.WS.WriteMessage(websocket.TextMessage, entryWSJson)
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
