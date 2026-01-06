package request

import (
	"context"
	"database/sql"
	"fmt"
	"goaway/backend/api/models"
	"goaway/backend/database"
	model "goaway/backend/dns/server/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	SaveRequestLog(entries []model.RequestLogEntry) error

	GetClientName(ip string) string
	GetDistinctRequestIP() int
	GetRequestSummaryByInterval(interval int) ([]model.RequestLogIntervalSummary, error)
	GetResponseSizeSummaryByInterval(intervalMinutes int) ([]model.ResponseSizeSummary, error)
	GetUniqueQueryTypes() ([]models.QueryTypeCount, error)
	FetchQueries(q models.QueryParams) ([]model.RequestLogEntry, error)
	FetchClient(ip string) (*model.Client, error)
	FetchAllClients() (map[string]model.Client, error)
	GetClientDetailsWithDomains(clientIP string) (ClientRequestDetails, string, map[string]int, error)
	GetClientHistory(clientIP string) ([]models.DomainHistory, error)
	GetTopBlockedDomains(blockedRequests int) ([]map[string]interface{}, error)
	GetTopClients() ([]map[string]interface{}, error)
	CountQueries(search string) (int, error)

	UpdateClientName(ip string, name string) error
	UpdateClientBypass(ip string, bypass bool) error

	DeleteRequestLogsTimebased(vacuum vacuumFunc, requestThreshold, maxRetries int, retryDelay time.Duration) error
}

type ClientRequestDetails struct {
	LastSeen          string
	MostQueriedDomain string
	TotalRequests     int
	UniqueDomains     int
	BlockedRequests   int
	CachedRequests    int
	AvgResponseTimeMs float64
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) SaveRequestLog(entries []model.RequestLogEntry) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, entry := range entries {
			rl := database.RequestLog{
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
				rl.IPs = append(rl.IPs, database.RequestLogIP{
					IP:         resolvedIP.IP,
					RecordType: resolvedIP.RType,
				})
			}

			if err := tx.Create(&rl).Error; err != nil {
				return fmt.Errorf("could not save request log: %w", err)
			}
		}
		return nil
	})
}

func (r *repository) GetClientName(ip string) string {
	var hostname string
	err := r.db.Model(&database.RequestLog{}).
		Select("client_name").
		Where("client_ip = ? AND client_name != ?", ip, "unknown").
		Limit(1).
		Scan(&hostname).Error

	if err != nil || hostname == "" {
		return "unknown"
	}

	return strings.TrimSuffix(hostname, ".")
}

func (r *repository) GetDistinctRequestIP() int {
	var count int64

	err := r.db.Model(&database.RequestLog{}).
		Select("COUNT(DISTINCT client_ip)").
		Scan(&count).Error
	if err != nil {
		return 0
	}

	return int(count)
}

func (r *repository) GetRequestSummaryByInterval(interval int) ([]model.RequestLogIntervalSummary, error) {
	minutes := interval * 60

	type tempSummary struct {
		IntervalStartUnix int64 `gorm:"column:interval_start_unix"`
		BlockedCount      int   `gorm:"column:blocked_count"`
		CachedCount       int   `gorm:"column:cached_count"`
		AllowedCount      int   `gorm:"column:allowed_count"`
	}

	var rawSummaries []tempSummary

	query := fmt.Sprintf(`
		SELECT
			((CAST(STRFTIME('%%s', timestamp) AS INTEGER) / %d) * %d) AS interval_start_unix,
			SUM(blocked) AS blocked_count,
			SUM(cached) AS cached_count,
			SUM(NOT blocked AND NOT cached) AS allowed_count
		FROM request_logs
		WHERE timestamp >= DATETIME('now', '-1 day')
		GROUP BY (CAST(STRFTIME('%%s', timestamp) AS INTEGER) / %d)
		ORDER BY interval_start_unix ASC
	`, minutes, minutes, minutes)

	err := r.db.Raw(query).Scan(&rawSummaries).Error
	if err != nil {
		return nil, err
	}

	summaries := make([]model.RequestLogIntervalSummary, len(rawSummaries))
	for i := range rawSummaries {
		summaries[i] = model.RequestLogIntervalSummary{
			IntervalStart: time.Unix(rawSummaries[i].IntervalStartUnix, 0).Format("2006-01-02 15:04:05"),
			BlockedCount:  rawSummaries[i].BlockedCount,
			CachedCount:   rawSummaries[i].CachedCount,
			AllowedCount:  rawSummaries[i].AllowedCount,
		}
	}

	return summaries, nil
}

