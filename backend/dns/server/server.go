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
	notification "goaway/backend/notifications"
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
	Config              *settings.Config
	Blacklist           *lists.Blacklist
	Whitelist           *lists.Whitelist
	DBManager           *database.DatabaseManager
	logIntervalSeconds  int
	lastLogTime         time.Time
	Cache               sync.Map
	clientCache         sync.Map
	hostnameCache       sync.Map
	WebServer           *gin.Engine
	logEntryChannel     chan model.RequestLogEntry
	WSQueries           *websocket.Conn
	WSCommunication     *websocket.Conn
	WSCommunicationLock sync.Mutex
	dnsClient           *dns.Client
	Notifications       *notification.Manager
	Alerts              *alert.Manager
	Audits              *audit.Manager
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
	ResponseWriter dns.ResponseWriter
	Msg            *dns.Msg
	Question       dns.Question
	Sent           time.Time
	Client         *model.Client
	Prefetch       bool
	Protocol       model.Protocol
}

type communicationMessage struct {
	Client   bool   `json:"client"`
	Upstream bool   `json:"upstream"`
	DNS      bool   `json:"dns"`
	Ip       string `json:"ip"`
}

func NewDNSServer(config *settings.Config, dbManager *database.DatabaseManager, notificationsManager *notification.Manager, alertManager *alert.Manager, auditManager *audit.Manager, cert tls.Certificate) (*DNSServer, error) {
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
		lastLogTime:        time.Now(),
		logEntryChannel:    make(chan model.RequestLogEntry, 1000),
		dnsClient:          &client,
		Notifications:      notificationsManager,
		Alerts:             alertManager,
		Audits:             auditManager,
	}

	return server, nil
}

func (s *DNSServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) != 1 {
		log.Warning("Query container more than one question, ignoring!")
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
		Ip:       client.IP,
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
		Ip:       client.IP,
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

func (s *DNSServer) PopulateHostnameCache() (err error) {
	rows, err := s.DBManager.Conn.Query(`
		SELECT DISTINCT client_ip, client_name
		FROM request_log
		WHERE client_name IS NOT NULL AND client_name != 'unknown'
	`)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := rows.Close()
		if err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		var ip, name string
		if scanErr := rows.Scan(&ip, &name); scanErr != nil {
			return scanErr
		}

		if _, exists := s.hostnameCache.Load(name); !exists {
			s.hostnameCache.Store(name, ip)
		}
	}

	return rows.Err()
}

func (s *DNSServer) WSCom(message communicationMessage) {
	if s.WSCommunication == nil {
		return
	}

	s.WSCommunicationLock.Lock()
	defer s.WSCommunicationLock.Unlock()

	if err := s.WSCommunication.WriteControl(
		websocket.PingMessage,
		nil,
		time.Now().Add(2*time.Second),
	); err != nil {
		log.Debug("Websocket connection not alive, skipping message: %v", err)
		return
	}

	entryWSJson, err := json.Marshal(message)
	if err != nil {
		log.Error("Failed to marshal websocket message: %v", err)
		return
	}

	if err := s.WSCommunication.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		log.Warning("Failed to set websocket write deadline: %v", err)
	}

	if err := s.WSCommunication.WriteMessage(websocket.TextMessage, entryWSJson); err != nil {
		log.Debug("Failed to write websocket message: %v", err)
	}
}
