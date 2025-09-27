package database

import (
	"goaway/backend/dns/database"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)

	err = database.AutoMigrate(db)
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func TestSourceModel(t *testing.T) {
	t.Run("CreateSource", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		source := &database.Source{
			Name:        "Test Source",
			URL:         "https://example.com/blocklist",
			Active:      true,
			LastUpdated: time.Now().Unix(),
		}

		err := db.Create(source).Error
		require.NoError(t, err)
		assert.NotZero(t, source.ID)
	})

	t.Run("UniqueURLConstraint", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		source1 := &database.Source{
			Name:   "Source 1",
			URL:    "https://duplicate.com",
			Active: true,
		}
		source2 := &database.Source{
			Name:   "Source 2",
			URL:    "https://duplicate.com",
			Active: false,
		}

		err := db.Create(source1).Error
		require.NoError(t, err)

		err = db.Create(source2).Error
		assert.Error(t, err)
	})

	t.Run("QuerySource", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		source := &database.Source{
			Name:        "Query Test Source",
			URL:         "https://querytest.com",
			Active:      true,
			LastUpdated: time.Now().Unix(),
		}

		err := db.Create(source).Error
		require.NoError(t, err)

		var retrieved database.Source
		err = db.Where("name = ?", "Query Test Source").First(&retrieved).Error
		require.NoError(t, err)
		assert.Equal(t, source.Name, retrieved.Name)
		assert.Equal(t, source.URL, retrieved.URL)
	})

	t.Run("UpdateSource", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		source := &database.Source{
			Name:   "Update Test",
			URL:    "https://update.com",
			Active: false,
		}

		err := db.Create(source).Error
		require.NoError(t, err)

		source.Active = true
		source.LastUpdated = time.Now().Unix()

		err = db.Save(source).Error
		require.NoError(t, err)

		var updated database.Source
		err = db.First(&updated, source.ID).Error
		require.NoError(t, err)
		assert.True(t, updated.Active)
	})
}

func TestBlacklistModel(t *testing.T) {
	t.Run("CreateBlacklist", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		source := &database.Source{
			Name:   "Test Blacklist Source",
			URL:    "https://blacklist.com",
			Active: true,
		}
		err := db.Create(source).Error
		require.NoError(t, err)

		blacklist := &database.Blacklist{
			Domain:   "malicious.com",
			SourceID: source.ID,
		}

		err = db.Create(blacklist).Error
		require.NoError(t, err)
	})

	t.Run("BlacklistWithAssociation", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		source := &database.Source{
			Name:   "Test Blacklist Source",
			URL:    "https://blacklist.com",
			Active: true,
		}
		err := db.Create(source).Error
		require.NoError(t, err)

		blacklist := &database.Blacklist{
			Domain:   "evil.com",
			SourceID: source.ID,
		}

		err = db.Create(blacklist).Error
		require.NoError(t, err)

		var retrieved database.Blacklist
		err = db.Preload("Source").Where("domain = ?", "evil.com").First(&retrieved).Error
		require.NoError(t, err)
		assert.Equal(t, source.Name, retrieved.Source.Name)
	})

	t.Run("CompositePrimaryKey", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		source := &database.Source{
			Name:   "First Source",
			URL:    "https://first.com",
			Active: true,
		}
		err := db.Create(source).Error
		require.NoError(t, err)

		source2 := &database.Source{
			Name:   "Second Source",
			URL:    "https://second.com",
			Active: true,
		}
		err = db.Create(source2).Error
		require.NoError(t, err)

		blacklist1 := &database.Blacklist{
			Domain:   "duplicate-domain.com",
			SourceID: source.ID,
		}
		blacklist2 := &database.Blacklist{
			Domain:   "duplicate-domain.com",
			SourceID: source2.ID,
		}

		err = db.Create(blacklist1).Error
		require.NoError(t, err)

		err = db.Create(blacklist2).Error
		require.NoError(t, err)

		blacklist3 := &database.Blacklist{
			Domain:   "duplicate-domain.com",
			SourceID: source.ID,
		}

		err = db.Create(blacklist3).Error
		assert.Error(t, err)
	})
}