func (r *repository) GetResponseSizeSummaryByInterval(intervalMinutes int) ([]model.ResponseSizeSummary, error) {
	intervalSeconds := int64(intervalMinutes * 60)
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour)

	var summaries []model.ResponseSizeSummary

	query := `
		WITH logs_unix AS (
			SELECT
				CAST(strftime('%s', timestamp) AS INTEGER) AS ts_unix,
				response_size_bytes
			FROM request_logs
			WHERE timestamp >= ? AND response_size_bytes IS NOT NULL
		)
		SELECT
			(ts_unix / ?) * ? AS start_unix,
			SUM(response_size_bytes) AS total_size_bytes,
			ROUND(AVG(response_size_bytes)) AS avg_response_size_bytes,
			MIN(response_size_bytes) AS min_response_size_bytes,
			MAX(response_size_bytes) AS max_response_size_bytes
		FROM logs_unix
		GROUP BY ts_unix / ?
		ORDER BY start_unix ASC
	`

	err := r.db.Raw(query,
		twentyFourHoursAgo,
		intervalSeconds,
		intervalSeconds,
		intervalSeconds,
	).Scan(&summaries).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query response size summary: %w", err)
	}

	for i := range summaries {
		summaries[i].Start = time.Unix(summaries[i].StartUnix, 0)
	}

	return summaries, nil
}

func (r *repository) GetUniqueQueryTypes() ([]models.QueryTypeCount, error) {
	stats := make([]models.QueryTypeCount, 0, 5)
	if err := r.db.Model(&database.RequestLog{}).
		Select("query_type, COUNT(*) as count").
		Where("query_type != ?", "").
		Group("query_type").
		Order("count DESC").
		Find(&stats).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *repository) FetchQueries(q models.QueryParams) ([]model.RequestLogEntry, error) {
	var logs []database.RequestLog
	query := r.db.Model(&database.RequestLog{})

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
			ID:                log.ID,
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
			results[i].IP[j] = model.ResolvedIP{IP: ip.IP, RType: ip.RecordType}
		}
	}

	return results, nil
}

func (r *repository) FetchClient(ip string) (*model.Client, error) {
	var row struct {
		ClientIP   string         `gorm:"column:client_ip"`
		ClientName string         `gorm:"column:client_name"`
		Timestamp  time.Time      `gorm:"column:timestamp"`
		Mac        sql.NullString `gorm:"column:mac"`
		Vendor     sql.NullString `gorm:"column:vendor"`
		Bypass     sql.NullBool   `gorm:"column:bypass"`
	}

	subquery := r.db.Table("request_logs").
		Select("MAX(timestamp)").
		Where("client_ip = ?", ip)

	if err := r.db.Table("request_logs r").
		Select("r.client_ip, r.client_name, r.timestamp, m.mac, m.vendor, m.bypass").
		Joins("LEFT JOIN mac_addresses m ON r.client_ip = m.ip").
		Where("r.client_ip = ?", ip).
		Where("r.timestamp = (?)", subquery).
		Scan(&row).Error; err != nil {
		return nil, err
	}

	if row.ClientIP == "" {
		return nil, fmt.Errorf("client with ip '%s' was not found", ip)
	}

	client := &model.Client{
		IP:       row.ClientIP,
		Name:     row.ClientName,
		LastSeen: row.Timestamp,
		Mac:      row.Mac.String,
		Vendor:   row.Vendor.String,
		Bypass:   row.Bypass.Bool,
	}

	return client, nil
}

