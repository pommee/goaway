package notification

import (
	"goaway/backend/database"

	"gorm.io/gorm"
)

type Repository interface {
	CreateNotification(newNotification *database.Notification) error
	GetNotifications(page, limit int) (int64, []database.Notification, error)
	MarkNotificationsAsRead(notificationIDs []int) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateNotification(newNotification *database.Notification) error {
	tx := r.db.Create(&newNotification)
	return tx.Error
}

func (r *repository) GetNotifications(page, limit int) (int64, []database.Notification, error) {
	var notifications []database.Notification
	var total int64

	if err := r.db.Model(&database.Notification{}).Where("read = ?", false).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	offset := (page - 1) * limit
	result := r.db.
		Where("read = ?", false).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications)

	if result.Error != nil {
		return 0, nil, result.Error
	}

	return total, notifications, nil
}

func (r *repository) MarkNotificationsAsRead(notificationIDs []int) error {
	if len(notificationIDs) == 0 {
		return nil
	}

	result := r.db.Model(&database.Notification{}).
		Where("id IN ?", notificationIDs).
		Update("read", true)

	if result.Error != nil {
		return result.Error
	}

	return nil
}
