package test

import (
	"goaway/internal/server"
	"testing"
	"time"

	"gotest.tools/assert"
)

var mockServerConfig = server.ServerConfig{
	Port:           9871,
	WebsitePort:    9090,
	UpstreamDNS:    "8.8.8.8:53",
	BlacklistPath:  "mock_blacklist.json",
	CountersFile:   "counters.json",
	RequestLogFile: "requests.json",
	CacheTTL:       time.Minute,
}

func mockServer() server.DNSServer {
	mockServer, _ := server.NewDNSServer(mockServerConfig)
	return mockServer
}

func TestIsBlacklisted(t *testing.T) {
	mockServer := mockServer()

	blockedDomains := []string{
		"blockeddomain.com",
		"another.blocked.domain.com",
	}
	allowedDomains := []string{
		"thisisfine.com",
		"jondoe.com",
	}

	for _, domain := range blockedDomains {
		assert.Equal(t, true, mockServer.Blacklist.Domains[domain])
	}

	for _, domain := range allowedDomains {
		assert.Equal(t, false, mockServer.Blacklist.Domains[domain])
	}

}
