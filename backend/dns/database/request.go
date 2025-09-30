package database

import (
	"database/sql"
	"fmt"
	"goaway/backend/api/models"
	dbModel "goaway/backend/dns/database/models"
	model "goaway/backend/dns/server/models"
	"goaway/backend/logging"
	"strings"
	"time"

	"gorm.io/gorm"
)

var log = logging.GetLogger()

func GetClientNameFromRequestLog(db *gorm.DB, ip string) string {
	var hostname string

	err := db.Model(&RequestLog{}).
		Select("client_name").
		Where("client_ip = ? AND client_name != ?", ip, "unknown").
		Limit(1).
		Scan(&hostname).Error

	if err != nil || hostname == "" {
		return "unknown"
	}

	return strings.TrimSuffix(hostname, ".")
}

func GetDistinctRequestIP(db *gorm.DB) int {
	var count int64

	err := db.Model(&RequestLog{}).
		Select("COUNT(DISTINCT client_ip)").
		Scan(&count).Error
	if err != nil {
		return 0
	}

	return int(count)
}

func GetRequestSummaryByInterval(interval int, db *gorm.DB) ([]model.RequestLogIntervalSummary, error) {
	minutes := interval * 60

	var rawSummaries []model.RequestLogIntervalSummary
	err := db.Table("request_logs").
		Select(`
			DATETIME((STRFTIME('%s', timestamp) / ?) * ?, 'unixepoch') AS interval_start,
			SUM(blocked) AS blocked_count,
			SUM(cached) AS cached_count,
			SUM(NOT blocked AND NOT cached) AS allowed_count
		`, minutes, minutes, minutes).
		Where("timestamp >= DATETIME('now', '-1 day')").
		Group("(STRFTIME('%s', timestamp) / ?)").
		Order("interval_start").
		Scan(&rawSummaries).Error
	if err != nil {
		return nil, err
	}

	summaries := make([]model.RequestLogIntervalSummary, len(rawSummaries))
	for i := range rawSummaries {
		t, err := time.Parse("2006-01-02 15:04:05", rawSummaries[i].IntervalStart)
		if err != nil {
			return nil, err
		}
		summaries[i] = model.RequestLogIntervalSummary{
			IntervalStart: t.String(),
			BlockedCount:  rawSummaries[i].BlockedCount,
			CachedCount:   rawSummaries[i].CachedCount,
			AllowedCount:  rawSummaries[i].AllowedCount,
		}
	}

	return summaries, nil
}

func GetResponseSizeSummaryByInterval(intervalMinutes int, db *gorm.DB) ([]model.ResponseSizeSummary, error) {
	intervalSeconds := int64(intervalMinutes * 60)
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour).Unix()

	var summaries []model.ResponseSizeSummary

	query := `
		SELECT
			((strftime('%s', timestamp) / ?) * ?) AS start_unix,
			SUM(response_size_bytes) AS total_size_bytes,
			ROUND(AVG(response_size_bytes)) AS avg_response_size_bytes,
			MIN(response_size_bytes) AS min_response_size_bytes,
			MAX(response_size_bytes) AS max_response_size_bytes
		FROM request_logs
		WHERE strftime('%s', timestamp) >= ? AND response_size_bytes IS NOT NULL
		GROUP BY (strftime('%s', timestamp) / ?)
		ORDER BY start_unix ASC
	`

	err := db.Raw(query, intervalSeconds, intervalSeconds, twentyFourHoursAgo, intervalSeconds).
		Scan(&summaries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query response size summary: %w", err)
	}

	for i := range summaries {
		summaries[i].Start = time.Unix(summaries[i].StartUnix, 0)
	}

	return summaries, nil
}

func GetUniqueQueryTypes(db *gorm.DB) ([]interface{}, error) {
	query := `
		SELECT COUNT(*) AS count, query_type
		FROM request_logs
		WHERE query_type <> ''
		GROUP BY query_type
		ORDER BY count DESC`

	rows, err := db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			log.Error("Failed to close rows: %v", err)
		}
	}(rows)

	var queries []any
	for rows.Next() {
		query := struct {
			Count     int    `json:"count"`
			QueryType string `json:"queryType"`
		}{}

		if err := rows.Scan(&query.Count, &query.QueryType); err != nil {
			return nil, err
		}

		queries = append(queries, query)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return queries, nil
}

