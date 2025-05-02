package main

import (
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

var server = "localhost:6121"

func BenchmarkDNSRequest(b *testing.B) {
	c := new(dns.Client)

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("google.com"), dns.TypeA)
	m.RecursionDesired = true

	b.ReportAllocs()
	b.ResetTimer()
	b.StartTimer()
	for b.Loop() {
		_, _, err := c.Exchange(m, server)
		if err != nil {
			b.Errorf("DNS query failed: %v", err)
		}
	}
}

func BenchmarkDNSRequestParallel(b *testing.B) {
	c := new(dns.Client)
	c.Timeout = 2 * time.Second

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("google.com"), dns.TypeA)
	m.RecursionDesired = true

	b.ReportAllocs()
	b.ResetTimer()
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, err := c.Exchange(m, server)
			if err != nil {
				b.Errorf("DNS query failed: %v", err)
			}
		}
	})
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
