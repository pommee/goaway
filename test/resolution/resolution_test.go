package resolution

import (
	"testing"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/stretchr/testify/require"

	"goaway/backend/dns/server"
	"goaway/backend/resolution"
	"goaway/backend/settings"
)

type fakeRepo struct {
	entries map[string]string
}

func (r *fakeRepo) CreateResolution(ip, domain string) error        { return nil }
func (r *fakeRepo) FindResolution(domain string) (string, error)    { return r.entries[domain], nil }
func (r *fakeRepo) FindResolutions() (map[string]string, error)     { return r.entries, nil }
func (r *fakeRepo) DeleteResolution(ip, domain string) (int, error) { return 0, nil }

func newServer(entries map[string]string) *server.DNSServer {
	cfg := &settings.Config{}
	cfg.DNS.CacheTTL = 60
	return &server.DNSServer{
		Config:            cfg,
		ResolutionService: resolution.NewService(&fakeRepo{entries: entries}),
	}
}

func newRequest(name string, qtype dns.RR) *server.Request {
	return &server.Request{
		Msg:      new(dns.Msg),
		Question: qtype,
	}
}

func TestResolveLocalEntryRespectsQueryType(t *testing.T) {
	s := newServer(map[string]string{"test.local.": "10.0.10.3"})

	aReq := newRequest("test.local.", &dns.A{Hdr: dns.Header{Name: "test.local."}})
	answers, _, status := s.Resolve(aReq)
	require.Equal(t, dnsutil.CodeToString(dns.RcodeSuccess), status)
	require.Len(t, answers, 1)
	_, ok := answers[0].(*dns.A)
	require.True(t, ok)

	aaaaReq := newRequest("test.local.", &dns.AAAA{Hdr: dns.Header{Name: "test.local."}})
	answers, _, status = s.Resolve(aaaaReq)
	require.Equal(t, dnsutil.CodeToString(dns.RcodeSuccess), status)
	require.Empty(t, answers)
}
