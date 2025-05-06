package server

import (
	"encoding/json"
	"goaway/backend/dns/database"
	model "goaway/backend/dns/server/models"
	"time"

	"github.com/gorilla/websocket"
)

const BATCH_SIZE = 500

func (s *DNSServer) ProcessLogEntries() {
	var batch []model.RequestLogEntry
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case entry := <-s.logEntryChannel:
			log.Debug("%+v", entry)
			if s.WS != nil {
				entryWSJson, _ := json.Marshal(entry)
				_ = s.WS.WriteMessage(websocket.TextMessage, entryWSJson)
			}

			batch = append(batch, entry)
			if len(batch) >= BATCH_SIZE {
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
	database.SaveRequestLog(s.DB, entries)
	dbMutex.Unlock()
}

func (s *DNSServer) ClearOldEntries() {
	const (
		maxRetries      = 10
		retryDelay      = 150 * time.Millisecond
		cleanupInterval = 1 * time.Minute
	)

	for {
		requestThreshold := ((60 * 60) * 24) * s.Config.StatisticsRetention
		log.Debug("Next cleanup running at %s", time.Now().Add(cleanupInterval).Format(time.DateTime))
		time.Sleep(cleanupInterval)

		database.DeleteRequestLogsTimebased(s.Blacklist.Vacuum, s.DB, requestThreshold, maxRetries, retryDelay)
	}
}
