package notification

import (
	"goaway/backend/database"

	"gorm.io/gorm"
)

type Repository interface {
	CreateNotification(newNotification *database.Notification) error
	GetNotifications() ([]database.Notification, error)
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

func (r *repository) GetNotifications() ([]database.Notification, error) {
	var notifications []database.Notification

	result := r.db.Where("read = ?", false).Find(&notifications)
	if result.Error != nil {
		return nil, result.Error
	}

	return notifications, nil
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