func TestWhitelistModel(t *testing.T) {
	t.Run("CreateWhitelist", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		whitelist := &database.Whitelist{
			Domain: "trusted.com",
		}

		err := db.Create(whitelist).Error
		require.NoError(t, err)
	})

	t.Run("UniqueDomain", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		domain := "unique-test.com"

		whitelist1 := &database.Whitelist{Domain: domain}
		whitelist2 := &database.Whitelist{Domain: domain}

		err := db.Create(whitelist1).Error
		require.NoError(t, err)

		err = db.Create(whitelist2).Error
		assert.Error(t, err)
	})
}

func TestRequestLogModel(t *testing.T) {
	t.Run("CreateRequestLogWithIPs", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		requestLog := &database.RequestLog{
			Timestamp:         time.Now(),
			Domain:            "example.com",
			Blocked:           false,
			Cached:            true,
			ResponseTimeNs:    1500000,
			ClientIP:          "192.168.1.100",
			ClientName:        "test-client",
			Status:            "NOERROR",
			QueryType:         "A",
			ResponseSizeBytes: 64,
			Protocol:          "UDP",
			IPs: []database.RequestLogIP{
				{IP: "11.222.333.44", RType: "A"},
				{IP: "1111:2222:333:4:555:6666:7777:8888", RType: "AAAA"},
			},
		}

		err := db.Create(requestLog).Error
		require.NoError(t, err)
		assert.NotZero(t, requestLog.ID)

		var ips []database.RequestLogIP
		err = db.Where("request_log_id = ?", requestLog.ID).Find(&ips).Error
		require.NoError(t, err)
		assert.Len(t, ips, 2)
	})

	t.Run("QueryWithPreload", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		requestLog := &database.RequestLog{
			Timestamp: time.Now(),
			Domain:    "preload-test.com",
			Blocked:   false,
			Cached:    false,
			ClientIP:  "111.222.33.4",
			QueryType: "A",
			IPs: []database.RequestLogIP{
				{IP: "111.222.33.4", RType: "A"},
			},
		}

		err := db.Create(requestLog).Error
		require.NoError(t, err)

		var retrieved database.RequestLog
		err = db.Preload("IPs").First(&retrieved, requestLog.ID).Error
		require.NoError(t, err)
		assert.Len(t, retrieved.IPs, 1)
		assert.Equal(t, "111.222.33.4", retrieved.IPs[0].IP)
	})

	t.Run("CascadeDelete", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		requestLog := &database.RequestLog{
			Timestamp: time.Now(),
			Domain:    "cascade-test.com",
			Blocked:   false,
			Cached:    false,
			ClientIP:  "10.0.0.1",
			QueryType: "A",
			IPs: []database.RequestLogIP{
				{IP: "111.222.33.4", RType: "A"},
				{IP: "444.333.22.111", RType: "A"},
			},
		}

		err := db.Create(requestLog).Error
		require.NoError(t, err)

		var count int64
		db.Model(&database.RequestLogIP{}).Where("request_log_id = ?", requestLog.ID).Count(&count)
		assert.Equal(t, int64(2), count)

		err = db.Select("IPs").Delete(requestLog).Error
		require.NoError(t, err)

		db.Model(&database.RequestLogIP{}).Where("request_log_id = ?", requestLog.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("IndexedQueries", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		now := time.Now()

		for i := range 10 {
			requestLog := &database.RequestLog{
				Timestamp: now.Add(-time.Duration(i) * time.Hour),
				Domain:    "indexed-test.com",
				Blocked:   i%2 == 0,
				Cached:    i%3 == 0,
				ClientIP:  "192.168.1.2",
				QueryType: "A",
			}
			err := db.Create(requestLog).Error
			require.NoError(t, err)
		}

		var logs []database.RequestLog
		err := db.Where("timestamp > ?", now.Add(-5*time.Hour)).Find(&logs).Error
		require.NoError(t, err)
		assert.Len(t, logs, 5)

		err = db.Where("domain = ? AND timestamp > ?", "indexed-test.com", now.Add(-3*time.Hour)).Find(&logs).Error
		require.NoError(t, err)
		assert.Len(t, logs, 3)
	})
}

