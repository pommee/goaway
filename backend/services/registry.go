package services

import (
	"embed"
	"fmt"
	"goaway/backend/api"
	"goaway/backend/api/key"
	"goaway/backend/blacklist"
	"goaway/backend/logging"
	"goaway/backend/notification"
	"goaway/backend/prefetch"
	"goaway/backend/request"
	"goaway/backend/resolution"
	"goaway/backend/user"
	"goaway/backend/whitelist"
	"net/http"
	"sync"

	"github.com/miekg/dns"
)

var log = logging.GetLogger()

// Manages all servers and services
type ServiceRegistry struct {
	APIServer *api.API
	errorChan chan ServiceError
	dotServer *dns.Server
	readyChan chan struct{}
	content   embed.FS
	udpServer *dns.Server
	dohServer *http.Server
	tcpServer *dns.Server
	Context   *AppContext
	version   string
	date      string
	commit    string
	wg        sync.WaitGroup

	ResolutionService   *resolution.Service
	RequestService      *request.Service
	PrefetchService     *prefetch.Service
	UserService         *user.Service
	KeyService          *key.Service
	NotificationService *notification.Service
	BlacklistService    *blacklist.Service
	WhitelistService    *whitelist.Service
}

type ServiceError struct {
	Err     error
	Service string
}

func NewServiceRegistry(ctx *AppContext, version, commit, date string, content embed.FS) *ServiceRegistry {
	return &ServiceRegistry{
		Context:   ctx,
		version:   version,
		commit:    commit,
		date:      date,
		content:   content,
		readyChan: make(chan struct{}),
		errorChan: make(chan ServiceError, 10),
	}
}

func (r *ServiceRegistry) Initialize() error {
	r.setupDNSServers()

	if r.Context.Certificate.Certificate != nil {
		if err := r.setupSecureServers(); err != nil {
			return err
		}
	}

	r.setupAPIServer()

	return nil
}

func (r *ServiceRegistry) setupDNSServers() {
	config := r.Context.Config

	notifyReady := func() {
		log.Info("Started DNS server on: %s:%d", config.DNS.Address, config.DNS.Ports.TCPUDP)
		close(r.readyChan)
	}

	r.udpServer = &dns.Server{
		Addr:      fmt.Sprintf("%s:%d", config.DNS.Address, config.DNS.Ports.TCPUDP),
		Net:       "udp",
		Handler:   r.Context.DNSServer,
		ReusePort: true,
		UDPSize:   config.DNS.UDPSize,
	}

	r.tcpServer = &dns.Server{
		Addr:              fmt.Sprintf("%s:%d", config.DNS.Address, config.DNS.Ports.TCPUDP),
		Net:               "tcp",
		Handler:           r.Context.DNSServer,
		ReusePort:         true,
		UDPSize:           config.DNS.UDPSize,
		NotifyStartedFunc: notifyReady,
	}
}

func (r *ServiceRegistry) setupSecureServers() error {
	dotServer, err := r.Context.DNSServer.InitDoT(r.Context.Certificate)
	if err != nil {
		return fmt.Errorf("failed to initialize DoT server: %w", err)
	}
	r.dotServer = dotServer

	dohServer, err := r.Context.DNSServer.InitDoH(r.Context.Certificate)
	if err != nil {
		return fmt.Errorf("failed to initialize DoH server: %w", err)
	}
	r.dohServer = dohServer

	return nil
}

func (r *ServiceRegistry) setupAPIServer() {
	r.APIServer = &api.API{
		DNS:             r.Context.DNSServer,
		Authentication:  r.Context.Config.API.Authentication,
		Config:          r.Context.Config,
		DNSPort:         r.Context.Config.DNS.Ports.TCPUDP,
		Version:         r.version,
		Commit:          r.commit,
		Date:            r.date,
		DNSServer:       r.Context.DNSServer,
		DBConn:          r.Context.DBConn,
		WSQueries:       r.Context.DNSServer.WSQueries,
		WSCommunication: r.Context.DNSServer.WSCommunication,

		ResolutionService:   r.ResolutionService,
		RequestService:      r.RequestService,
		PrefetchService:     r.PrefetchService,
		NotificationService: r.NotificationService,
		UserService:         r.UserService,
		KeyService:          r.KeyService,
		BlacklistService:    r.BlacklistService,
		WhitelistService:    r.WhitelistService,
	}
}

func (r *ServiceRegistry) StartAll() {
	r.startDNSServers()

	if r.Context.Certificate.Certificate != nil {
		r.startSecureServers()
	}

	r.startAPIServer()
}

func (r *ServiceRegistry) startDNSServers() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		if err := r.udpServer.ListenAndServe(); err != nil {
			r.errorChan <- ServiceError{Service: "UDP", Err: err}
		}
	}()

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		if err := r.tcpServer.ListenAndServe(); err != nil {
			r.errorChan <- ServiceError{Service: "TCP", Err: err}
		}
	}()
}

func (r *ServiceRegistry) startSecureServers() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		if err := r.dotServer.ListenAndServe(); err != nil {
			r.errorChan <- ServiceError{Service: "DoT", Err: err}
		}
	}()

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		if serverIP, err := api.GetServerIP(); err == nil {
			log.Info("DoH (dns-over-https) server running at https://%s:%d/dns-query",
				serverIP, r.Context.Config.DNS.Ports.DoH)
		} else {
			log.Info("DoH (dns-over-https) server running on port :%d", r.Context.Config.DNS.Ports.DoH)
		}

		if err := r.dohServer.ListenAndServeTLS(
			r.Context.Config.DNS.TLS.Cert,
			r.Context.Config.DNS.TLS.Key,
		); err != nil {
			r.errorChan <- ServiceError{Service: "DoH", Err: err}
		}
	}()
}

func (r *ServiceRegistry) startAPIServer() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		<-r.readyChan

		errorChan := make(chan struct{}, 1)
		go func() {
			<-errorChan
			r.errorChan <- ServiceError{Service: "API", Err: fmt.Errorf("API server stopped")}
		}()

		r.APIServer.Start(r.content, errorChan)
	}()
}

func (r *ServiceRegistry) WaitGroup() *sync.WaitGroup {
	return &r.wg
}

func (r *ServiceRegistry) ReadyChannel() <-chan struct{} {
	return r.readyChan
}

func (r *ServiceRegistry) ErrorChannel() <-chan ServiceError {
	return r.errorChan
}

func (r *ServiceRegistry) GetPrefetcher() *prefetch.Service {
	return r.APIServer.PrefetchService
}
