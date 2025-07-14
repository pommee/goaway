package server

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/miekg/dns"
)

func (s *DNSServer) InitDoT(cert tls.Certificate) (*dns.Server, error) {
	notifyReady := func() {
		log.Info("Started DoT (dns-over-tls) server on port %d", s.Config.DNS.DoTPort)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	server := &dns.Server{
		Addr:              fmt.Sprintf("%s:%d", s.Config.DNS.Address, s.Config.DNS.DoTPort),
		Net:               "tcp-tls",
		Handler:           s,
		TLSConfig:         tlsConfig,
		ReusePort:         true,
		UDPSize:           s.Config.DNS.UDPSize,
		NotifyStartedFunc: notifyReady,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	return server, nil
}
