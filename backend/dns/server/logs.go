package server

import (
	"encoding/json"
	model "goaway/backend/dns/server/models"
	"time"

	"github.com/gorilla/websocket"
)

const batchSize = 1000

func (s *DNSServer) ProcessLogEntries() {
	var batch []model.RequestLogEntry
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case entry := <-s.logEntryChannel:
			log.Debug("%s", entry.String())
			if s.WSQueries != nil {
				entryWSJson, _ := json.Marshal(entry)
				_ = s.WSQueries.WriteMessage(websocket.TextMessage, entryWSJson)
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
	err := s.RequestService.SaveRequestLog(entries)
	if err != nil {
		log.Warning("Error while saving logs, reason: %v", err)
	}
}

// Removes old log entries based on the configured retention period.
func (s *DNSServer) ClearOldEntries() {
	const (
		maxRetries      = 10
		retryDelay      = 150 * time.Millisecond
		cleanupInterval = 5 * time.Minute
	)

	for {
		requestThreshold := ((60 * 60) * 24) * s.Config.Misc.StatisticsRetention
		log.Debug("Next cleanup running at %s", time.Now().Add(cleanupInterval).Format(time.DateTime))
		time.Sleep(cleanupInterval)

		s.RequestService.DeleteRequestLogsTimebased(s.BlacklistService.Vacuum, requestThreshold, maxRetries, retryDelay)
	}
}
