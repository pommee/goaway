package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"codeberg.org/miekg/dns"
)

func (s *DNSServer) InitDoT(cert tls.Certificate) (*dns.Server, error) {
	notifyReady := func(context.Context) {
		log.Info("Started DoT (dns-over-tls) server on port %d", s.Config.DNS.Ports.DoT)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
	server := &dns.Server{
		Addr:              fmt.Sprintf("%s:%d", s.Config.DNS.Address, s.Config.DNS.Ports.DoT),
		Net:               "tcp",
		Handler:           s,
		TLSConfig:         tlsConfig,
		ReusePort:         true,
		UDPSize:           s.Config.DNS.UDPSize,
		NotifyStartedFunc: notifyReady,
		ReadTimeout:       5 * time.Second,
	}

	return server, nil
}
