package main

import (
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/miekg/dns"
)

var (
	server        = "localhost:6121"
	questionCache *dns.Msg
	clientCache   *dns.Client
	once          sync.Once
)

func initCache() {
	questionCache = new(dns.Msg)
	questionCache.SetQuestion(dns.Fqdn("google.com"), dns.TypeA)
	questionCache.RecursionDesired = true

	clientCache = new(dns.Client)
	clientCache.Timeout = 2 * time.Second
	clientCache.Net = "udp"
}

func BenchmarkDNSRequest(b *testing.B) {
	once.Do(initCache)

	b.ReportAllocs()

	for b.Loop() {
		m := questionCache.Copy()
		_, _, err := clientCache.Exchange(m, server)
		if err != nil {
			b.Errorf("DNS query failed: %v", err)
			continue
		}
	}
}

func BenchmarkDNSRequestWithConn(b *testing.B) {
	once.Do(initCache)

	conn, err := net.Dial("udp", server)
	if err != nil {
		b.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	dnsConn := &dns.Conn{Conn: conn}
	defer func(dnsConn *dns.Conn) {
		_ = dnsConn.Close()
	}(dnsConn)

	b.ReportAllocs()

	for b.Loop() {
		m := questionCache.Copy()

		if err := dnsConn.WriteMsg(m); err != nil {
			b.Errorf("Failed to send DNS query: %v", err)
			continue
		}

		_, err := dnsConn.ReadMsg()
		if err != nil {
			b.Errorf("Failed to read DNS response: %v", err)
			continue
		}
	}
}

func BenchmarkDNSRequestParallelWithConn(b *testing.B) {
	once.Do(initCache)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		conn, err := net.Dial("udp", server)
		if err != nil {
			b.Fatalf("Failed to create connection: %v", err)
		}
		defer func(conn net.Conn) {
			_ = conn.Close()
		}(conn)

		dnsConn := &dns.Conn{Conn: conn}
		defer func(dnsConn *dns.Conn) {
			_ = dnsConn.Close()
		}(dnsConn)

		for pb.Next() {
			m := questionCache.Copy()

			if err := dnsConn.WriteMsg(m); err != nil {
				b.Errorf("Failed to send DNS query: %v", err)
				continue
			}

			_, err := dnsConn.ReadMsg()
			if err != nil {
				b.Errorf("Failed to read DNS response: %v", err)
				continue
			}
		}
	})
}

func TestDNSConnectivity(t *testing.T) {
	once.Do(initCache)

	r, rtt, err := clientCache.Exchange(questionCache.Copy(), server)

	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			t.Fatalf("Timeout connecting to DNS server at %s", server)
		}
	}

	if r == nil || r.Rcode != dns.RcodeSuccess {
		t.Fatalf("DNS query failed with response code: %v", r.Rcode)
	}

	t.Logf("DNS server responded in %v", rtt)
	if len(r.Answer) > 0 {
		t.Logf("Received %d answers", len(r.Answer))
	} else {
		t.Logf("No answers in response")
	}
}

func main() {
	testing.Main(func(pat, str string) (bool, error) { return true, nil },
		[]testing.InternalTest{
			{Name: "TestDNSConnectivity", F: TestDNSConnectivity},
		},
		[]testing.InternalBenchmark{
			{Name: "BenchmarkDNSRequest", F: BenchmarkDNSRequest},
			{Name: "BenchmarkDNSRequestWithConn", F: BenchmarkDNSRequestWithConn},
			{Name: "BenchmarkDNSRequestParallelWithConn", F: BenchmarkDNSRequestParallelWithConn},
		},
		[]testing.InternalExample{})
}
