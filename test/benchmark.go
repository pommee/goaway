package main

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"codeberg.org/miekg/dns"
)

var (
	server        = "localhost:6121"
	questionCache *dns.Msg
	clientCache   *dns.Client
	once          sync.Once
)

func initCache() {
	questionCache = dns.NewMsg("google.com.", dns.TypeA)
	questionCache.RecursionDesired = true

	clientCache = dns.NewClient()
	clientCache.Dialer.Timeout = 2 * time.Second
}

func BenchmarkDNSRequest(b *testing.B) {
	once.Do(initCache)

	b.ReportAllocs()

	for b.Loop() {
		m := questionCache.Copy()
		_, _, err := clientCache.Exchange(context.TODO(), m, "udp", server)
		if err != nil {
			b.Errorf("DNS query failed: %v", err)
			continue
		}
	}
}

func BenchmarkDNSRequestParallel(b *testing.B) {
	once.Do(initCache)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		client := dns.NewClient()
		client.Dialer.Timeout = 2 * time.Second

		for pb.Next() {
			m := questionCache.Copy()

			_, _, err := client.Exchange(context.TODO(), m, "udp", server)
			if err != nil {
				b.Errorf("DNS query failed: %v", err)
				continue
			}
		}
	})
}

func TestDNSConnectivity(t *testing.T) {
	once.Do(initCache)

	r, _, err := clientCache.Exchange(context.TODO(), questionCache.Copy(), "udp", server)

	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			t.Fatalf("Timeout connecting to DNS server at %s", server)
		}
	}

	if r == nil || r.Rcode != dns.RcodeSuccess {
		t.Fatalf("DNS query failed with response code: %v", r.Rcode)
	}

	t.Logf("DNS server responded")
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
			{Name: "BenchmarkDNSRequestParallelWithConn", F: BenchmarkDNSRequestParallel},
		},
		[]testing.InternalExample{})
}
