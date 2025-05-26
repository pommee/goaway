package server

import (
	"encoding/json"
	"goaway/backend/dns/database"
	model "goaway/backend/dns/server/models"
	"time"

	"github.com/gorilla/websocket"
)

const BatchSize = 1000

func (s *DNSServer) ProcessLogEntries() {
	var batch []model.RequestLogEntry
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case entry := <-s.logEntryChannel:
			log.Debug("%+v", entry)
			if s.WSQueries != nil {
				entryWSJson, _ := json.Marshal(entry)
				_ = s.WSQueries.WriteMessage(websocket.TextMessage, entryWSJson)
			}

			batch = append(batch, entry)
			if len(batch) >= BatchSize {
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
	s.DBManager.Mutex.Lock()
	err := database.SaveRequestLog(s.DBManager.Conn, entries)
	s.DBManager.Mutex.Unlock()
	if err != nil {
		log.Warning("Error while saving logs, reason: %v", err)
	}
}

func (s *DNSServer) ClearOldEntries() {
	const (
		maxRetries      = 10
		retryDelay      = 150 * time.Millisecond
		cleanupInterval = 5 * time.Minute
	)

	for {
		requestThreshold := ((60 * 60) * 24) * s.Config.StatisticsRetention
		log.Debug("Next cleanup running at %s", time.Now().Add(cleanupInterval).Format(time.DateTime))
		time.Sleep(cleanupInterval)

		database.DeleteRequestLogsTimebased(s.Blacklist.Vacuum, s.DBManager.Conn, requestThreshold, maxRetries, retryDelay)
	}
}
