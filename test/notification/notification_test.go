package notification

import (
	"goaway/backend/database"
	notificationPkg "goaway/backend/notification"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupNotificationTestDB(t *testing.T) (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		TranslateError: true,
	})
	require.NoError(t, err)

	err = database.AutoMigrate(db)
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	return db, cleanup
}

func TestSendNotification(t *testing.T) {
	t.Run("SendValidNotification", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		service.SendNotification(
			notificationPkg.SeverityWarning,
			notificationPkg.CategoryDNS,
			"Test DNS warning",
		)

		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)

		assert.Len(t, result.Notifications, 1)
		assert.Equal(t, "warning", result.Notifications[0].Severity)
		assert.Equal(t, "dns", result.Notifications[0].Category)
		assert.Equal(t, "Test DNS warning", result.Notifications[0].Text)
		assert.False(t, result.Notifications[0].Read)
	})

	t.Run("SendMultipleNotifications", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Info 1")
		service.SendNotification(notificationPkg.SeverityWarning, notificationPkg.CategoryDNS, "Warning 1")
		service.SendNotification(notificationPkg.SeverityError, notificationPkg.CategoryAPI, "Error 1")

		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)

		assert.Len(t, result.Notifications, 3)
	})

	t.Run("NotificationHasDefaultValues", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Test")

		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)

		assert.Len(t, result.Notifications, 1)
		assert.False(t, result.Notifications[0].Read)
		assert.NotZero(t, result.Notifications[0].CreatedAt)
	})
}

func TestGetNotifications(t *testing.T) {
	t.Run("GetNotificationsFirstPage", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		// Create 15 notifications using service
		for i := 1; i <= 15; i++ {
			service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Notification ")
		}

		result, err := service.GetNotifications(1, 10)
		require.NoError(t, err)

		assert.NotNil(t, result)
		assert.Len(t, result.Notifications, 10)
		assert.Equal(t, int64(15), result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.Limit)
		assert.Equal(t, 2, result.TotalPages)
	})

	t.Run("GetNotificationsSecondPage", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		// Create 25 notifications using service
		for i := 1; i <= 25; i++ {
			service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Notification ")
		}

		result, err := service.GetNotifications(2, 10)
		require.NoError(t, err)

		assert.Len(t, result.Notifications, 10)
		assert.Equal(t, int64(25), result.Total)
		assert.Equal(t, 2, result.Page)
		assert.Equal(t, 3, result.TotalPages)
	})

	t.Run("GetNotificationsLastPage", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		// Create 25 notifications using service
		for i := 1; i <= 25; i++ {
			service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Notification ")
		}

		result, err := service.GetNotifications(3, 10)
		require.NoError(t, err)

		assert.Len(t, result.Notifications, 5)
		assert.Equal(t, int64(25), result.Total)
		assert.Equal(t, 3, result.Page)
		assert.Equal(t, 3, result.TotalPages)
	})

	t.Run("GetNotificationsOnlyUnread", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		// Create 5 unread and 5 read notifications
		for i := 1; i <= 5; i++ {
			service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Unread notification")
		}
		// Mark 5 as read
		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)
		var readIds []int
		for _, notif := range result.Notifications {
			readIds = append(readIds, int(notif.ID))
		}
		service.MarkNotificationsAsRead(readIds)

		// Create 5 more unread notifications
		for i := 1; i <= 5; i++ {
			service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Unread notification")
		}

		result, err = service.GetNotifications(1, 50)
		require.NoError(t, err)

		assert.Len(t, result.Notifications, 5)
		assert.Equal(t, int64(5), result.Total)

		// Verify all are unread
		for _, notif := range result.Notifications {
			assert.False(t, notif.Read)
		}
	})

	t.Run("GetNotificationsOrderedByNewest", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		// Create notifications using service
		for i := 1; i <= 3; i++ {
			service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Notification ")
			time.Sleep(10 * time.Millisecond)
		}

		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)

		assert.Len(t, result.Notifications, 3)
		// Most recent should be first
		assert.Greater(t, result.Notifications[0].CreatedAt, result.Notifications[2].CreatedAt)
	})

	t.Run("GetNotificationsEmptyDatabase", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)

		assert.Len(t, result.Notifications, 0)
		assert.Equal(t, int64(0), result.Total)
		assert.Equal(t, 1, result.TotalPages)
	})

	t.Run("InvalidPageNumberDefaultsToOne", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Test")

		result, err := service.GetNotifications(0, 10)
		require.NoError(t, err)

		assert.Equal(t, 1, result.Page)
		assert.Len(t, result.Notifications, 1)
	})

	t.Run("InvalidLimitDefaultsToFifty", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		// Create 60 notifications to test the limit using service
		for i := 1; i <= 60; i++ {
			service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Notification ")
		}

		result, err := service.GetNotifications(1, 0)
		require.NoError(t, err)

		assert.Equal(t, 50, result.Limit)
		assert.Len(t, result.Notifications, 50)
	})

	t.Run("CalculatesTotalPagesCorrectly", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		// Create 37 notifications using service
		for i := 1; i <= 37; i++ {
			service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Notification ")
		}

		result, err := service.GetNotifications(1, 10)
		require.NoError(t, err)

		// 37 items with limit 10 = 4 pages (10 + 10 + 10 + 7)
		assert.Equal(t, 4, result.TotalPages)
	})
}

