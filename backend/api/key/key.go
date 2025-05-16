package key

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
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
	db        *sql.DB
	keyCache  map[string]ApiKey
	cacheMu   sync.RWMutex
	cacheTime time.Time
	cacheTTL  time.Duration
}

var log = logging.GetLogger()

func NewApiKeyManager(db *sql.DB) *ApiKeyManager {
	return &ApiKeyManager{
		db:       db,
		keyCache: make(map[string]ApiKey),
		cacheTTL: 1 * time.Hour,
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

	rows, err := m.db.Query(`SELECT name, key, created_at FROM apikey`)
	if err != nil {
		return err
	}
	defer rows.Close()

	newCache := make(map[string]ApiKey)
	for rows.Next() {
		var name string
		var key string
		var createdAt time.Time
		if err := rows.Scan(&name, &key, &createdAt); err != nil {
			return err
		}
		newCache[key] = ApiKey{Name: name, Key: key, CreatedAt: createdAt}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	m.keyCache = newCache
	m.cacheTime = time.Now()
	return nil
}

func (m *ApiKeyManager) VerifyApiKey(apiKey string) bool {
	if err := m.refreshCache(); err != nil {
		log.Warning("Failed to refresh API key cache: %v", err)

		var exists bool
		err := m.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM apikey WHERE key = ?)`, apiKey).Scan(&exists)
		if err != nil {
			log.Warning("Failed to verify API key in database: %v", err)
			return false
		}
		return exists
	}

	m.cacheMu.RLock()
	defer m.cacheMu.RUnlock()
	_, exists := m.keyCache[apiKey]
	return exists
}

func generateApiKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (m *ApiKeyManager) CreateApiKey(name string) (string, error) {
	apiKey, err := generateApiKey()
	if err != nil {
		return "", err
	}

	createdAt := time.Now()
	_, err = m.db.Exec(`INSERT INTO apikey (name, key, created_at) VALUES (?, ?, ?)`, name, apiKey, createdAt)
	if err != nil {
		return "", fmt.Errorf("key with name '%s' already exists", name)
	}

	m.cacheMu.Lock()
	m.keyCache[apiKey] = ApiKey{Name: name, Key: apiKey, CreatedAt: createdAt}
	m.cacheMu.Unlock()

	return apiKey, nil
}

func (m *ApiKeyManager) DeleteApiKey(apiKey string) error {
	_, err := m.db.Exec(`DELETE FROM apikey WHERE key = ?`, apiKey)
	if err != nil {
		return err
	}

	m.cacheMu.Lock()
	delete(m.keyCache, apiKey)
	m.cacheMu.Unlock()

	return nil
}

func (m *ApiKeyManager) GetAllApiKeys() ([]ApiKey, error) {
	if err := m.refreshCache(); err != nil {
		return nil, err
	}

	m.cacheMu.RLock()
	defer m.cacheMu.RUnlock()

	keys := make([]ApiKey, 0, len(m.keyCache))
	for _, apiKey := range m.keyCache {
		keys = append(keys, apiKey)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[j].CreatedAt.Before(keys[i].CreatedAt)
	})

	return keys, nil
}
