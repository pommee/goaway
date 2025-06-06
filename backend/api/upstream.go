package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	probing "github.com/prometheus-community/pro-bing"
)

func (api *API) registerUpstreamRoutes() {
	api.routes.POST("/upstream", api.createUpstream)

	api.routes.GET("/upstreams", api.getUpstreams)

	api.routes.PUT("/preferredUpstream", api.updatePreferredUpstream)

	api.routes.DELETE("/upstream", api.deleteUpstream)
}

func (api *API) createUpstream(c *gin.Context) {
	type UpstreamRequest struct {
		Upstream string `json:"upstream"`
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request UpstreamRequest
	if err := json.Unmarshal(body, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	upstream := request.Upstream
	if upstream == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream is required"})
		return
	}

	if !strings.Contains(upstream, ":") {
		upstream += ":53"
	}

	if slices.Contains(api.Config.DNS.UpstreamDNS, upstream) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream already exists"})
		return
	}

	api.Config.DNS.UpstreamDNS = append(api.Config.DNS.UpstreamDNS, upstream)
	api.Config.Save()

	log.Info("Added %s as a new upstream", upstream)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Added %s as a new upstream", upstream)})
}

func (api *API) getUpstreams(c *gin.Context) {
	upstreams := api.Config.DNS.UpstreamDNS
	results := make([]map[string]any, len(upstreams))

	preferredUpstream := api.Config.DNS.PreferredUpstream

	var wg sync.WaitGroup
	wg.Add(len(upstreams))

	for i, upstream := range upstreams {
		go func(i int, upstream string) {
			defer wg.Done()
			results[i] = getUpstreamDetails(upstream, preferredUpstream)
		}(i, upstream)
	}

	wg.Wait()

	c.JSON(http.StatusOK, gin.H{
		"upstreams":         results,
		"preferredUpstream": preferredUpstream,
	})
}

func getUpstreamDetails(upstream, preferredUpstream string) map[string]any {
	host := strings.TrimSuffix(upstream, ":53")
	entry := map[string]any{
		"upstream":  upstream,
		"preferred": upstream == preferredUpstream,
	}

	entry["name"], entry["dnsPing"] = getDNSDetails(host)
	entry["icmpPing"] = getICMPPing(host)

	return entry
}

func getICMPPing(host string) string {
	pinger, err := probing.NewPinger(host)
	if err == nil {
		pinger.Count = 3
		pinger.Timeout = 2 * time.Second
		pinger.SetPrivileged(false)

		err = pinger.Run()
		if err == nil {
			stats := pinger.Statistics()
			return stats.AvgRtt.String()
		}
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", host+":53", 2*time.Second)
	if err != nil {
		log.Warning("Could not ping host %s via ICMP or TCP: %s", host, err.Error())
		return "Unreachable"
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	duration := time.Since(start)
	return duration.String()
}

func getDNSDetails(host string) (string, string) {
	start := time.Now()
	ips, err := net.LookupIP(host)
	duration := time.Since(start)

	if err != nil {
		return "Error: " + err.Error(), "Error: " + err.Error()
	}
	if len(ips) > 0 {
		return ips[0].String(), duration.String()
	}
	return "No IP found", duration.String()
}

func (api *API) updatePreferredUpstream(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Error("Failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var request struct {
		Upstream string `json:"upstream"`
	}

	if err := json.Unmarshal(body, &request); err != nil {
		log.Error("Failed to parse JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if request.Upstream == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Upstream is required"})
		return
	}

	if !slices.Contains(api.Config.DNS.UpstreamDNS, request.Upstream) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Upstream not found"})
		return
	}

	if api.Config.DNS.PreferredUpstream == request.Upstream {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Preferred upstream already set to %s", request.Upstream)})
		return
	}

	api.Config.DNS.PreferredUpstream = request.Upstream
	message := fmt.Sprintf("Preferred upstream set to %s", request.Upstream)
	log.Info("%s", message)

	api.Config.Save()
	c.JSON(http.StatusOK, gin.H{"message": message})
}

func (api *API) deleteUpstream(c *gin.Context) {
	upstreamToDelete := c.Query("upstream")

	if upstreamToDelete == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'upstream' query parameter"})
		return
	}

	var updatedUpstreams []string
	for _, upstream := range api.Config.DNS.UpstreamDNS {
		if upstream != upstreamToDelete {
			updatedUpstreams = append(updatedUpstreams, upstream)
		}
	}

	api.Config.DNS.UpstreamDNS = updatedUpstreams
	api.Config.Save()

	c.JSON(http.StatusOK, gin.H{
		"message": "Upstream removed successfully",
	})
}