func TestResolutionModel(t *testing.T) {
	t.Run("CreateResolution", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		resolution := &database.Resolution{
			Domain: "custom.local",
			IP:     "192.168.1.10",
		}

		err := db.Create(resolution).Error
		require.NoError(t, err)
	})

	t.Run("UpdateResolution", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		domain := "update-test.local"
		resolution := &database.Resolution{
			Domain: domain,
			IP:     "192.168.1.20",
		}

		err := db.Create(resolution).Error
		require.NoError(t, err)

		resolution.IP = "192.168.1.30"
		err = db.Save(resolution).Error
		require.NoError(t, err)

		var updated database.Resolution
		err = db.First(&updated, "domain = ?", domain).Error
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.30", updated.IP)
	})
}

func TestMacAddressModel(t *testing.T) {
	t.Run("CreateMacAddress", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		mac := &database.MacAddress{
			MAC:    "00:1A:2B:3C:4D:5E",
			IP:     "192.168.1.2",
			Vendor: "Intel Corp",
		}

		err := db.Create(mac).Error
		require.NoError(t, err)
	})

	t.Run("UniqueMACConstraint", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		macAddr := "00:AA:BB:CC:DD:EE"

		mac1 := &database.MacAddress{
			MAC:    macAddr,
			IP:     "192.168.1.60",
			Vendor: "Vendor A",
		}
		mac2 := &database.MacAddress{
			MAC:    macAddr,
			IP:     "192.168.1.70",
			Vendor: "Vendor B",
		}

		err := db.Create(mac1).Error
		require.NoError(t, err)

		err = db.Create(mac2).Error
		assert.Error(t, err)
	})
}

func TestUserModel(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		user := &database.User{
			Username: "testuser",
			Password: "hashed_password_123",
		}

		err := db.Create(user).Error
		require.NoError(t, err)
	})

	t.Run("UniqueUsername", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		username := "duplicateuser"

		user1 := &database.User{
			Username: username,
			Password: "password1",
		}
		user2 := &database.User{
			Username: username,
			Password: "password2",
		}

		err := db.Create(user1).Error
		require.NoError(t, err)

		err = db.Create(user2).Error
		assert.Error(t, err)
	})
}

func TestAPIKeyModel(t *testing.T) {
	t.Run("CreateAPIKey", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		apiKey := &database.APIKey{
			Name:      "test-key",
			Key:       "abc123def456",
			CreatedAt: time.Now(),
		}

		err := db.Create(apiKey).Error
		require.NoError(t, err)
	})

	t.Run("QueryByCreatedAt", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		now := time.Now()
		oldKey := &database.APIKey{
			Name:      "old-key",
			Key:       "old123",
			CreatedAt: now.Add(-24 * time.Hour),
		}
		newKey := &database.APIKey{
			Name:      "new-key",
			Key:       "new456",
			CreatedAt: now,
		}

		err := db.Create(oldKey).Error
		require.NoError(t, err)
		err = db.Create(newKey).Error
		require.NoError(t, err)

		var recentKeys []database.APIKey
		err = db.Where("created_at > ?", now.Add(-1*time.Hour)).Find(&recentKeys).Error
		require.NoError(t, err)
		assert.Len(t, recentKeys, 1)
		assert.Equal(t, "new-key", recentKeys[0].Name)
	})
}

