package server

import (
	"fmt"

	"github.com/miekg/dns"
)

func (s *DNSServer) InitTCP() (*dns.Server, error) {
	server := &dns.Server{
		Addr:      fmt.Sprintf("%s:%d", s.Config.DNS.Address, s.Config.DNS.Port),
		Net:       "tcp",
		Handler:   s,
		ReusePort: true,
		UDPSize:   s.Config.DNS.UDPSize,
	}

	return server, nil
}
