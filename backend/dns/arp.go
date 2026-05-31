package arp

import (
	"context"
	"encoding/json"
	"fmt"
	"goaway/backend/logging"
	"io"
	"net/http"
	"net/netip"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

var log = logging.GetLogger()

type vendorResponse struct {
	Company string `json:"company"`
	Success bool   `json:"success"`
	Found   bool   `json:"found"`
}

type Cache struct {
	table map[string]string
	mu    sync.RWMutex
}

type vendorCacheEntry struct {
	vendor    string
	err       error
	timestamp time.Time
}

type VendorCache struct {
	entries map[string]*vendorCacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
}

var (
	cache       = &Cache{table: make(map[string]string)}
	vendorCache = &VendorCache{
		entries: make(map[string]*vendorCacheEntry),
		ttl:     60 * time.Second,
	}
	httpClient = &http.Client{Timeout: 15 * time.Second}
)

func ProcessARPTable() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Update on first startup
	updateARPTable()

	for range ticker.C {
		updateARPTable()
	}
}

func CleanVendorResponseCache() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		vendorCache.cleanup()
	}
}

func updateARPTable() {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "arp", "-a")
	out, err := cmd.Output()
	if err != nil {
		log.Warning("Error running ARP command: %v", err)
		return
	}

	newTable := make(map[string]string)

	if runtime.GOOS != "windows" {
		parseUnixARP(string(out), newTable)
	} else {
		parseWindowsARP(string(out), newTable)
	}

	cache.mu.Lock()
	cache.table = newTable
	cache.mu.Unlock()
}

func parseWindowsARP(output string, table map[string]string) {
	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Interface:") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 {
			ip := fields[0]
			mac := strings.ToLower(strings.ReplaceAll(fields[1], "-", ":"))

			if isValidMAC(mac) {
				table[ip] = mac
			}
		}
	}
}

func parseUnixARP(output string, table map[string]string) {
	lines := strings.SplitSeq(output, "\n")
	for line := range lines {
		line = strings.Trim(line, " \t\r")
		if line == "" {
			continue
		}

		line = strings.ReplaceAll(line, "(", "")
		line = strings.ReplaceAll(line, ")", "")

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			ip := fields[1]
			mac := strings.ToLower(fields[3])
			if isValidMAC(mac) {
				table[ip] = mac
			}
		}
	}
}

func GetMacAddress(ip netip.Addr) string {
	cache.mu.RLock()
	mac, exists := cache.table[ip.String()]
	cache.mu.RUnlock()

	if exists {
		return mac
	}
	return "unknown"
}

func GetMacVendor(mac string) (string, error) {
	if mac == "" || mac == "unknown" {
		return "", fmt.Errorf("invalid MAC address")
	}

	mac = strings.ReplaceAll(mac, ":", "")
	mac = strings.ReplaceAll(mac, "-", "")
	mac = strings.ToLower(mac)

	if vendor, err, found := vendorCache.get(mac); found {
		return vendor, err
	}

	const maxAttempts = 3
	var lastErr error

	for attempt := range maxAttempts {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}

		vendor, err := fetchMacVendor(mac)
		if err == nil {
			vendorCache.set(mac, vendor, nil)
			return vendor, nil
		}
		lastErr = err
	}

	vendorCache.set(mac, "", lastErr)
	return "", lastErr
}

func fetchMacVendor(mac string) (string, error) {
	url := fmt.Sprintf("https://api.maclookup.app/v2/macs/%s", mac)
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch MAC vendor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result vendorResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !result.Found {
		return "", fmt.Errorf("vendor not found for mac %s", mac)
	}

	return result.Company, nil
}

func isValidMAC(mac string) bool {
	cleanMAC := strings.ReplaceAll(mac, ":", "")
	cleanMAC = strings.ReplaceAll(cleanMAC, "-", "")

	return len(cleanMAC) == 12 && cleanMAC != "000000000000"
}

func (vc *VendorCache) get(mac string) (string, error, bool) {
	vc.mu.RLock()
	defer vc.mu.RUnlock()

	entry, exists := vc.entries[mac]
	if !exists {
		return "", nil, false
	}

	if time.Since(entry.timestamp) > vc.ttl {
		return "", nil, false
	}

	return entry.vendor, entry.err, true
}

func (vc *VendorCache) set(mac, vendor string, err error) {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	vc.entries[mac] = &vendorCacheEntry{
		vendor:    vendor,
		err:       err,
		timestamp: time.Now(),
	}
}

func (vc *VendorCache) cleanup() {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	now := time.Now()
	for mac, entry := range vc.entries {
		if now.Sub(entry.timestamp) > vc.ttl {
			delete(vc.entries, mac)
		}
	}
}
