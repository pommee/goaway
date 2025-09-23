package user

import (
	"context"
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
	log.Info("Creating a new user with name '%s'", user.Username)

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		log.Error("Failed to hash password: %v", err)
		return err
	}

	user.Password = hashedPassword

	result := db.Create(user)
	if result.Error != nil {
		log.Error("Failed to create user: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		log.Error("User creation failed: no rows affected")
		return errors.New("user creation failed: no rows affected")
	}

	log.Debug("User created successfully")
	return nil
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

func (user *User) Exists(db *gorm.DB) bool {
	query := database.User{}
	db.Where("username = ?", user.Username).Find(&query)
	return query.Username != ""
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

	affected, err := gorm.G[database.User](db).Where("username = ?", user.Username).Update(context.Background(), "password", hashedPassword)
	if err != nil {
		log.Error("Failed to update password: %v", err)
		return err
	}

	if affected == 0 {
		log.Error("Password update failed: no rows affected")
		return errors.New("password update failed: no rows affected")
	}

	log.Debug("Password updated successfully")
	return nil
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
