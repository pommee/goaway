package server

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	model "goaway/backend/dns/server/models"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const (
	MaxDoHRequestSize = 4096
	DoHTimeout        = 20 * time.Second
	DoHReadTimeout    = 8 * time.Second
	DoHWriteTimeout   = 8 * time.Second
	MB                = 1 << 20
)

func (s *DNSServer) InitDoH(cert tls.Certificate) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns-query", s.handleDoHRequest)
	mux.HandleFunc("/health", s.handleHealthCheck)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.Config.DNS.Address, s.Config.DNS.DoHPort),
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS12,
			MaxVersion:               tls.VersionTLS13,
			PreferServerCipherSuites: true,
			NextProtos:               []string{"h2", "http/1.1"},
		},
		ReadTimeout:       DoHReadTimeout,
		WriteTimeout:      DoHWriteTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 * MB,
	}

	return server, nil
}

func (s *DNSServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
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
	ctx, cancel := context.WithTimeout(r.Context(), DoHTimeout)
	defer cancel()

	r = r.WithContext(ctx)

	log.Debug("DoH request received: %s %s from %s", r.Method, r.URL.String(), r.RemoteAddr)

	if r.ContentLength > MaxDoHRequestSize {
		log.Warning("DoH request too large: %d bytes from %s", r.ContentLength, r.RemoteAddr)
		http.Error(w, "Request too large", http.StatusRequestEntityTooLarge)
		return
	}

	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	go s.WSCom(communicationMessage{true, false, false, clientIP})

	var (
		dnsQuery []byte
		err      error
	)

	switch r.Method {
	case "GET":
		dnsQuery, err = s.handleDoHGet(r)
	case "POST":
		dnsQuery, err = s.handleDoHPost(r)
	default:
		log.Warning("DoH request invalid method: %s from %s", r.Method, r.RemoteAddr)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err != nil {
		log.Warning("DoH request processing failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	msg := new(dns.Msg)
	if err := msg.Unpack(dnsQuery); err != nil {
		log.Warning("DoH request invalid DNS message: %v", err)
		http.Error(w, "Invalid DNS message", http.StatusBadRequest)
		return
	}

	if len(msg.Question) == 0 {
		log.Warning("DoH request missing question section")
		http.Error(w, "Missing question section", http.StatusBadRequest)
		return
	}

	msg.Compress = true

	responseWriter := &DoHResponseWriter{
		httpWriter: w,
		remoteAddr: r.RemoteAddr,
		DoHPort:    s.Config.DNS.DoHPort,
	}

	req := &Request{
		ResponseWriter: responseWriter,
		Msg:            msg,
		Question:       msg.Question[0],
		Sent:           time.Now(),
		Client:         s.getClientInfo(r.RemoteAddr),
		Prefetch:       false,
		Protocol:       model.DoH,
	}

	logEntry := s.processQuery(req)

	go s.WSCom(communicationMessage{false, false, true, clientIP})

	select {
	case s.logEntryChannel <- logEntry:
	case <-time.After(1 * time.Second):
		log.Warning("Log entry channel full, dropping log entry")
	}
}

func (s *DNSServer) handleDoHGet(r *http.Request) ([]byte, error) {
	dnsParam := r.URL.Query().Get("dns")
	if dnsParam == "" {
		return nil, fmt.Errorf("missing dns parameter")
	}

	if len(dnsParam) > MaxDoHRequestSize {
		return nil, fmt.Errorf("dns parameter too long")
	}

	dnsQuery, err := base64.RawURLEncoding.DecodeString(dnsParam)
	if err != nil {
		return nil, fmt.Errorf("invalid dns parameter: %w", err)
	}

	if len(dnsQuery) == 0 {
		return nil, fmt.Errorf("empty dns query")
	}

	return dnsQuery, nil
}

func (s *DNSServer) handleDoHPost(r *http.Request) ([]byte, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/dns-message" {
		return nil, fmt.Errorf("invalid content type: %s", contentType)
	}

	limitedReader := io.LimitReader(r.Body, MaxDoHRequestSize)
	dnsQuery, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	if len(dnsQuery) == 0 {
		return nil, fmt.Errorf("empty request body")
	}

	return dnsQuery, nil
}

type DoHResponseWriter struct {
	httpWriter http.ResponseWriter
	msg        *dns.Msg
	remoteAddr string
	mu         sync.Mutex
	DoHPort    int
}

func (w *DoHResponseWriter) LocalAddr() net.Addr {
	addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("127.0.0.1:%d", w.DoHPort))
	return addr
}

func (w *DoHResponseWriter) RemoteAddr() net.Addr {
	addr, err := net.ResolveTCPAddr("tcp", w.remoteAddr)
	if err != nil {
		log.Warning("Failed to resolve remote address %s: %v", w.remoteAddr, err)
		addr, _ = net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	}
	return addr
}

func (w *DoHResponseWriter) WriteMsg(msg *dns.Msg) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.msg = msg

	responseBytes, err := msg.Pack()
	if err != nil {
		log.Error("DoH failed to pack response: %v", err)
		return err
	}

	w.httpWriter.Header().Set("Content-Type", "application/dns-message")
	w.httpWriter.Header().Set("Cache-Control", "max-age=0")
	w.httpWriter.Header().Set("Content-Length", fmt.Sprintf("%d", len(responseBytes)))
	w.httpWriter.WriteHeader(http.StatusOK)

	n, err := w.httpWriter.Write(responseBytes)
	if err != nil {
		log.Error("DoH failed to write response: %v", err)
		return err
	}

	log.Debug("DoH response sent: %d bytes written", n)
	return nil
}

func (w *DoHResponseWriter) Write(b []byte) (int, error) {
	return w.httpWriter.Write(b)
}

func (w *DoHResponseWriter) Close() error {
	return nil
}

func (w *DoHResponseWriter) TsigStatus() error {
	return nil
}

func (w *DoHResponseWriter) TsigTimersOnly(bool) {}

func (w *DoHResponseWriter) Hijack() {}

func (w *DoHResponseWriter) Header() http.Header { return nil }

func (w *DoHResponseWriter) Network() string { return "tcp" }

func (w *DoHResponseWriter) WriteHeader(statusCode int) {}
