package server

import (
	"encoding/json"
	"fmt"
	notification "goaway/backend"
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
	log = logging.GetLogger()
)

type MacVendor struct {
	Vendor string `json:"company"`
}

type DNSServer struct {
	Config              settings.Config
	Blacklist           *blacklist.Blacklist
	DBManager           *database.DatabaseManager
	logIntervalSeconds  int
	lastLogTime         time.Time
	Cache               sync.Map
	clientCache         sync.Map
	WebServer           *gin.Engine
	logEntryChannel     chan model.RequestLogEntry
	WSQueries           *websocket.Conn
	WSCommunication     *websocket.Conn
	WSCommunicationLock sync.Mutex
	dnsClient           *dns.Client
	Status              settings.Status
	Notifications       *notification.Manager
}

type QueryResponse struct {
	IPAddresses []dns.RR
	Ttl         uint32
	Status      int
}

type CachedRecord struct {
	IPAddresses []dns.RR
	ExpiresAt   time.Time
	CachedAt    time.Time
	OriginalTTL uint32
	Key         string
	Domain      string
}

type Request struct {
	W        dns.ResponseWriter
	Msg      *dns.Msg
	Question dns.Question
	Sent     time.Time
	Client   *model.Client
	Prefetch bool
}

type communicationMessage struct {
	Client   bool   `json:"client"`
	Upstream bool   `json:"upstream"`
	DNS      bool   `json:"dns"`
	Ip       string `json:"ip"`
}

func NewDNSServer(config settings.Config, dbManager *database.DatabaseManager, notificationsManager *notification.Manager) (*DNSServer, error) {
	blacklistEntry, err := blacklist.Initialize(dbManager)
	if err != nil {
		log.Error("Failed to initialize blacklist")
	}

	server := &DNSServer{
		Config:             config,
		Blacklist:          blacklistEntry,
		DBManager:          dbManager,
		logIntervalSeconds: 1,
		lastLogTime:        time.Now(),
		logEntryChannel:    make(chan model.RequestLogEntry, 1000),
		dnsClient:          new(dns.Client),
		Notifications:      notificationsManager,
	}

	return server, nil
}

func (s *DNSServer) Init() (int, *dns.Server) {
	server := &dns.Server{
		Addr:      fmt.Sprintf(":%d", s.Config.DNS.Port),
		Net:       "udp",
		Handler:   s,
		ReusePort: true,
	}

	domains, _ := s.Blacklist.CountDomains()
	return domains, server
}

func (s *DNSServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) != 1 {
		log.Warning("Query container more than one question, ignoring!")
		r.SetRcode(r, dns.RcodeFormatError)
		_ = w.WriteMsg(r)
		return
	}

	sent := time.Now()
	client := s.getClientInfo(w.RemoteAddr().String())
	go s.WSCom(communicationMessage{true, false, false, client.IP})

	entry := s.processQuery(&Request{w, r, r.Question[0], sent, client, false})

	go s.WSCom(communicationMessage{false, false, true, client.IP})
	s.logEntryChannel <- entry
}

func (s *DNSServer) WSCom(message communicationMessage) {
	if s.WSCommunication != nil {
		entryWSJson, _ := json.Marshal(message)
		s.WSCommunicationLock.Lock()
		_ = s.WSCommunication.WriteMessage(websocket.TextMessage, entryWSJson)
		s.WSCommunicationLock.Unlock()
	}
}
