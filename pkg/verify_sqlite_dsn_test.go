//go:build test
// +build test

package pkg

import (
	"strings"
	"testing"
)

// TestSQLiteDSNFormat verifies the SQLite DSN format with examples
func TestSQLiteDSNFormat(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected []string // Expected substrings in DSN
	}{
		{
			name: "SQLite with basic path",
			config: DatabaseConfig{
				Driver:   "sqlite3",
				Database: "test.db",
				Options:  make(map[string]string),
			},
			expected: []string{
				"test.db",
				"_journal_mode=WAL",
				"_foreign_keys=ON",
				"_busy_timeout=5000",
			},
		},
		{
			name: "SQLite with file path",
			config: DatabaseConfig{
				Driver:   "sqlite",
				Database: "./data/app.sqlite",
				Options:  make(map[string]string),
			},
			expected: []string{
				"./data/app.sqlite",
				"_journal_mode=WAL",
				"_foreign_keys=ON",
				"_busy_timeout=5000",
			},
		},
		{
			name: "SQLite with custom options",
			config: DatabaseConfig{
				Driver:   "sqlite3",
				Database: "custom.db",
				Options: map[string]string{
					"_cache_size":  "2000",
					"_synchronous": "NORMAL",
				},
			},
			expected: []string{
				"custom.db",
				"_journal_mode=WAL",
				"_foreign_keys=ON",
				"_busy_timeout=5000",
				"_cache_size=2000",
				"_synchronous=NORMAL",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dm := &databaseManager{}
			dsn, err := dm.buildDSN(tt.config)

			if err != nil {
				t.Fatalf("buildDSN() error = %v", err)
			}

			// Verify all expected substrings are present
			for _, expected := range tt.expected {
				if !strings.Contains(dsn, expected) {
					t.Errorf("DSN missing expected substring '%s'. Got DSN: %s", expected, dsn)
				}
			}

			// Verify DSN format (should have ? separator)
			if !strings.Contains(dsn, "?") {
				t.Errorf("DSN should contain '?' separator. Got: %s", dsn)
			}

			t.Logf("Generated DSN: %s", dsn)
		})
	}
}
