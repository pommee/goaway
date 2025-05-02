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

type DNSServer struct {
	Config             settings.Config
	Blacklist          *blacklist.Blacklist
	DB                 *sql.DB
	logIntervalSeconds int
	lastLogTime        time.Time
	cache              sync.Map
	clientCache        sync.Map
	WebServer          *gin.Engine
	logEntryChannel    chan model.RequestLogEntry
	WS                 *websocket.Conn
	dnsClient          *dns.Client
	Status             settings.Status
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

type Request struct {
	w        dns.ResponseWriter
	msg      *dns.Msg
	question dns.Question
	sent     time.Time
	client   *model.Client
}

func NewDNSServer(config settings.Config) (*DNSServer, error) {
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
		Config:             config,
		Blacklist:          blacklistEntry,
		DB:                 db.Con,
		logIntervalSeconds: 1,
		lastLogTime:        time.Now(),
		cache:              sync.Map{},
		logEntryChannel:    make(chan model.RequestLogEntry, 1000),
		dnsClient:          dnsClient,
	}

	return server, nil
}

func (s *DNSServer) Init() (int, *dns.Server) {
	server := &dns.Server{
		Addr:      fmt.Sprintf(":%d", s.Config.DNSPort),
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
