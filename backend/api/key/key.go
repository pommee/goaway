package key

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"goaway/backend/dns/database"
	"goaway/backend/logging"
	"sort"
	"sync"
	"time"
)

type ApiKey struct {
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"createdAt"`
}

type ApiKeyManager struct {
	dbManager *database.DatabaseManager
	keyCache  map[string]ApiKey
	cacheMu   sync.RWMutex
	cacheTime time.Time
	cacheTTL  time.Duration
}

var log = logging.GetLogger()

func NewApiKeyManager(dbManager *database.DatabaseManager) *ApiKeyManager {
	return &ApiKeyManager{
		dbManager: dbManager,
		keyCache:  make(map[string]ApiKey),
		cacheTTL:  1 * time.Hour,
	}
}

func (m *ApiKeyManager) refreshCache() error {
	m.cacheMu.RLock()
	if time.Since(m.cacheTime) < m.cacheTTL && len(m.keyCache) > 0 {
		m.cacheMu.RUnlock()
		return nil
	}
	m.cacheMu.RUnlock()

	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()

	if time.Since(m.cacheTime) < m.cacheTTL && len(m.keyCache) > 0 {
		return nil
	}

	var apiKeys []database.APIKey
	result := m.dbManager.Conn.Find(&apiKeys)
	if result.Error != nil {
		return result.Error
	}

	newCache := make(map[string]ApiKey)
	for _, apiKey := range apiKeys {
		newCache[apiKey.Key] = ApiKey{
			Name:      apiKey.Name,
			Key:       apiKey.Key,
			CreatedAt: apiKey.CreatedAt,
		}
	}

	m.keyCache = newCache
	m.cacheTime = time.Now()
	return nil
}

func (m *ApiKeyManager) VerifyKey(apiKey string) bool {
	if err := m.refreshCache(); err != nil {
		log.Warning("Failed to refresh API key cache: %v", err)

		var count int64
		result := m.dbManager.Conn.Model(&database.APIKey{}).Where("key = ?", apiKey).Count(&count)
		if result.Error != nil {
			log.Warning("Failed to verify API key in database: %v", result.Error)
			return false
		}
		return count > 0
	}

	m.cacheMu.RLock()
	defer m.cacheMu.RUnlock()
	for _, value := range m.keyCache {
		if value.Key == apiKey {
			return true
		}
	}

	return false
}

func generateKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (m *ApiKeyManager) CreateKey(name string) (string, error) {
	apiKey, err := generateKey()
	if err != nil {
		return "", err
	}

	newAPIKey := database.APIKey{
		Name:      name,
		Key:       apiKey,
		CreatedAt: time.Now(),
	}

	result := m.dbManager.Conn.Create(&newAPIKey)
	if result.Error != nil {
		return "", fmt.Errorf("key with name '%s' already exists", name)
	}

	m.cacheMu.Lock()
	m.keyCache[apiKey] = ApiKey{
		Name:      name,
		Key:       apiKey,
		CreatedAt: newAPIKey.CreatedAt,
	}
	m.cacheMu.Unlock()

	log.Info("Created new API key with name: %s", name)

	return apiKey, nil
}

func (m *ApiKeyManager) DeleteKey(keyName string) error {
	result := m.dbManager.Conn.Where("name = ?", keyName).Delete(&database.APIKey{})
	if result.Error != nil {
		return result.Error
	}

	m.cacheMu.Lock()
	for key, value := range m.keyCache {
		if value.Name == keyName {
			delete(m.keyCache, key)
			break
		}
	}
	m.cacheMu.Unlock()

	if err := m.refreshCache(); err != nil {
		log.Warning("%v", err)
	}

	return nil
}

func (m *ApiKeyManager) GetAllKeys() ([]ApiKey, error) {
	if err := m.refreshCache(); err != nil {
		return nil, err
	}

	m.cacheMu.RLock()
	defer m.cacheMu.RUnlock()

	keys := make([]ApiKey, 0, len(m.keyCache))
	for _, apiKey := range m.keyCache {
		keyCopy := apiKey
		keyCopy.Key = "redacted"
		keys = append(keys, keyCopy)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[j].CreatedAt.Before(keys[i].CreatedAt)
	})

	return keys, nil
}