func FetchQueries(db *gorm.DB, q models.QueryParams) ([]model.RequestLogEntry, error) {
	var logs []RequestLog
	query := db.Model(&RequestLog{})

	if q.Column == "ip" {
		query = query.Joins("LEFT JOIN request_log_ips ri ON request_logs.id = ri.request_log_id")
	}

	if q.Search != "" {
		query = query.Where("request_logs.domain LIKE ?", "%"+q.Search+"%")
	}

	if q.FilterClient != "" {
		query = query.Where("request_logs.client_ip LIKE ? OR request_logs.client_name LIKE ?",
			"%"+q.FilterClient+"%", "%"+q.FilterClient+"%")
	}

	if q.Column == "ip" {
		query = query.Group("request_logs.id").Order("MAX(ri.ip) " + q.Direction)
	} else {
		query = query.Order("request_logs." + q.Column + " " + q.Direction)
	}

	if q.PageSize > 0 {
		query = query.Limit(q.PageSize)
	}
	if q.Offset > 0 {
		query = query.Offset(q.Offset)
	}

	query = query.Preload("IPs")

	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}

	results := make([]model.RequestLogEntry, len(logs))
	for i, log := range logs {
		results[i] = model.RequestLogEntry{
			ID:                int64(log.ID),
			Timestamp:         log.Timestamp,
			Domain:            log.Domain,
			Blocked:           log.Blocked,
			Cached:            log.Cached,
			ResponseTime:      time.Duration(log.ResponseTimeNs),
			ClientInfo:        &model.Client{IP: log.ClientIP, Name: log.ClientName},
			Status:            log.Status,
			QueryType:         log.QueryType,
			ResponseSizeBytes: log.ResponseSizeBytes,
			Protocol:          model.Protocol(log.Protocol),
			IP:                make([]model.ResolvedIP, len(log.IPs)),
		}

		for j, ip := range log.IPs {
			results[i].IP[j] = model.ResolvedIP{IP: ip.IP, RType: ip.RType}
		}
	}

	return results, nil
}

func FetchAllClients(db *gorm.DB) (map[string]dbModel.Client, error) {
	var rows []struct {
		ClientIP   string         `gorm:"column:client_ip"`
		ClientName string         `gorm:"column:client_name"`
		Timestamp  time.Time      `gorm:"column:timestamp"`
		Mac        sql.NullString `gorm:"column:mac"`
		Vendor     sql.NullString `gorm:"column:vendor"`
	}

	subquery := db.Table("request_logs").
		Select("client_ip, MAX(timestamp) as max_timestamp").
		Group("client_ip")

	if err := db.Table("request_logs r").
		Select("r.client_ip, r.client_name, r.timestamp, m.mac, m.vendor").
		Joins("INNER JOIN (?) latest ON r.client_ip = latest.client_ip AND r.timestamp = latest.max_timestamp", subquery).
		Joins("LEFT JOIN mac_addresses m ON r.client_ip = m.ip").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	uniqueClients := make(map[string]dbModel.Client, len(rows))
	for _, row := range rows {
		macStr := ""
		vendorStr := ""
		if row.Mac.Valid {
			macStr = row.Mac.String
		}
		if row.Vendor.Valid {
			vendorStr = row.Vendor.String
		}

		uniqueClients[row.ClientIP] = dbModel.Client{
			Name:     row.ClientName,
			LastSeen: row.Timestamp,
			Mac:      macStr,
			Vendor:   vendorStr,
		}
	}

	return uniqueClients, nil
}

func GetClientDetailsWithDomains(db *gorm.DB, clientIP string) (dbModel.ClientRequestDetails, string, map[string]int, error) {
	var crd dbModel.ClientRequestDetails
	err := db.Table("request_logs").
		Select(`
			COUNT(*) as total_requests,
			COUNT(DISTINCT domain) as unique_domains,
			SUM(CASE WHEN blocked THEN 1 ELSE 0 END) as blocked_requests,
			SUM(CASE WHEN cached THEN 1 ELSE 0 END) as cached_requests,
			AVG(response_time_ns) / 1e6 as avg_response_time_ms,
			MAX(timestamp) as last_seen`).
		Where("client_ip = ?", clientIP).
		Scan(&crd).Error

	if err != nil {
		return crd, "", nil, err
	}

	var rows []struct {
		Domain string `gorm:"column:domain"`
		Count  int    `gorm:"column:query_count"`
	}

	err = db.Table("request_logs").
		Select("domain, COUNT(*) as query_count").
		Where("client_ip = ?", clientIP).
		Group("domain").
		Order("query_count DESC").
		Scan(&rows).Error

	if err != nil {
		return crd, "", nil, err
	}

	domainQueryCounts := make(map[string]int, len(rows))
	for _, r := range rows {
		domainQueryCounts[r.Domain] = r.Count
	}

	mostQueriedDomain := ""
	maxCount := 0
	for domain, count := range domainQueryCounts {
		if count > maxCount {
			maxCount = count
			mostQueriedDomain = domain
		}
	}

	return crd, mostQueriedDomain, domainQueryCounts, nil
}

