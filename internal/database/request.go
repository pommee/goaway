package database

import (
	"database/sql"
	"fmt"
	"goaway/internal/api/models"
	dbModel "goaway/internal/database/models"
	"goaway/internal/logging"
	model "goaway/internal/server/models"
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

func GetRequestSummaryByInterval(db *sql.DB) ([]model.RequestLogIntervalSummary, error) {
	query := `
SELECT
  (timestamp / 120) * 120 AS interval_start,
  SUM(CASE WHEN blocked = 1 THEN 1 ELSE 0 END) AS blocked_count,
  SUM(CASE WHEN cached = 1 THEN 1 ELSE 0 END) AS cached_count,
  SUM(CASE WHEN blocked = 0 AND cached = 0 THEN 1 ELSE 0 END) AS allowed_count
FROM request_log
WHERE timestamp >= strftime('%s', 'now', '-24 hours')
GROUP BY interval_start
ORDER BY interval_start;
`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
	query := fmt.Sprintf(`
		SELECT timestamp, domain, ip, blocked, cached, response_time_ns, client_ip, client_name, status, query_type, response_size_bytes
		FROM request_log
		WHERE domain LIKE ?
		ORDER BY %s %s
		LIMIT ? OFFSET ?`, q.Column, q.Direction)

	rows, err := db.Query(query, "%"+q.Search+"%", q.PageSize, q.Offset)
	if err != nil {
		log.Error("Database query error: %v", err)
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var queries []model.RequestLogEntry
	for rows.Next() {
		var query model.RequestLogEntry
		var ipString string
		query.ClientInfo = &model.Client{}

		if err := rows.Scan(
			&query.Timestamp, &query.Domain, &ipString,
			&query.Blocked, &query.Cached, &query.ResponseTime,
			&query.ClientInfo.IP, &query.ClientInfo.Name, &query.Status, &query.QueryType, &query.ResponseSizeBytes,
		); err != nil {
			return nil, err
		}

		query.IP = strings.Split(ipString, ",")
		queries = append(queries, query)
	}

	return queries, nil
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
		var timestamp time.Time
		var mac, vendor sql.NullString

		if err := rows.Scan(&ip, &name, &timestamp, &mac, &vendor); err != nil {
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

func SaveRequestLog(db *sql.DB, entries []model.RequestLogEntry) {
	tx, err := db.Begin()
	if err != nil {
		log.Error("Could not start database transaction %v", err)
		return
	}
	defer func() {
		if err := tx.Commit(); err != nil {
			log.Warning("DB commit error %v", err)
		}
	}()

	query := "INSERT INTO request_log (timestamp, domain, ip, blocked, cached, response_time_ns, client_ip, client_name, status, query_type, response_size_bytes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := tx.Prepare(query)

	if err != nil {
		log.Error("Could not create a prepared statement for request logs, reason: %v", err)
		return
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)

	for _, entry := range entries {
		if _, err := stmt.Exec(
			entry.Timestamp.Unix(),
			entry.Domain,
			strings.Join(entry.IP, ","),
			entry.Blocked,
			entry.Cached,
			entry.ResponseTime,
			entry.ClientInfo.IP,
			entry.ClientInfo.Name,
			entry.Status,
			entry.QueryType,
			entry.ResponseSizeBytes,
		); err != nil {
			log.Error("Could not save request log. Reason: %v", err)
			return
		}
	}
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

func CountAllowedAndBlockedRequest(db *sql.DB) (int, int, error) {
	var blockedCount, allowedCount int

	err := db.QueryRow("SELECT COUNT(*) FROM request_log WHERE blocked = 1").Scan(&blockedCount)
	if err != nil {
		log.Error("Failed to get blocked requests count: %s", err)
		return 0, 0, err
	}

	err = db.QueryRow("SELECT COUNT(*) FROM request_log WHERE blocked = 0").Scan(&allowedCount)
	if err != nil {
		log.Error("Failed to get allowed requests count: %s", err)
		return 0, 0, err
	}

	return blockedCount, allowedCount, nil
}
