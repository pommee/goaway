package arp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

var arpTable = map[string]string{}

func ProcessARPTable() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Update first time server is started
	updateARPTable()

	for {
		select {
		case <-ticker.C:
			updateARPTable()
		}
	}
}

func updateARPTable() {
	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		fmt.Println("Error running ARP command:", err)
		return
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.ReplaceAll(line, "(", "")
		line = strings.ReplaceAll(line, ")", "")

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			ip := fields[1]
			mac := fields[3]
			arpTable[ip] = mac
		}
	}
}

func GetMacAddress(ip string) string {
	mac, exists := arpTable[ip]
	if exists {
		return mac
	}

	return "unknown"
}

func GetMacVendor(mac string) (string, error) {
	url := fmt.Sprintf("https://api.maclookup.app/v2/macs/%s", mac)
	resp, err := http.Get(url)
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

	var result struct {
		Success bool   `json:"success"`
		Found   bool   `json:"found"`
		Company string `json:"company"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Found {
		return result.Company, nil
	}

	return "", fmt.Errorf("vendor not found")
}
