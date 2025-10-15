package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"goaway/backend/alert"
	"goaway/backend/audit"
	"goaway/backend/dns/database"
	"goaway/backend/dns/lists"
	model "goaway/backend/dns/server/models"
	"goaway/backend/logging"
	"goaway/backend/mac"
	notification "goaway/backend/notifications"
	"goaway/backend/resolution"
	"goaway/backend/settings"
	"net"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/miekg/dns"
)

var (
	log = logging.GetLogger()
)

type DNSServer struct {
	Whitelist           *lists.Whitelist
	DBManager           *database.Manager
	Blacklist           *lists.Blacklist
	dnsClient           *dns.Client
	Notifications       *notification.Manager
	Config              *settings.Config
	WebServer           *gin.Engine
	logEntryChannel     chan model.RequestLogEntry
	WSQueries           *websocket.Conn
	WSCommunication     *websocket.Conn
	hostnameCache       sync.Map
	clientCache         sync.Map
	Cache               sync.Map
	logIntervalSeconds  int
	WSCommunicationLock sync.Mutex

	AuditService      *audit.Service
	AlertService      *alert.Service
	MACService        *mac.Service
	ResolutionService *resolution.Service
}

type CachedRecord struct {
	ExpiresAt   time.Time
	CachedAt    time.Time
	Key         string
	Domain      string
	IPAddresses []dns.RR
	OriginalTTL uint32
}

type Request struct {
	Sent           time.Time
	ResponseWriter dns.ResponseWriter
	Msg            *dns.Msg
	Client         *model.Client
	Protocol       model.Protocol
	Question       dns.Question
	Prefetch       bool
}

type communicationMessage struct {
	IP       string `json:"ip"`
	Client   bool   `json:"client"`
	Upstream bool   `json:"upstream"`
	DNS      bool   `json:"dns"`
}

func NewDNSServer(config *settings.Config,
	dbManager *database.Manager,
	notificationsManager *notification.Manager,
	cert tls.Certificate,
) (*DNSServer, error) {
	whitelistEntry, err := lists.InitializeWhitelist(dbManager)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize whitelist: %w", err)
	}

	var client dns.Client
	if cert.Certificate != nil {
		client = dns.Client{Net: "tcp-tls"}
	}

	server := &DNSServer{
		Config:             config,
		Whitelist:          whitelistEntry,
		DBManager:          dbManager,
		logIntervalSeconds: 1,
		logEntryChannel:    make(chan model.RequestLogEntry, 1000),
		dnsClient:          &client,
		Notifications:      notificationsManager,
		Cache:              sync.Map{},
	}

	return server, nil
}

func (s *DNSServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) != 1 {
		log.Warning("Query contains more than one question, ignoring!")
		r.SetRcode(r, dns.RcodeFormatError)
		_ = w.WriteMsg(r)
		return
	}

	client := s.getClientInfo(w.RemoteAddr().String())
	protocol := s.detectProtocol(w)

	go s.WSCom(communicationMessage{
		Client:   true,
		Upstream: false,
		DNS:      false,
		IP:       client.IP,
	})

	entry := s.processQuery(&Request{
		ResponseWriter: w,
		Msg:            r,
		Question:       r.Question[0],
		Sent:           time.Now(),
		Client:         client,
		Prefetch:       false,
		Protocol:       protocol,
	})

	go s.WSCom(communicationMessage{
		Client:   false,
		Upstream: false,
		DNS:      true,
		IP:       client.IP,
	})
	s.logEntryChannel <- entry
}

func (s *DNSServer) detectProtocol(w dns.ResponseWriter) model.Protocol {
	if conn, ok := w.(interface{ ConnectionState() *tls.ConnectionState }); ok {
		if conn.ConnectionState() != nil {
			return model.DoT
		}
	}

	if conn, ok := w.(interface{ RemoteAddr() net.Addr }); ok {
		addr := conn.RemoteAddr()
		if addr.Network() == "tcp" {
			return model.TCP
		}
	}

	return model.UDP
}

func (s *DNSServer) PopulateHostnameCache() error {
	type Result struct {
		ClientIP   string
		ClientName string
	}

	var results []Result

	if err := s.DBManager.Conn.
		Model(&database.RequestLog{}).
		Select("DISTINCT client_ip, client_name").
		Where("client_name IS NOT NULL AND client_name != ?", "unknown").
		Find(&results).Error; err != nil {
		return fmt.Errorf("failed to fetch hostnames: %w", err)
	}

	for _, r := range results {
		if _, exists := s.hostnameCache.Load(r.ClientName); !exists {
			s.hostnameCache.Store(r.ClientName, r.ClientIP)
		}
	}

	return nil
}

func (s *DNSServer) WSCom(message communicationMessage) {
	if s.WSCommunication == nil {
		return
	}

	s.WSCommunicationLock.Lock()
	defer s.WSCommunicationLock.Unlock()

	if s.WSCommunication == nil {
		return
	}

	entryWSJson, err := json.Marshal(message)
	if err != nil {
		log.Error("Failed to marshal websocket message: %v", err)
		return
	}

	if err := s.WSCommunication.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Warning("Failed to set websocket write deadline: %v", err)
		return
	}

	if err := s.WSCommunication.WriteMessage(websocket.TextMessage, entryWSJson); err != nil {
		log.Debug("Failed to write websocket message: %v", err)
		s.WSCommunication = nil
	}
}
