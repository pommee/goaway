package server

import (
	"context"
	"crypto/tls"
	"fmt"
	model "goaway/backend/dns/server/models"
	"net"
	"net/http"
	"time"

	"codeberg.org/miekg/dns/dnshttp"
)

const (
	doHReadTimeout  = 8 * time.Second
	doHWriteTimeout = 8 * time.Second
	megabyte        = 1 << 20
)

func (s *DNSServer) InitDoH(cert tls.Certificate) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc(dnshttp.Path, s.handleDoHRequest)
	mux.HandleFunc("/health", s.handleHealthCheck)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.Config.DNS.Address, s.Config.DNS.Ports.DoH),
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS12,
			MaxVersion:               tls.VersionTLS13,
			PreferServerCipherSuites: true,
			NextProtos:               dnshttp.NextProtos,
		},
		ReadTimeout:       doHReadTimeout,
		WriteTimeout:      doHWriteTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 * megabyte,
	}

	return server, nil
}

func (s *DNSServer) handleHealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status":"healthy"}`))
	if err != nil {
		log.Error("Failed to write health check response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (s *DNSServer) handleDoHRequest(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	ctxVal := context.WithValue(ctx, model.DoH, true)
	defer cancel()

	log.Debug("DoH request received: %s %s from %s", r.Method, r.URL.String(), r.RemoteAddr)

	// Parse the HTTP request into a DNS message
	msg, err := dnshttp.Request(r)
	if err != nil {
		log.Warning("DoH request parsing failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Local address for the ResponseWriter
	laddr := r.Context().Value(http.LocalAddrContextKey).(net.Addr)

	// DNS ResponseWriter wrapper for HTTP
	hw := dnshttp.NewResponseWriter(w, r, laddr)

	s.ServeDNS(ctxVal, hw, msg)
}