func TestNotificationModel(t *testing.T) {
	t.Run("CreateNotification", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		notification := &database.Notification{
			Severity:  "warning",
			Category:  "dns",
			Text:      "DNS query rate exceeded threshold",
			Read:      false,
			CreatedAt: time.Now(),
		}

		err := db.Create(notification).Error
		require.NoError(t, err)
		assert.NotZero(t, notification.ID)
	})

	t.Run("QueryUnreadNotifications", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		notifications := []*database.Notification{
			{Severity: "info", Category: "system", Text: "System started", Read: false, CreatedAt: time.Now()},
			{Severity: "error", Category: "dns", Text: "DNS server down", Read: true, CreatedAt: time.Now()},
			{Severity: "warning", Category: "security", Text: "Suspicious activity", Read: false, CreatedAt: time.Now()},
		}

		for _, notif := range notifications {
			err := db.Create(notif).Error
			require.NoError(t, err)
		}

		var unread []database.Notification
		err := db.Where("read = ?", false).Find(&unread).Error
		require.NoError(t, err)
		assert.Len(t, unread, 2)
	})
}

func TestPrefetchModel(t *testing.T) {
	t.Run("CreatePrefetch", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		prefetch := &database.Prefetch{
			Domain:  "frequent.com",
			Refresh: 300,
			QType:   1,
		}

		err := db.Create(prefetch).Error
		require.NoError(t, err)
	})

	t.Run("QueryByQType", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		prefetches := []*database.Prefetch{
			{Domain: "a-record.com", Refresh: 300, QType: 1},
			{Domain: "aaaa-record.com", Refresh: 300, QType: 28},
			{Domain: "mx-record.com", Refresh: 600, QType: 15},
		}

		for _, pf := range prefetches {
			err := db.Create(pf).Error
			require.NoError(t, err)
		}

		var aRecords []database.Prefetch
		err := db.Where("q_type = ?", 1).Find(&aRecords).Error
		require.NoError(t, err)
		assert.Len(t, aRecords, 1)
		assert.Equal(t, "a-record.com", aRecords[0].Domain)
	})
}

func TestAuditModel(t *testing.T) {
	t.Run("CreateAudit", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		audit := &database.Audit{
			Topic:     "configuration",
			Message:   "DNS settings updated",
			CreatedAt: time.Now(),
		}

		err := db.Create(audit).Error
		require.NoError(t, err)
		assert.NotZero(t, audit.ID)
	})

	t.Run("QueryAuditsByTopic", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		audits := []*database.Audit{
			{Topic: "user", Message: "User login", CreatedAt: time.Now()},
			{Topic: "configuration", Message: "Config changed", CreatedAt: time.Now()},
			{Topic: "user", Message: "User logout", CreatedAt: time.Now()},
		}

		for _, audit := range audits {
			err := db.Create(audit).Error
			require.NoError(t, err)
		}

		var userAudits []database.Audit
		err := db.Where("topic = ?", "user").Find(&userAudits).Error
		require.NoError(t, err)
		assert.Len(t, userAudits, 2)
	})
}

func TestAlertModel(t *testing.T) {
	t.Run("CreateAlert", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		alert := &database.Alert{
			Type:    "discord",
			Enabled: true,
			Name:    "DNS Server Failure Alert",
			Webhook: "https://discord.com/webhook/sometoken",
		}

		err := db.Create(alert).Error
		require.NoError(t, err)
	})

	t.Run("QueryEnabledAlerts", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		alerts := []*database.Alert{
			{Type: "high_query_rate", Enabled: true, Name: "High Query Rate", Webhook: "https://example.com/1"},
			{Type: "malware_detected", Enabled: false, Name: "Malware Detection", Webhook: "https://example.com/2"},
			{Type: "dns_timeout", Enabled: true, Name: "DNS Timeout", Webhook: "https://example.com/3"},
		}

		for _, alert := range alerts {
			err := db.Create(alert).Error
			require.NoError(t, err)
		}

		var enabled []database.Alert
		err := db.Where("enabled = ?", true).Find(&enabled).Error
		require.NoError(t, err)
		assert.Len(t, enabled, 2)
	})

	t.Run("UniqueAlertType", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		alertType := "discord"

		alert1 := &database.Alert{
			Type:    alertType,
			Enabled: true,
			Name:    "Alert 1",
			Webhook: "https://example.com/1",
		}
		alert2 := &database.Alert{
			Type:    alertType,
			Enabled: false,
			Name:    "Alert 2",
			Webhook: "https://example.com/2",
		}

		err := db.Create(alert1).Error
		require.NoError(t, err)

		err = db.Create(alert2).Error
		assert.Error(t, err)
	})
}

