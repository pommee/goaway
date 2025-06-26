package database

import (
	"database/sql"
	"fmt"
	"goaway/backend/api/models"
	dbModel "goaway/backend/dns/database/models"
	model "goaway/backend/dns/server/models"
	"goaway/backend/logging"
	"strconv"
	"strings"
	"time"
)

var log = logging.GetLogger()

func GetClientNameFromRequestLog(db *sql.DB, ip string) string {
	hostname := "unknown"

	rows, err := db.Query("SELECT client_name FROM request_log WHERE client_ip = ? AND client_name != 'unknown' LIMIT 1", ip)
	if err == nil {
		defer func(rows *sql.Rows) {
			_ = rows.Close()
		}(rows)
		if rows.Next() {
			_ = rows.Scan(&hostname)
			hostname = strings.TrimSuffix(hostname, ".")
		}
	}

	return hostname
}

func GetDistinctRequestIP(db *sql.DB) int {
	query := "SELECT COUNT(DISTINCT client_ip) FROM request_log"
	var clientCount int

	err := db.QueryRow(query).Scan(&clientCount)
	if err != nil {
		clientCount = 0
	}

	return clientCount
}

func GetRequestSummaryByInterval(interval int, db *sql.DB) ([]model.RequestLogIntervalSummary, error) {
	minutes := strconv.Itoa(interval * 60)
	query := fmt.Sprintf(`
	SELECT
	(timestamp / %s) * %s AS interval_start,
	SUM(CASE WHEN blocked = 1 THEN 1 ELSE 0 END) AS blocked_count,
	SUM(CASE WHEN cached = 1 THEN 1 ELSE 0 END) AS cached_count,
	SUM(CASE WHEN blocked = 0 AND cached = 0 THEN 1 ELSE 0 END) AS allowed_count
	FROM request_log
	WHERE timestamp >= strftime('%%s', 'now', '-24 hours')
	GROUP BY interval_start
	ORDER BY interval_start;`,
		minutes, minutes,
	)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var summaries []model.RequestLogIntervalSummary
	for rows.Next() {
		var ts int64
		var summary model.RequestLogIntervalSummary
		if err := rows.Scan(&ts, &summary.BlockedCount, &summary.CachedCount, &summary.AllowedCount); err != nil {
			return nil, err
		}
		summary.IntervalStart = time.Unix(ts, 0)
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func GetResponseSizeSummaryByInterval(intervalMinutes int, db *sql.DB) ([]model.ResponseSizeSummary, error) {
	intervalSeconds := int64(intervalMinutes * 60)
	twentyFourHoursAgo := time.Now().Add(-24 * time.Hour).Unix()

	query := `
		SELECT 
			(timestamp / ?) * ? as start,
			SUM(response_size_bytes) as total_response_size_bytes,
			ROUND(AVG(response_size_bytes)) as avg_response_size_bytes,
			MIN(response_size_bytes) as min_response_size_bytes,
			MAX(response_size_bytes) as max_response_size_bytes
		FROM request_log 
		WHERE timestamp >= ? AND response_size_bytes IS NOT NULL
		GROUP BY (timestamp / ?)
		ORDER BY start ASC
	`

	rows, err := db.Query(query, intervalSeconds, intervalSeconds, twentyFourHoursAgo, intervalSeconds)
	if err != nil {
		return nil, fmt.Errorf("failed to query response size summary: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var summaries []model.ResponseSizeSummary
	for rows.Next() {
		var summary model.ResponseSizeSummary
		var startUnix int64

		err := rows.Scan(
			&startUnix,
			&summary.TotalSizeBytes,
			&summary.AvgResponseSizeBytes,
			&summary.MinResponseSizeBytes,
			&summary.MaxResponseSizeBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan response size summary row: %w", err)
		}

		summary.Start = time.Unix(startUnix, 0)
		summaries = append(summaries, summary)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating response size summary rows: %w", err)
	}

	return summaries, nil
}

func GetUniqueQueryTypes(db *sql.DB) ([]interface{}, error) {
	query := "SELECT COUNT(*) AS count, query_type FROM request_log WHERE query_type <> '' GROUP BY query_type ORDER BY count DESC"
	rows, err := db.Query(query)
	if err != nil {
		log.Error("Error: %v", err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
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

func FetchQueries(db *sql.DB, q models.QueryParams) ([]model.RequestLogEntry, error) {
	var query string
	var args []interface{}

	baseQuery := `
        SELECT rl.id, rl.timestamp, rl.domain, rl.blocked, rl.cached, rl.response_time_ns,
               rl.client_ip, rl.client_name, rl.status, rl.query_type, rl.response_size_bytes
        FROM request_log rl`

	if q.Column == "ip" {
		baseQuery += ` LEFT JOIN resolved_ips ri ON rl.id = ri.request_log_id`
	}

	whereClause := ""
	if q.Search != "" {
		whereClause = " WHERE rl.domain LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	orderClause := " ORDER BY rl." + q.Column + " " + q.Direction

	if q.Column == "ip" {
		orderClause = " ORDER BY MAX(ri.ip) " + q.Direction
	}

	limitClause := " LIMIT ? OFFSET ?"
	args = append(args, q.PageSize, q.Offset)

	query = baseQuery + whereClause + orderClause + limitClause

	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	queries := make([]model.RequestLogEntry, 0, q.PageSize)
	var requestLogIDs []int64

	for rows.Next() {
		var query model.RequestLogEntry
		var timestamp int64
		var requestLogID int64

		query.ClientInfo = &model.Client{}

		if err := rows.Scan(
			&requestLogID, &timestamp, &query.Domain, &query.Blocked, &query.Cached,
			&query.ResponseTime, &query.ClientInfo.IP, &query.ClientInfo.Name,
			&query.Status, &query.QueryType, &query.ResponseSizeBytes,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		query.Timestamp = time.Unix(timestamp, 0)
		query.ID = requestLogID
		queries = append(queries, query)
		requestLogIDs = append(requestLogIDs, requestLogID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	if len(requestLogIDs) > 0 {
		ipMap, err := fetchResolvedIPs(db, requestLogIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch resolved IPs: %w", err)
		}

		for i := range queries {
			if ips, exists := ipMap[queries[i].ID]; exists {
				queries[i].IP = ips
			}
		}
	}

	return queries, nil
}

func fetchResolvedIPs(db *sql.DB, requestLogIDs []int64) (map[int64][]model.ResolvedIP, error) {
	if len(requestLogIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(requestLogIDs))
	args := make([]interface{}, len(requestLogIDs))
	for i, id := range requestLogIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT request_log_id, ip, rtype 
		FROM request_log_ips 
		WHERE request_log_id IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	ipMap := make(map[int64][]model.ResolvedIP)
	for rows.Next() {
		var requestLogID int64
		var resolvedIP model.ResolvedIP

		if err := rows.Scan(&requestLogID, &resolvedIP.IP, &resolvedIP.RType); err != nil {
			return nil, err
		}

		ipMap[requestLogID] = append(ipMap[requestLogID], resolvedIP)
	}

	return ipMap, rows.Err()
}

func FetchAllClients(db *sql.DB) (map[string]dbModel.Client, error) {
	uniqueClients := make(map[string]dbModel.Client)

	rows, err := db.Query(`
		SELECT r.client_ip, r.client_name, r.timestamp, m.mac, m.vendor
		FROM request_log r
		LEFT JOIN mac_addresses m ON r.client_ip = m.ip
	`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var ip, name string
		var dbTimestamp int
		var mac, vendor sql.NullString

		if err := rows.Scan(&ip, &name, &dbTimestamp, &mac, &vendor); err != nil {
			return nil, err
		}

		macStr := ""
		vendorStr := ""
		if mac.Valid {
			macStr = mac.String
		}
		if vendor.Valid {
			vendorStr = vendor.String
		}

		timestamp := time.Unix(int64(dbTimestamp), 0)
		if existing, exists := uniqueClients[ip]; !exists || timestamp.After(existing.LastSeen) {
			uniqueClients[ip] = dbModel.Client{
				Name:     name,
				LastSeen: timestamp,
				Mac:      macStr,
				Vendor:   vendorStr,
			}
		}
	}

	return uniqueClients, nil
}

func GetClientRequestDetails(db *sql.DB, clientIP string) (dbModel.ClientRequestDetails, error) {
	query := `
		SELECT
			COUNT(*),
			COUNT(DISTINCT domain),
			SUM(CASE WHEN blocked THEN 1 ELSE 0 END),
			SUM(CASE WHEN cached THEN 1 ELSE 0 END),
			AVG(response_time_ns) / 1e6,
			MAX(timestamp)
		FROM request_log
		WHERE client_ip LIKE ?`
	searchPattern := "%" + clientIP + "%"

	crd := dbModel.ClientRequestDetails{}
	err := db.QueryRow(query, searchPattern).Scan(
		&crd.TotalRequests, &crd.UniqueDomains, &crd.BlockedRequests,
		&crd.CachedRequests, &crd.AvgResponseTimeMs, &crd.LastSeen)
	if err != nil {
		return crd, err
	}

	return crd, nil
}

func GetMostQueriedDomainByIP(db *sql.DB, clientIP string) (string, error) {
	query := `
		SELECT domain FROM request_log
		WHERE client_ip LIKE ?
		GROUP BY domain
		ORDER BY COUNT(*) DESC
		LIMIT 1`
	searchPattern := "%" + clientIP + "%"

	mostQueriedDomain := ""
	err := db.QueryRow(query, searchPattern).Scan(&mostQueriedDomain)
	if err != nil {
		return "", err
	}

	return mostQueriedDomain, nil
}

func GetAllQueriedDomainsByIP(db *sql.DB, clientIP string) (map[string]int, error) {
	query := `
		SELECT domain, COUNT(*) as query_count
		FROM request_log
		WHERE client_ip LIKE ?
		GROUP BY domain
		ORDER BY query_count DESC`
	searchPattern := "%" + clientIP + "%"

	rows, err := db.Query(query, searchPattern)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	domainQueryCounts := make(map[string]int)
	for rows.Next() {
		var domain string
		var count int
		if err := rows.Scan(&domain, &count); err != nil {
			return nil, err
		}
		domainQueryCounts[domain] = count
	}

	return domainQueryCounts, nil
}

func GetTopBlockedDomains(db *sql.DB, blockedRequests int) ([]map[string]interface{}, error) {
	query := `
	SELECT domain, COUNT(*) as hits
	FROM request_log
	WHERE blocked = 1
	GROUP BY domain
	ORDER BY hits DESC
	LIMIT 5
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var topBlockedDomains []map[string]interface{}
	for rows.Next() {
		var (
			domain string
			hits   int
			freq   int
		)
		if err := rows.Scan(&domain, &hits); err != nil {
			return nil, err
		}
		if blockedRequests > 0 {
			freq = (hits * 100) / blockedRequests
		}
		topBlockedDomains = append(topBlockedDomains, map[string]any{
			"name":      domain,
			"hits":      hits,
			"frequency": freq,
		})
	}

	return topBlockedDomains, nil
}

func GetTopClients(db *sql.DB) ([]map[string]interface{}, error) {
	query := `
	SELECT client_ip, COUNT(*) AS request_count,
	(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM request_log)) AS frequency
	FROM request_log
	GROUP BY client_ip
	ORDER BY request_count DESC
	LIMIT 5;
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var clients []map[string]interface{}
	for rows.Next() {
		var clientIp string
		var requestCount int
		var frequency float32

		if err := rows.Scan(&clientIp, &requestCount, &frequency); err != nil {
			return nil, err
		}

		clients = append(clients, map[string]interface{}{
			"client":       clientIp,
			"requestCount": requestCount,
			"frequency":    frequency,
		})
	}

	return clients, nil
}

func CountQueries(db *sql.DB, search string) (int, error) {
	query := "SELECT COUNT(*) FROM request_log WHERE domain LIKE ?"
	var total int
	err := db.QueryRow(query, "%"+search+"%").Scan(&total)
	return total, err
}

func SaveRequestLog(db *sql.DB, entries []model.RequestLogEntry) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			log.Warning("Could not save request log, performing rollback. Error: %s", p)
		} else if err != nil {
			_ = tx.Rollback()
		} else if commitErr := tx.Commit(); commitErr != nil {
			log.Warning("Commit error: %v", commitErr)
		}
	}()

	for _, entry := range entries {
		var logID int64
		err = tx.QueryRow(`
            INSERT INTO request_log (timestamp, domain, blocked, cached, response_time_ns, 
                                   client_ip, client_name, status, query_type, response_size_bytes) 
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
            RETURNING id`,
			entry.Timestamp.Unix(),
			entry.Domain,
			entry.Blocked,
			entry.Cached,
			entry.ResponseTime,
			entry.ClientInfo.IP,
			entry.ClientInfo.Name,
			entry.Status,
			entry.QueryType,
			entry.ResponseSizeBytes,
		).Scan(&logID)

		if err != nil {
			return fmt.Errorf("could not save request log: %v", err)
		}

		for _, resolvedIP := range entry.IP {
			_, err = tx.Exec(`
                INSERT INTO request_log_ips (request_log_id, ip, rtype) 
                VALUES ($1, $2, $3)`,
				logID, resolvedIP.IP, resolvedIP.RType)

			if err != nil {
				return fmt.Errorf("could not save resolved IP: %v", err)
			}
		}
	}

	return nil
}

type vacuumFunc func()

func DeleteRequestLogsTimebased(vacuum vacuumFunc, db *sql.DB, requestThreshold, maxRetries int, retryDelay time.Duration) {
	query := fmt.Sprintf("DELETE FROM request_log WHERE strftime('%%s', 'now') - timestamp > %d", requestThreshold)

	for retryCount := range maxRetries {
		result, err := db.Exec(query)
		if err != nil {
			if err.Error() == "database is locked" {
				log.Warning("Database is locked; retrying (%d/%d)", retryCount+1, maxRetries)
				time.Sleep(retryDelay)
				continue
			}
			log.Error("Failed to clear old entries: %s", err)
			break
		}

		if affected, err := result.RowsAffected(); err == nil && affected > 0 {
			vacuum()
			log.Debug("Cleared %d old entries", affected)
		}
		break
	}
}