func GetTopBlockedDomains(db *gorm.DB, blockedRequests int) ([]map[string]interface{}, error) {
	var rows []struct {
		Domain string `gorm:"column:domain"`
		Hits   int    `gorm:"column:hits"`
	}

	if err := db.Table("request_logs").
		Select("domain, COUNT(*) as hits").
		Where("blocked = ?", true).
		Group("domain").
		Order("hits DESC").
		Limit(5).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	var topBlockedDomains []map[string]interface{}
	for _, r := range rows {
		freq := 0
		if blockedRequests > 0 {
			freq = (r.Hits * 100) / blockedRequests
		}
		topBlockedDomains = append(topBlockedDomains, map[string]interface{}{
			"name":      r.Domain,
			"hits":      r.Hits,
			"frequency": freq,
		})
	}
	return topBlockedDomains, nil
}

func GetTopClients(db *gorm.DB) ([]map[string]interface{}, error) {
	var total int64
	if err := db.Table("request_logs").Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []struct {
		ClientIP     string  `gorm:"column:client_ip"`
		RequestCount int     `gorm:"column:request_count"`
		Frequency    float32 `gorm:"column:frequency"`
	}

	if err := db.Table("request_logs").
		Select("? as frequency, client_ip, COUNT(*) as request_count", 0).
		Group("client_ip").
		Order("request_count DESC").
		Limit(5).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	clients := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		freq := float32(r.RequestCount) * 100 / float32(total)
		clients = append(clients, map[string]interface{}{
			"client":       r.ClientIP,
			"requestCount": r.RequestCount,
			"frequency":    freq,
		})
	}
	return clients, nil
}

func CountQueries(db *gorm.DB, search string) (int, error) {
	var total int64
	err := db.Table("request_logs").
		Where("domain LIKE ?", "%"+search+"%").
		Count(&total).Error
	return int(total), err
}

func SaveRequestLog(db *gorm.DB, entries []model.RequestLogEntry) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, entry := range entries {
			rl := RequestLog{
				Timestamp:         entry.Timestamp,
				Domain:            entry.Domain,
				Blocked:           entry.Blocked,
				Cached:            entry.Cached,
				ResponseTimeNs:    entry.ResponseTime.Nanoseconds(),
				ClientIP:          entry.ClientInfo.IP,
				ClientName:        entry.ClientInfo.Name,
				Status:            entry.Status,
				QueryType:         entry.QueryType,
				ResponseSizeBytes: entry.ResponseSizeBytes,
				Protocol:          string(entry.Protocol),
			}

			for _, resolvedIP := range entry.IP {
				rl.IPs = append(rl.IPs, RequestLogIP{
					IP:    resolvedIP.IP,
					RType: resolvedIP.RType,
				})
			}

			if err := tx.Create(&rl).Error; err != nil {
				return fmt.Errorf("could not save request log: %v", err)
			}
		}
		return nil
	})
}

type vacuumFunc func()

func DeleteRequestLogsTimebased(vacuum vacuumFunc, db *gorm.DB, requestThreshold, maxRetries int, retryDelay time.Duration) {
	cutoffTime := time.Now().Add(-time.Duration(requestThreshold) * time.Second)

	for retryCount := range maxRetries {
		result := db.Where("timestamp < ?", cutoffTime).Delete(&RequestLog{})
		if result.Error != nil {
			if result.Error.Error() == "database is locked" {
				log.Warning("Database is locked; retrying (%d/%d)", retryCount+1, maxRetries)
				time.Sleep(retryDelay)
				continue
			}
			log.Error("Failed to clear old entries: %s", result.Error)
			break
		}

		if affected := result.RowsAffected; affected > 0 {
			vacuum()
			log.Debug("Cleared %d old entries", affected)
		}
		break
	}
}
