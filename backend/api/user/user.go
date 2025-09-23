package user

import (
	"errors"
	"goaway/backend/dns/database"
	"goaway/backend/logging"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var log = logging.GetLogger()

type Credentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (user *User) Create(db *gorm.DB) error {
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		log.Error("Failed to hash password: %v", err)
		return err
	}

	tx := db.Create(&database.User{
		Username: user.Username,
		Password: hashedPassword,
	})

	return tx.Commit().Error
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

func (user *User) Exists(db *gorm.DB) bool {
	query := database.User{Username: user.Username}
	result := db.First(&query)
	return result.RowsAffected > 0
}

func (user *User) Authenticate(db *gorm.DB) bool {
	var dbUser User
	if err := db.Where("username = ?", user.Username).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error("User not found: %s", user.Username)
			return false
		}
		log.Error("Query error: %v", err)
		return false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		return false
	}
	return true
}

func (user *User) UpdatePassword(db *gorm.DB) error {
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		log.Error("Failed to hash new password: %v", err)
		return err
	}

	tx := db.Save(&database.User{
		Password: hashedPassword,
	})

	return tx.Commit().Error
}

func (c *Credentials) Validate() error {
	c.Username = strings.TrimSpace(c.Username)
	c.Password = strings.TrimSpace(c.Password)

	if c.Username == "" || c.Password == "" {
		return errors.New("username and password cannot be empty")
	}

	if len(c.Username) > 60 {
		return errors.New("username too long")
	}
	if len(c.Password) > 120 {
		return errors.New("password too long")
	}

	for _, r := range c.Username {
		if r < 32 || r == 127 {
			return errors.New("username contains invalid characters")
		}
	}

	return nil
}
