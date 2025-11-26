package pkg

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// pluginStorageImpl provides isolated storage for a plugin using database backend
type pluginStorageImpl struct {
	pluginName string
	database   DatabaseManager
	mu         sync.RWMutex
	cache      map[string]interface{} // In-memory cache for performance
}

// NewPluginStorage creates a new plugin storage instance with database backend
func NewPluginStorage(pluginName string, database DatabaseManager) PluginStorage {
	return &pluginStorageImpl{
		pluginName: pluginName,
		database:   database,
		cache:      make(map[string]interface{}),
	}
}

// Get retrieves a value from plugin storage
// First checks in-memory cache, then falls back to database
func (s *pluginStorageImpl) Get(key string) (interface{}, error) {
	if key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	s.mu.RLock()
	// Check cache first
	if value, exists := s.cache[key]; exists {
		s.mu.RUnlock()
		return value, nil
	}
	s.mu.RUnlock()

	// If not in cache, load from database
	if s.database != nil {
		value, err := s.loadFromDatabase(key)
		if err != nil {
			return nil, err
		}

		// Update cache
		s.mu.Lock()
		s.cache[key] = value
		s.mu.Unlock()

		return value, nil
	}

	return nil, fmt.Errorf("key not found: %s", key)
}

// Set stores a value in plugin storage
// Stores in both in-memory cache and database for persistence
func (s *pluginStorageImpl) Set(key string, value interface{}) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Update cache
	s.cache[key] = value

	// Persist to database if available
	if s.database != nil {
		return s.saveToDatabase(key, value)
	}

	return nil
}

// Delete removes a value from plugin storage
// Removes from both in-memory cache and database
func (s *pluginStorageImpl) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from cache
	delete(s.cache, key)

	// Remove from database if available
	if s.database != nil {
		return s.deleteFromDatabase(key)
	}

	return nil
}

// List returns all keys in plugin storage
// Returns keys from in-memory cache (which should be synced with database)
func (s *pluginStorageImpl) List() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// If database is available, load all keys from database to ensure completeness
	if s.database != nil {
		keys, err := s.listFromDatabase()
		if err != nil {
			return nil, err
		}
		return keys, nil
	}

	// Otherwise return from cache
	keys := make([]string, 0, len(s.cache))
	for key := range s.cache {
		keys = append(keys, key)
	}

	return keys, nil
}

// Clear removes all values from plugin storage
// Clears both in-memory cache and database entries
func (s *pluginStorageImpl) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear cache
	s.cache = make(map[string]interface{})

	// Clear database if available
	if s.database != nil {
		return s.clearDatabase()
	}

	return nil
}

// Database operations

// loadFromDatabase retrieves a value from the database
func (s *pluginStorageImpl) loadFromDatabase(key string) (interface{}, error) {
	query, err := s.database.GetQuery("load_plugin_storage")
	if err != nil {
		return nil, fmt.Errorf("failed to load load_plugin_storage query: %w", err)
	}

	var valueJSON string
	err = s.database.QueryRow(query, s.pluginName, key).Scan(&valueJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("key not found: %s", key)
		}
		return nil, fmt.Errorf("failed to load from database: %w", err)
	}

	// Deserialize JSON
	var value interface{}
	if err := json.Unmarshal([]byte(valueJSON), &value); err != nil {
		return nil, fmt.Errorf("failed to deserialize value: %w", err)
	}

	return value, nil
}

// saveToDatabase persists a value to the database
func (s *pluginStorageImpl) saveToDatabase(key string, value interface{}) error {
	// Serialize value to JSON
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	query, err := s.database.GetQuery("save_plugin_storage")
	if err != nil {
		return fmt.Errorf("failed to load save_plugin_storage query: %w", err)
	}

	now := time.Now()
	_, err = s.database.Exec(query, s.pluginName, key, string(valueJSON), now, now)
	if err != nil {
		return fmt.Errorf("failed to save to database: %w", err)
	}

	return nil
}

// deleteFromDatabase removes a value from the database
func (s *pluginStorageImpl) deleteFromDatabase(key string) error {
	query, err := s.database.GetQuery("delete_plugin_storage")
	if err != nil {
		return fmt.Errorf("failed to load delete_plugin_storage query: %w", err)
	}

	_, err = s.database.Exec(query, s.pluginName, key)
	if err != nil {
		return fmt.Errorf("failed to delete from database: %w", err)
	}

	return nil
}

// listFromDatabase retrieves all keys for this plugin from the database
func (s *pluginStorageImpl) listFromDatabase() ([]string, error) {
	query, err := s.database.GetQuery("list_plugin_storage_keys")
	if err != nil {
		return nil, fmt.Errorf("failed to load list_plugin_storage_keys query: %w", err)
	}

	rows, err := s.database.Query(query, s.pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to list from database: %w", err)
	}
	defer rows.Close()

	keys := make([]string, 0)
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, fmt.Errorf("failed to scan key: %w", err)
		}
		keys = append(keys, key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return keys, nil
}

// clearDatabase removes all entries for this plugin from the database
func (s *pluginStorageImpl) clearDatabase() error {
	query, err := s.database.GetQuery("clear_plugin_storage")
	if err != nil {
		return fmt.Errorf("failed to load clear_plugin_storage query: %w", err)
	}

	_, err = s.database.Exec(query, s.pluginName)
	if err != nil {
		return fmt.Errorf("failed to clear database: %w", err)
	}

	return nil
}