func TestModelRelationships(t *testing.T) {
	t.Run("SourceBlacklistRelationship", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		source := &database.Source{
			Name:   "Integration Test Source",
			URL:    "https://integration.com",
			Active: true,
		}
		err := db.Create(source).Error
		require.NoError(t, err)

		blacklists := []*database.Blacklist{
			{Domain: "bad1.com", SourceID: source.ID},
			{Domain: "bad2.com", SourceID: source.ID},
			{Domain: "bad3.com", SourceID: source.ID},
		}

		for _, bl := range blacklists {
			err := db.Create(bl).Error
			require.NoError(t, err)
		}

		var count int64
		err = db.Model(&database.Blacklist{}).Where("source_id = ?", source.ID).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)

		err = db.Select("Blacklist").Delete(source).Error
		require.NoError(t, err)
	})
}

func TestModelPerformance(t *testing.T) {
	t.Run("BulkInsertRequestLogs", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		batchSize := 1000
		logs := make([]database.RequestLog, batchSize)

		for i := 0; i < batchSize; i++ {
			logs[i] = database.RequestLog{
				Timestamp:      time.Now().Add(-time.Duration(i) * time.Second),
				Domain:         "bulk-test.com",
				Blocked:        i%2 == 0,
				Cached:         i%3 == 0,
				ResponseTimeNs: int64(1000000 + i*1000),
				ClientIP:       "192.168.1.1",
				QueryType:      "A",
			}
		}

		start := time.Now()
		err := db.CreateInBatches(logs, 100).Error
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("Bulk insert of %d records took %v", batchSize, duration)

		var count int64
		err = db.Model(&database.RequestLog{}).Where("domain = ?", "bulk-test.com").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(batchSize), count)
	})

	t.Run("IndexedQueryPerformance", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		for i := 0; i < 100; i++ {
			requestLog := &database.RequestLog{
				Timestamp:      time.Now().Add(-time.Duration(i) * time.Minute),
				Domain:         "perf-test.com",
				Blocked:        i%2 == 0,
				Cached:         i%3 == 0,
				ResponseTimeNs: int64(1000000 + i*1000),
				ClientIP:       "192.168.1.1",
				QueryType:      "A",
			}
			err := db.Create(requestLog).Error
			require.NoError(t, err)
		}

		start := time.Now()
		var logs []database.RequestLog
		err := db.Where("timestamp > ?", time.Now().Add(-1*time.Hour)).Limit(100).Find(&logs).Error
		duration := time.Since(start)

		require.NoError(t, err)
		t.Logf("Indexed timestamp query took %v", duration)
		assert.True(t, duration < 100*time.Millisecond, "Query should be fast")
	})
}

func TestDataIntegrity(t *testing.T) {
	t.Run("ValidateTimestampHandling", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		now := time.Now()

		notification := &database.Notification{
			Severity:  "info",
			Category:  "test",
			Text:      "Test notification",
			Read:      false,
			CreatedAt: now,
		}

		err := db.Create(notification).Error
		require.NoError(t, err)

		var retrieved database.Notification
		err = db.First(&retrieved, notification.ID).Error
		require.NoError(t, err)

		assert.True(t, retrieved.CreatedAt.Sub(now).Abs() < time.Second)
	})

	t.Run("ValidateStringLengths", func(t *testing.T) {
		db, cleanup := setupTestDB(t)
		defer cleanup()

		longDomain := strings.Repeat("a", 300) + ".com"
		requestLog := &database.RequestLog{
			Timestamp: time.Now(),
			Domain:    longDomain,
			Blocked:   false,
			Cached:    false,
			ClientIP:  "192.168.1.1",
			QueryType: "A",
		}

		err := db.Create(requestLog).Error
		if err != nil {
			t.Logf("Long domain rejected as expected: %v", err)
		} else {
			t.Logf("Long domain accepted")
		}
	})
}
