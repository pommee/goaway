package key

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"goaway/backend/database"
	"goaway/backend/logging"
	"sort"
	"sync"
	"time"
)

// Service handles business logic for API keys
type Service struct {
	repository Repository
	cacheTime  time.Time
	cacheTTL   time.Duration

	cacheMu  sync.RWMutex
	keyCache map[string]APIKey
}

var log = logging.GetLogger()

func NewService(repo Repository) *Service {
	return &Service{
		repository: repo,
		keyCache:   make(map[string]APIKey),
		cacheTTL:   1 * time.Hour,
	}
}

func (s *Service) VerifyKey(apiKey string) bool {
	if err := s.refreshCache(); err != nil {
		log.Warning("Failed to refresh API key cache: %v", err)

		count, err := s.repository.CountByKey(apiKey)
		if err != nil {
			log.Warning("Failed to verify API key in database: %v", err)
			return false
		}
		return count > 0
	}

	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	for _, value := range s.keyCache {
		if value.Key == apiKey {
			return true
		}
	}

	return false
}

// CreateKey generates and stores a new API key
func (s *Service) CreateKey(name string) (string, error) {
	apiKey, err := generateKey()
	if err != nil {
		return "", err
	}

	newAPIKey := database.APIKey{
		Name:      name,
		Key:       apiKey,
		CreatedAt: time.Now(),
	}

	if err := s.repository.Create(&newAPIKey); err != nil {
		return "", fmt.Errorf("key with name '%s' already exists", name)
	}

	s.cacheMu.Lock()
	s.keyCache[apiKey] = APIKey{
		Name:      name,
		Key:       apiKey,
		CreatedAt: newAPIKey.CreatedAt,
	}
	s.cacheMu.Unlock()

	log.Info("Created new API key with name: %s", name)

	return apiKey, nil
}

// DeleteKey removes an API key by name
func (s *Service) DeleteKey(keyName string) error {
	if err := s.repository.DeleteByName(keyName); err != nil {
		return err
	}

	s.cacheMu.Lock()
	for key, value := range s.keyCache {
		if value.Name == keyName {
			delete(s.keyCache, key)
			break
		}
	}
	s.cacheMu.Unlock()

	if err := s.refreshCache(); err != nil {
		log.Warning("%v", err)
	}

	return nil
}

// GetAllKeys returns all API keys with redacted key values
func (s *Service) GetAllKeys() ([]APIKey, error) {
	if err := s.refreshCache(); err != nil {
		return nil, err
	}

	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	keys := make([]APIKey, 0, len(s.keyCache))
	for _, apiKey := range s.keyCache {
		keyCopy := apiKey
		keyCopy.Key = "redacted"
		keys = append(keys, keyCopy)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[j].CreatedAt.Before(keys[i].CreatedAt)
	})

	return keys, nil
}

// refreshCache updates the in-memory cache from the database
func (s *Service) refreshCache() error {
	s.cacheMu.RLock()
	if time.Since(s.cacheTime) < s.cacheTTL && len(s.keyCache) > 0 {
		s.cacheMu.RUnlock()
		return nil
	}
	s.cacheMu.RUnlock()

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	// Double-check after acquiring write lock
	if time.Since(s.cacheTime) < s.cacheTTL && len(s.keyCache) > 0 {
		return nil
	}

	apiKeys, err := s.repository.FindAll()
	if err != nil {
		return err
	}

	newCache := make(map[string]APIKey)
	for _, apiKey := range apiKeys {
		newCache[apiKey.Key] = APIKey{
			Name:      apiKey.Name,
			Key:       apiKey.Key,
			CreatedAt: apiKey.CreatedAt,
		}
	}

	s.keyCache = newCache
	s.cacheTime = time.Now()
	return nil
}

// generateKey creates a random hex-encoded API key
func generateKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