func TestMarkNotificationsAsRead(t *testing.T) {
	t.Run("MarkSingleNotificationAsRead", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Test")

		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)
		require.Len(t, result.Notifications, 1)

		notifID := int(result.Notifications[0].ID)
		err = service.MarkNotificationsAsRead([]int{notifID})
		require.NoError(t, err)

		result, err = service.GetNotifications(1, 50)
		require.NoError(t, err)
		// Should return no unread notifications
		assert.Len(t, result.Notifications, 0)
	})

	t.Run("MarkMultipleNotificationsAsRead", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		// Create 5 notifications
		for i := 1; i <= 5; i++ {
			service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Test ")
		}

		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)
		require.Len(t, result.Notifications, 5)

		var ids []int
		for _, notif := range result.Notifications {
			ids = append(ids, int(notif.ID))
		}

		err = service.MarkNotificationsAsRead(ids)
		require.NoError(t, err)

		result, err = service.GetNotifications(1, 50)
		require.NoError(t, err)
		// Should return no unread notifications
		assert.Len(t, result.Notifications, 0)
	})

	t.Run("MarkWithEmptyIdsList", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Test")

		// Try to mark empty list
		err := service.MarkNotificationsAsRead([]int{})
		require.NoError(t, err)

		// Notification should still be unread
		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)
		assert.Len(t, result.Notifications, 1)
		assert.False(t, result.Notifications[0].Read)
	})

	t.Run("MarkAlreadyReadNotifications", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Test")

		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)

		notifID := int(result.Notifications[0].ID)
		// Mark as read
		err = service.MarkNotificationsAsRead([]int{notifID})
		require.NoError(t, err)

		// Mark again as read (idempotent)
		err = service.MarkNotificationsAsRead([]int{notifID})
		require.NoError(t, err)

		// Should still be no unread notifications
		result, err = service.GetNotifications(1, 50)
		require.NoError(t, err)
		assert.Len(t, result.Notifications, 0)
	})

	t.Run("MarkNonexistentNotification", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		err := service.MarkNotificationsAsRead([]int{9999})
		require.NoError(t, err)

		// Should not error, just no notifications to mark
		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)
		assert.Len(t, result.Notifications, 0)
	})

	t.Run("MarkPartialListSingleAndMultiple", func(t *testing.T) {
		db, cleanup := setupNotificationTestDB(t)
		defer cleanup()

		repo := notificationPkg.NewRepository(db)
		service := notificationPkg.NewService(repo)

		// Create 5 notifications
		for i := 1; i <= 5; i++ {
			service.SendNotification(notificationPkg.SeverityInfo, notificationPkg.CategoryServer, "Notification ")
		}

		// Mark first 3 as read
		result, err := service.GetNotifications(1, 50)
		require.NoError(t, err)
		require.Len(t, result.Notifications, 5)

		var readIds []int
		for i := range 3 {
			readIds = append(readIds, int(result.Notifications[i].ID))
		}

		err = service.MarkNotificationsAsRead(readIds)
		require.NoError(t, err)

		// Should have 2 unread notifications remaining
		result, err = service.GetNotifications(1, 50)
		require.NoError(t, err)
		assert.Len(t, result.Notifications, 2)
	})
}
