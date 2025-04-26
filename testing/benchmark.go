package main

import (
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/miekg/dns"
)

var server = "localhost:6121"

func BenchmarkDNSRequest(b *testing.B) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("google.com"), dns.TypeA)
	m.RecursionDesired = true

	var totalDuration time.Duration
	b.ReportAllocs()

	for b.Loop() {
		start := time.Now()
		c := new(dns.Client)
		c.Timeout = 2 * time.Second
		_, _, err := c.Exchange(m, server)
		if err != nil {
			b.Fatalf("DNS query failed: %v", err)
		}
		totalDuration += time.Since(start)
	}

	avgDuration := totalDuration / time.Duration(b.N)
	b.ReportMetric(float64(avgDuration.Microseconds()), "avg_us/op")
}

func BenchmarkDNSRequestParallel(b *testing.B) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("google.com"), dns.TypeA)
	m.RecursionDesired = true

	var totalDuration int64
	b.ReportAllocs()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		c := new(dns.Client)
		c.Timeout = 2 * time.Second
		for pb.Next() {
			start := time.Now()
			_, _, err := c.Exchange(m, server)
			if err != nil {
				b.Fatalf("DNS query failed: %v", err)
			}
			atomic.AddInt64(&totalDuration, time.Since(start).Nanoseconds())
		}
	})

	avgDuration := time.Duration(totalDuration / int64(b.N))
	b.ReportMetric(float64(avgDuration.Microseconds()), "avg_us/op")
}

func TestDNSConnectivity(t *testing.T) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("google.com"), dns.TypeA)
	m.RecursionDesired = true

	c := new(dns.Client)
	c.Timeout = 2 * time.Second
	r, rtt, err := c.Exchange(m, server)

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			t.Fatalf("Timeout connecting to DNS server at %s", server)
		} else {
			t.Fatalf("Error connecting to DNS server: %v", err)
		}
	}

	t.Logf("DNS server responded in %v", rtt)
	t.Logf("Response: %v", r)
}

func main() {
	testing.Main(func(pat, str string) (bool, error) { return true, nil },
		[]testing.InternalTest{},
		[]testing.InternalBenchmark{
			{Name: "BenchmarkDNSRequest", F: BenchmarkDNSRequest},
			{Name: "BenchmarkDNSRequestParallel", F: BenchmarkDNSRequestParallel},
		},
		[]testing.InternalExample{})
}
