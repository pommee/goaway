package server

import (
	"encoding/json"
	"goaway/internal/database"
	model "goaway/internal/server/models"
	"time"

	"github.com/gorilla/websocket"
)

func (s *DNSServer) ProcessLogEntries() {
	var batch []model.RequestLogEntry
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case entry := <-s.logEntryChannel:
			if s.WS != nil {
				wsMutex.Lock()
				entryWSJson, _ := json.Marshal(entry)
				_ = s.WS.WriteMessage(websocket.TextMessage, entryWSJson)
				wsMutex.Unlock()
			}

			batch = append(batch, entry)
			if len(batch) >= batchSize {
				s.saveBatch(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.saveBatch(batch)
				batch = nil
			}
		}
	}
}

func (s *DNSServer) saveBatch(entries []model.RequestLogEntry) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	database.SaveRequestLog(s.DB, entries)
}

func (s *DNSServer) ClearOldEntries() {
	const (
		maxRetries      = 10
		retryDelay      = 150 * time.Millisecond
		cleanupInterval = 1 * time.Minute
	)

	for {
		requestThreshold := ((60 * 60) * 24) * s.StatisticsRetention
		log.Debug("Next cleanup running at %s", time.Now().Add(cleanupInterval).Format(time.DateTime))
		time.Sleep(cleanupInterval)

		database.DeleteRequestLogsTimebased(s.DB, requestThreshold, maxRetries, retryDelay)
		s.UpdateCounters()
	}
}

func (s *DNSServer) UpdateCounters() {
	blockedCount, allowedCount, err := database.CountAllowedAndBlockedRequest(s.DB)
	if err != nil {
		log.Error("%s %v", "failed to get counters: ", err)
	}
	s.Counters.BlockedRequests = blockedCount
	s.Counters.AllowedRequests = allowedCount
}
