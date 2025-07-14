package server

import (
	"fmt"

	"github.com/miekg/dns"
)

type notifyDNSReady func()

func (s *DNSServer) InitUDP(notifyReady notifyDNSReady) (*dns.Server, error) {
	server := &dns.Server{
		Addr:              fmt.Sprintf("%s:%d", s.Config.DNS.Address, s.Config.DNS.Port),
		Net:               "udp",
		Handler:           s,
		ReusePort:         true,
		UDPSize:           s.Config.DNS.UDPSize,
		NotifyStartedFunc: notifyReady,
	}

	return server, nil
}
