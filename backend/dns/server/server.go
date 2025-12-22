package server

import (
	"crypto/tls"
	"encoding/json"
	"goaway/backend/alert"
	"goaway/backend/audit"
	"goaway/backend/blacklist"
	model "goaway/backend/dns/server/models"
	"goaway/backend/logging"
	"goaway/backend/mac"
	"goaway/backend/notification"
	"goaway/backend/request"
	"goaway/backend/resolution"
	"goaway/backend/settings"
	"goaway/backend/user"
	"goaway/backend/whitelist"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/miekg/dns"
	"gorm.io/gorm"
)

var (
	log = logging.GetLogger()
)

// DNSServer encapsulates the DNS handling logic and the runtime state used by
// the various DNS transports (UDP/TCP), secure transports (DoT) and HTTP-based frontends (DoH).
type DNSServer struct {
	// Database connection used by services for persistence
	dbConn *gorm.DB

	// Client used when querying upstream servers
	dnsClient *dns.Client

	// Application level settings, mostly used for DNS behaviour
	Config *settings.Config

	// Central channel where processed request log entries are pushed
	logEntryChannel chan model.RequestLogEntry

	// Websocket connection used to stream query logs to the web UI
	WSQueries *websocket.Conn

	// Websocket connection used to stream communication events to the UI
	// Used to visualize client/upstream/DNS activity
	WSCommunication *websocket.Conn

	// Guards writes to WSCommunication.
	WSCommunicationLock sync.Mutex

	// Cache mapping hostnames to client metadata to avoid repeated lookups when resolving PTR/hostnames
	clientHostnameCache sync.Map

	// Cache mapping IP -> client info (name, mac) for quick lookup during request processing
	clientIPCache sync.Map

	// In-memory cache for resolved DNS records to speed up responses and reduce upstream queries
	DomainCache sync.Map

	// DNSServer delegates database-backed lookups and persistence to these services,
	// rather than performing raw DB operations itself.
	RequestService      *request.Service
	AuditService        *audit.Service
	UserService         *user.Service
	AlertService        *alert.Service
	MACService          *mac.Service
	ResolutionService   *resolution.Service
	NotificationService *notification.Service
	BlacklistService    *blacklist.Service
	WhitelistService    *whitelist.Service
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

func NewDNSServer(config *settings.Config, dbconn *gorm.DB, cert tls.Certificate) (*DNSServer, error) {
	var client dns.Client
	if cert.Certificate != nil {
		client = dns.Client{Net: "tcp-tls"}
	}

	server := &DNSServer{
		Config:          config,
		dbConn:          dbconn,
		logEntryChannel: make(chan model.RequestLogEntry, 1000),
		dnsClient:       &client,
		DomainCache:     sync.Map{},
	}

	return server, nil
}

func (s *DNSServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if !s.validQuery(w, r) {
		return
	}

	var clientIP net.IP

	switch addr := w.RemoteAddr().(type) {
	case *net.UDPAddr:
		clientIP = addr.IP
	case *net.TCPAddr:
		clientIP = addr.IP
	default:
		return
	}

	client := s.getClientInfo(clientIP)
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

func (s *DNSServer) PopulateClientCaches() error {
	clients, err := s.RequestService.FetchAllClients()

	if err != nil {
		log.Warning("Could not populate client caches, reason: %v", err)
		return err
	}

	for _, client := range clients {
		s.clientHostnameCache.Store(client.Name, client)
		s.clientIPCache.Store(client.IP, client)
	}

	log.Debug("Populated client caches with %d client(s)", len(clients))
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

func (s *DNSServer) validQuery(w dns.ResponseWriter, r *dns.Msg) bool {
	failedCallback := func() bool {
		r.SetRcode(r, dns.RcodeFormatError)
		_ = w.WriteMsg(r)
		return false
	}

	if len(r.Question) != 1 {
		log.Warning("Query contains more than one question, ignoring!")
		return failedCallback()
	}

	if len(r.Question[0].Name) <= 1 {
		log.Warning("Query contains invalid question name '%s', ignoring!", r.Question[0].Name)
		return failedCallback()
	}

	return true
}
