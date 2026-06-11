package resolution

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/netip"
	"testing"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"codeberg.org/miekg/dns/rdata"
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

func TestUpstreamQueryUsesRequestedType(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer pc.Close()

	gotType := make(chan uint16, 1)
	ready := make(chan struct{})
	srv := &dns.Server{
		PacketConn:        pc,
		NotifyStartedFunc: func(context.Context) { close(ready) },
		Handler: dns.HandlerFunc(func(_ context.Context, w dns.ResponseWriter, r *dns.Msg) {
			gotType <- dns.RRToType(r.Question[0])
			resp := new(dns.Msg)
			dnsutil.SetReply(resp, r)
			resp.Answer = []dns.RR{&dns.AAAA{
				Hdr:  dns.Header{Name: r.Question[0].Header().Name, TTL: 60, Class: dns.ClassINET},
				AAAA: rdata.AAAA{Addr: netip.MustParseAddr("2001:db8::1")},
			}}
			_, _ = io.Copy(w, resp)
		}),
	}
	go func() { _ = srv.ListenAndServe() }()
	defer srv.Shutdown(context.Background())
	<-ready

	cfg := &settings.Config{}
	cfg.DNS.CacheTTL = 60
	cfg.DNS.Upstream.Preferred = pc.LocalAddr().String()
	s, err := server.NewDNSServer(cfg, nil, tls.Certificate{})
	require.NoError(t, err)
	s.ResolutionService = resolution.NewService(&fakeRepo{entries: map[string]string{}})

	req := newRequest("example.com.", &dns.AAAA{Hdr: dns.Header{Name: "example.com."}})
	answers, _, _ := s.Resolve(req)

	require.Equal(t, dns.TypeAAAA, <-gotType)
	require.Len(t, answers, 1)
	_, ok := answers[0].(*dns.AAAA)
	require.True(t, ok)
}