func (r *repository) FetchAllClients() (map[string]model.Client, error) {
	var rows []struct {
		ClientIP   string         `gorm:"column:client_ip"`
		ClientName string         `gorm:"column:client_name"`
		Timestamp  time.Time      `gorm:"column:timestamp"`
		Mac        sql.NullString `gorm:"column:mac"`
		Vendor     sql.NullString `gorm:"column:vendor"`
		Bypass     sql.NullBool   `gorm:"column:bypass"`
	}

	subquery := r.db.Table("request_logs").
		Select("client_ip, MAX(timestamp) as max_timestamp").
		Group("client_ip")

	if err := r.db.Table("request_logs r").
		Select("r.client_ip, r.client_name, r.timestamp, m.mac, m.vendor, m.bypass").
		Joins("INNER JOIN (?) latest ON r.client_ip = latest.client_ip AND r.timestamp = latest.max_timestamp", subquery).
		Joins("LEFT JOIN mac_addresses m ON r.client_ip = m.ip").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	uniqueClients := make(map[string]model.Client, len(rows))
	for _, row := range rows {
		uniqueClients[row.ClientIP] = model.Client{
			IP:       row.ClientIP,
			Name:     row.ClientName,
			LastSeen: row.Timestamp,
			Mac:      row.Mac.String,
			Vendor:   row.Vendor.String,
			Bypass:   row.Bypass.Bool,
		}
	}

	return uniqueClients, nil
}

func (r *repository) GetClientDetailsWithDomains(clientIP string) (ClientRequestDetails, string, map[string]int, error) {
	var crd ClientRequestDetails
	err := r.db.Table("request_logs").
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

	err = r.db.Table("request_logs").
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

func (r *repository) GetClientHistory(clientIP string) ([]models.DomainHistory, error) {
	var history []models.DomainHistory

	err := r.db.Table("request_logs").
		Select("domain, timestamp").
		Where("client_ip = ?", clientIP).
		Order("timestamp DESC").
		Limit(1000).
		Scan(&history).Error

	if err != nil {
		return nil, err
	}

	return history, nil
}

func (r *repository) GetTopBlockedDomains(blockedRequests int) ([]map[string]interface{}, error) {
	var rows []struct {
		Domain string `gorm:"column:domain"`
		Hits   int    `gorm:"column:hits"`
	}

	if err := r.db.Table("request_logs").
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

func (r *repository) GetTopClients() ([]map[string]interface{}, error) {
	var total int64
	if err := r.db.Table("request_logs").Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []struct {
		ClientIP     string  `gorm:"column:client_ip"`
		ClientName   string  `gorm:"column:client_name"`
		RequestCount int     `gorm:"column:request_count"`
		Frequency    float32 `gorm:"column:frequency"`
	}

	if err := r.db.Table("request_logs").
		Select("? as frequency, client_ip, client_name, COUNT(*) as request_count", 0).
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
			"clientName":   r.ClientName,
			"requestCount": r.RequestCount,
			"frequency":    freq,
		})
	}
	return clients, nil
}

func (r *repository) CountQueries(search string) (int, error) {
	var total int64
	err := r.db.Table("request_logs").
		Where("domain LIKE ?", "%"+search+"%").
		Count(&total).Error
	return int(total), err
}

func (r *repository) UpdateClientName(ip, name string) error {
	err := r.db.Model(&database.RequestLog{}).
		Where("client_ip = ?", ip).
		Update("client_name", name).Error

	if err != nil {
		return fmt.Errorf("failed whiled updating client name: %w", err)
	}

	return nil
}

func (r *repository) UpdateClientBypass(ip string, bypass bool) error {
	err := r.db.Model(&database.MacAddress{}).
		Where("ip = ?", ip).
		Updates(map[string]any{
			"bypass":     bypass,
			"updated_at": time.Now(),
		}).Error

	if err != nil {
		return fmt.Errorf("failed to update client bypass: %w", err)
	}

	return nil
}

func (r *repository) DeleteRequestLogsTimebased(vacuum vacuumFunc, requestThreshold, maxRetries int, retryDelay time.Duration) error {
	cutoffTime := time.Now().Add(-time.Duration(requestThreshold) * time.Second)

	for retryCount := range maxRetries {
		result := r.db.Where("timestamp < ?", cutoffTime).Delete(&database.RequestLog{})
		if result.Error != nil {
			if result.Error.Error() == "database is locked" {
				log.Warning("Database is locked; retrying (%d/%d)", retryCount+1, maxRetries)
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("failed to clear old entries: %w", result.Error)
		}

		if affected := result.RowsAffected; affected > 0 {
			vacuum(context.Background())
			log.Debug("Cleared %d old entries", affected)
		}
		return nil // Success
	}

	return fmt.Errorf("failed to delete after %d retries", maxRetries)
}
