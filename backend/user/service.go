package user

import (
	"errors"
	"goaway/backend/database"
	"goaway/backend/logging"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repository Repository
}

var log = logging.GetLogger()

func NewService(repo Repository) *Service {
	return &Service{repository: repo}
}

func (s *Service) CreateUser(username, password string) error {
	log.Info("Creating a new user with name '%s'", username)

	hashedPassword, err := hashPassword(password)
	if err != nil {
		log.Error("Failed to hash password: %v", err)
		return err
	}

	newUser := &database.User{Username: username, Password: hashedPassword}

	if err := s.repository.Create(newUser); err != nil {
		log.Error("Failed to create user: %v", err)
		return err
	}

	log.Debug("User created successfully")
	return nil
}

func (s *Service) Exists(username string) bool {
	user, err := s.repository.FindByUsername(username)
	if err != nil {
		return false
	}

	if user != nil {
		return true
	}

	return false
}

func (s *Service) Authenticate(username, password string) bool {
	user, err := s.repository.FindByUsername(username)
	if err != nil {
		log.Error("Authentication failed for user '%s': %v", username, err)
		return false
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Debug("Invalid password for user '%s'", username)
		return false
	}

	log.Debug("User '%s' authenticated successfully", username)
	return true
}

func (s *Service) UpdatePassword(username, newPassword string) error {
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		log.Error("Failed to hash new password: %v", err)
		return err
	}

	if err := s.repository.UpdatePassword(username, hashedPassword); err != nil {
		log.Error("Failed to update password: %v", err)
		return err
	}

	log.Debug("Password updated successfully for user '%s'", username)
	return nil
}

func (s *Service) ValidateCredentials(user User) error {
	user.Username = strings.TrimSpace(user.Username)
	user.Password = strings.TrimSpace(user.Password)

	if user.Username == "" || user.Password == "" {
		return errors.New("username and password cannot be empty")
	}

	if len(user.Username) > 60 {
		return errors.New("username too long")
	}
	if len(user.Password) > 120 {
		return errors.New("password too long")
	}

	for _, r := range user.Username {
		if r < 32 || r == 127 {
			return errors.New("username contains invalid characters")
		}
	}

	return nil
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}
