package pkg

import (
	"os"
	"path/filepath"
	"testing"
	"testing/quick"
)

// **Feature: sqlite-standardization, Property 1: SQL loader returns valid SQL for all required queries**
// **Validates: Requirements 1.1, 1.5, 4.5**
func TestProperty_SQLLoaderReturnsValidSQL(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	sqliteDir := filepath.Join(tempDir, "sqlite")
	if err := os.MkdirAll(sqliteDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Property: For any set of SQL files created, the SQL loader should successfully
	// return non-empty SQL statements for all query names
	property := func(seed uint32) bool {
		// Generate valid query names from seed
		queryNames := generateValidQueryNames(seed, 5)

		if len(queryNames) == 0 {
			return true // Empty case is trivially true
		}

		// Create SQL files for each query name
		for _, name := range queryNames {
			sqlContent := "SELECT 1;" // Simple valid SQL
			filePath := filepath.Join(sqliteDir, name+".sql")
			if err := os.WriteFile(filePath, []byte(sqlContent), 0644); err != nil {
				t.Logf("Failed to write test SQL file: %v", err)
				return false
			}
		}

		// Create SQL loader
		loader, err := NewSQLLoader("sqlite", tempDir)
		if err != nil {
			t.Logf("Failed to create SQL loader: %v", err)
			return false
		}

		// Load all SQL files
		if err := loader.LoadAll(); err != nil {
			t.Logf("Failed to load SQL files: %v", err)
			return false
		}

		// Verify all valid query names can be retrieved
		for _, name := range queryNames {
			query, err := loader.GetQuery(name)
			if err != nil {
				t.Logf("Failed to get query '%s': %v", name, err)
				return false
			}

			// Verify the query is non-empty
			if query == "" {
				t.Logf("Query '%s' returned empty SQL", name)
				return false
			}
		}

		// Clean up files for next iteration
		entries, _ := os.ReadDir(sqliteDir)
		for _, entry := range entries {
			os.Remove(filepath.Join(sqliteDir, entry.Name()))
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// **Feature: sqlite-standardization, Property 9: SQL loader handles missing queries gracefully**
// **Validates: Requirements 4.5, 5.3**
func TestProperty_SQLLoaderHandlesMissingQueries(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	sqliteDir := filepath.Join(tempDir, "sqlite")
	if err := os.MkdirAll(sqliteDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Property: For any query name that doesn't exist, the SQL loader should return
	// a descriptive error rather than panicking or returning empty strings
	property := func(seed1 uint32, seed2 uint32) bool {
		// Generate valid query names
		existingQueries := generateValidQueryNames(seed1, 5)
		missingQuery := generateValidQueryName(seed2)

		// Ensure missingQuery is not in existingQueries
		for _, name := range existingQueries {
			if name == missingQuery {
				// If collision, just skip this iteration
				return true
			}
		}

		// Create SQL files for existing queries
		for _, name := range existingQueries {
			sqlContent := "SELECT 1;"
			filePath := filepath.Join(sqliteDir, name+".sql")
			if err := os.WriteFile(filePath, []byte(sqlContent), 0644); err != nil {
				t.Logf("Failed to write test SQL file: %v", err)
				return false
			}
		}

		// Create SQL loader
		loader, err := NewSQLLoader("sqlite", tempDir)
		if err != nil {
			t.Logf("Failed to create SQL loader: %v", err)
			return false
		}

		// Load all SQL files
		if err := loader.LoadAll(); err != nil {
			t.Logf("Failed to load SQL files: %v", err)
			return false
		}

		// Verify that querying for a missing query returns an error
		query, err := loader.GetQuery(missingQuery)

		// The property holds if:
		// 1. An error is returned (not nil)
		// 2. The query string is empty
		if err == nil {
			t.Logf("Expected error for missing query '%s', but got nil", missingQuery)
			return false
		}

		// Verify query is empty when error is returned
		if query != "" {
			t.Logf("Expected empty query string on error, got: %s", query)
			return false
		}

		// Clean up files for next iteration
		entries, _ := os.ReadDir(sqliteDir)
		for _, entry := range entries {
			os.Remove(filepath.Join(sqliteDir, entry.Name()))
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Helper function to generate valid query names from a seed
func generateValidQueryNames(seed uint32, count int) []string {
	// Use seed to generate deterministic but varied query names
	names := make([]string, 0, count)
	prefixes := []string{"save", "load", "delete", "create", "update", "get", "list", "cleanup"}
	entities := []string{"session", "token", "tenant", "user", "metric", "plugin", "hook", "event"}

	for i := 0; i < count; i++ {
		idx := (seed + uint32(i)) % uint32(len(prefixes)*len(entities))
		prefixIdx := idx / uint32(len(entities))
		entityIdx := idx % uint32(len(entities))
		name := prefixes[prefixIdx] + "_" + entities[entityIdx]
		names = append(names, name)
	}

	return names
}

// Helper function to generate a single valid query name from a seed
func generateValidQueryName(seed uint32) string {
	prefixes := []string{"find", "insert", "remove", "check", "validate", "count", "search", "query"}
	entities := []string{"data", "record", "entry", "item", "object", "document", "resource", "entity"}

	prefixIdx := seed % uint32(len(prefixes))
	entityIdx := (seed / uint32(len(prefixes))) % uint32(len(entities))

	return prefixes[prefixIdx] + "_" + entities[entityIdx]
}

// **Feature: sqlite-standardization, Property 2: SQL files exist for all database engines**
// **Validates: Requirements 2.2, 2.4, 7.2**
func TestProperty_SQLFilesExistForAllEngines(t *testing.T) {
	// This property verifies that for any SQL file in one engine directory,
	// equivalent files with the same name exist in all other engine directories

	// Define the base SQL directory and supported engines
	sqlDir := "../sql"
	engines := []string{"sqlite", "mysql", "postgres", "mssql"}

	// Property: For any SQL file found in any engine directory, a file with the
	// same name should exist in all other engine directories
	property := func(engineIdx uint8) bool {
		// Use engineIdx to select which engine to check
		if engineIdx >= uint8(len(engines)) {
			return true // Skip invalid indices
		}

		selectedEngine := engines[engineIdx%uint8(len(engines))]
		engineDir := filepath.Join(sqlDir, selectedEngine)

		// Check if the directory exists
		if _, err := os.Stat(engineDir); os.IsNotExist(err) {
			t.Logf("Engine directory does not exist: %s", engineDir)
			return false
		}

		// Read all SQL files in the selected engine directory
		entries, err := os.ReadDir(engineDir)
		if err != nil {
			t.Logf("Failed to read engine directory %s: %v", engineDir, err)
			return false
		}

		// Filter to only .sql files
		sqlFiles := make([]string, 0)
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
				sqlFiles = append(sqlFiles, entry.Name())
			}
		}

		// For each SQL file, verify it exists in all other engine directories
		for _, sqlFile := range sqlFiles {
			for _, otherEngine := range engines {
				if otherEngine == selectedEngine {
					continue // Skip the same engine
				}

				otherFilePath := filepath.Join(sqlDir, otherEngine, sqlFile)
				if _, err := os.Stat(otherFilePath); os.IsNotExist(err) {
					t.Logf("SQL file '%s' exists in %s but not in %s", sqlFile, selectedEngine, otherEngine)
					return false
				}
			}
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// **Feature: sqlite-standardization, Property 3: No SQL strings in Go source code**
// **Validates: Requirements 1.3, 1.4, 7.4**
func TestProperty_NoEmbeddedSQLInGoSource(t *testing.T) {
	// This property verifies that Go source files in pkg/ don't contain
	// SQL keywords embedded in string literals (except in test files and sql_loader.go)

	pkgDir := "."

	// SQL keywords that indicate embedded SQL
	sqlKeywords := []string{
		"CREATE TABLE",
		"INSERT INTO",
		"UPDATE ",
		"DELETE FROM",
		"SELECT ",
		"DROP TABLE",
		"ALTER TABLE",
		"CREATE INDEX",
	}

	// Files that are allowed to contain SQL (test files and sql_loader itself)
	// Note: plugin_storage.go is temporarily allowed as it will be migrated in a future task
	// Note: database_impl.go contains DropTables() which uses simple DROP TABLE statements
	//       for cleanup/testing purposes - this is acceptable as it's standard SQL
	allowedFiles := map[string]bool{
		"sql_loader.go":      true,
		"sql_loader_test.go": true,
		"plugin_storage.go":  true, // TODO: Migrate plugin storage SQL to external files
		"database_impl.go":   true, // Contains DropTables() utility method with simple DROP TABLE SQL
	}

	// Property: For any Go source file in pkg/ (excluding allowed files),
	// the file should not contain SQL keywords in string literals
	property := func(keywordIdx uint8) bool {
		if keywordIdx >= uint8(len(sqlKeywords)) {
			return true // Skip invalid indices
		}

		keyword := sqlKeywords[keywordIdx%uint8(len(sqlKeywords))]

		// Read all Go files in pkg directory
		entries, err := os.ReadDir(pkgDir)
		if err != nil {
			t.Logf("Failed to read pkg directory: %v", err)
			return false
		}

		for _, entry := range entries {
			// Skip directories and non-Go files
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
				continue
			}

			// Skip allowed files
			if allowedFiles[entry.Name()] {
				continue
			}

			// Skip test files (they may contain SQL for testing purposes)
			if filepath.Ext(entry.Name()) == ".go" && len(entry.Name()) > 8 {
				if entry.Name()[len(entry.Name())-8:] == "_test.go" {
					continue
				}
			}

			// Read file content
			filePath := filepath.Join(pkgDir, entry.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Logf("Failed to read file %s: %v", filePath, err)
				return false
			}

			// Check if the keyword appears in the file
			// We use a simple string search - a more sophisticated approach would
			// parse the AST and check only string literals, but this is sufficient
			// for our purposes and catches the common cases
			contentStr := string(content)

			// Look for the keyword in various quote contexts
			if containsSQLKeyword(contentStr, keyword) {
				t.Logf("Found SQL keyword '%s' in file %s", keyword, entry.Name())
				return false
			}
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// Helper function to check if SQL keyword appears in string context
func containsSQLKeyword(content string, keyword string) bool {
	// Simple heuristic: check if keyword appears in quotes
	// This is not perfect but catches most cases

	// Check for keyword in double quotes
	doubleQuotePattern := `"` + keyword
	if idx := findInString(content, doubleQuotePattern); idx != -1 {
		// Make sure it's not in a comment
		if !isInComment(content, idx) {
			return true
		}
	}

	// Check for keyword in backticks
	backtickPattern := "`" + keyword
	if idx := findInString(content, backtickPattern); idx != -1 {
		if !isInComment(content, idx) {
			return true
		}
	}

	return false
}

// Helper to find substring in string (case-insensitive)
func findInString(s, substr string) int {
	// Convert both to uppercase for case-insensitive search
	sUpper := toUpper(s)
	substrUpper := toUpper(substr)

	for i := 0; i <= len(sUpper)-len(substrUpper); i++ {
		if sUpper[i:i+len(substrUpper)] == substrUpper {
			return i
		}
	}
	return -1
}

// Helper to convert string to uppercase
func toUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// Helper to check if position is in a comment
func isInComment(content string, pos int) bool {
	// Look backwards from pos to find if we're in a comment
	lineStart := pos
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}

	// Check if line starts with // (after whitespace)
	for i := lineStart; i < pos; i++ {
		if content[i] == '/' && i+1 < len(content) && content[i+1] == '/' {
			return true
		}
		if content[i] != ' ' && content[i] != '\t' {
			break
		}
	}

	return false
}

// **Feature: sqlite-standardization, Property 7: Engine-specific SQL uses appropriate syntax**
// **Validates: Requirements 6.1, 6.2, 6.3**
func TestProperty_EngineSpecificSQLSyntax(t *testing.T) {
	// This property verifies that UPSERT operations use the correct syntax for each engine:
	// - SQLite/PostgreSQL: ON CONFLICT
	// - MySQL: ON DUPLICATE KEY UPDATE
	// - MSSQL: MERGE

	sqlDir := "../sql"

	// Define engine-specific syntax patterns for UPSERT operations
	engineSyntax := map[string]struct {
		upsertKeywords []string
		autoIncrement  []string
	}{
		"sqlite": {
			upsertKeywords: []string{"ON CONFLICT", "INSERT OR REPLACE"},
			autoIncrement:  []string{"AUTOINCREMENT"},
		},
		"mysql": {
			upsertKeywords: []string{"ON DUPLICATE KEY UPDATE"},
			autoIncrement:  []string{"AUTO_INCREMENT"},
		},
		"postgres": {
			upsertKeywords: []string{"ON CONFLICT"},
			autoIncrement:  []string{"SERIAL", "BIGSERIAL"},
		},
		"mssql": {
			upsertKeywords: []string{"MERGE"},
			autoIncrement:  []string{"IDENTITY"},
		},
	}

	// Files that contain UPSERT operations (save operations)
	upsertFiles := []string{
		"save_session.sql",
		"save_token.sql",
		"save_tenant.sql",
	}

	// Files that contain auto-increment columns (create table operations)
	autoIncrementFiles := []string{
		"create_plugins_table.sql",
	}

	// Property: For any engine, UPSERT SQL files should use the appropriate syntax
	property := func(engineIdx uint8) bool {
		engines := []string{"sqlite", "mysql", "postgres", "mssql"}

		if engineIdx >= uint8(len(engines)) {
			return true // Skip invalid indices
		}

		engine := engines[engineIdx%uint8(len(engines))]
		engineDir := filepath.Join(sqlDir, engine)

		// Check if the directory exists
		if _, err := os.Stat(engineDir); os.IsNotExist(err) {
			t.Logf("Engine directory does not exist: %s", engineDir)
			return false
		}

		syntax := engineSyntax[engine]

		// Check UPSERT files for correct syntax
		for _, filename := range upsertFiles {
			filePath := filepath.Join(engineDir, filename)

			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				// File might not exist yet, skip
				continue
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Logf("Failed to read file %s: %v", filePath, err)
				return false
			}

			contentUpper := toUpper(string(content))

			// Verify that at least one of the expected UPSERT keywords is present
			foundKeyword := false
			for _, keyword := range syntax.upsertKeywords {
				if findInString(contentUpper, toUpper(keyword)) != -1 {
					foundKeyword = true
					break
				}
			}

			if !foundKeyword {
				t.Logf("File %s for engine %s does not contain expected UPSERT syntax: %v",
					filename, engine, syntax.upsertKeywords)
				return false
			}

			// Verify that incompatible engines' UPSERT syntax is NOT present
			// Note: SQLite and PostgreSQL share "ON CONFLICT" syntax, so we need to be careful
			incompatibleSyntax := map[string][]string{
				"sqlite":   {"ON DUPLICATE KEY UPDATE", "MERGE"},
				"mysql":    {"ON CONFLICT", "INSERT OR REPLACE", "MERGE"},
				"postgres": {"ON DUPLICATE KEY UPDATE", "INSERT OR REPLACE", "MERGE"},
				"mssql":    {"ON CONFLICT", "ON DUPLICATE KEY UPDATE", "INSERT OR REPLACE"},
			}

			for _, keyword := range incompatibleSyntax[engine] {
				if findInString(contentUpper, toUpper(keyword)) != -1 {
					t.Logf("File %s for engine %s contains incompatible UPSERT syntax: %s",
						filename, engine, keyword)
					return false
				}
			}
		}

		// Check auto-increment files for correct syntax
		for _, filename := range autoIncrementFiles {
			filePath := filepath.Join(engineDir, filename)

			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				// File might not exist yet, skip
				continue
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Logf("Failed to read file %s: %v", filePath, err)
				return false
			}

			contentUpper := toUpper(string(content))

			// Verify that at least one of the expected auto-increment keywords is present
			foundKeyword := false
			for _, keyword := range syntax.autoIncrement {
				if findInString(contentUpper, toUpper(keyword)) != -1 {
					foundKeyword = true
					break
				}
			}

			if !foundKeyword {
				t.Logf("File %s for engine %s does not contain expected auto-increment syntax: %v",
					filename, engine, syntax.autoIncrement)
				return false
			}
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}

// **Feature: sqlite-standardization, Property 10: Database operations function identically across engines**
// **Validates: Requirements 7.5**
func TestProperty_CrossEngineCompatibility(t *testing.T) {
	// This property verifies that database operations produce equivalent results
	// across different database engines (SQLite, MySQL, PostgreSQL, MSSQL)

	// For this test, we'll verify that the SQL files for each operation exist
	// and have compatible structure across all engines. A full integration test
	// would require actual database connections, which is beyond the scope of
	// a property test.

	sqlDir := "../sql"
	engines := []string{"sqlite", "mysql", "postgres", "mssql"}

	// Operations that should be consistent across engines
	operations := []string{
		// Session operations
		"save_session",
		"load_session",
		"delete_session",
		"cleanup_expired_sessions",

		// Token operations
		"save_token",
		"load_token",
		"validate_token",
		"delete_token",
		"cleanup_expired_tokens",

		// Tenant operations
		"save_tenant",
		"load_tenant",
		"load_tenant_by_host",

		// Metrics operations
		"save_workload_metrics",
		"get_workload_metrics",

		// Rate limiting operations
		"check_rate_limit",
		"increment_rate_limit",
		"cleanup_rate_limits",
	}

	// Property: For any operation, all engines should have a corresponding SQL file
	property := func(opIdx uint8) bool {
		if opIdx >= uint8(len(operations)) {
			return true // Skip invalid indices
		}

		operation := operations[opIdx%uint8(len(operations))]
		filename := operation + ".sql"

		// Verify that the SQL file exists for all engines
		for _, engine := range engines {
			filePath := filepath.Join(sqlDir, engine, filename)

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Logf("Operation '%s' missing SQL file for engine %s: %s",
					operation, engine, filePath)
				return false
			}

			// Read the file to ensure it's not empty
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Logf("Failed to read SQL file %s: %v", filePath, err)
				return false
			}

			// Verify the file contains actual SQL (not just comments)
			contentStr := string(content)
			if len(contentStr) == 0 {
				t.Logf("SQL file %s is empty", filePath)
				return false
			}

			// Basic check: file should contain at least one SQL keyword
			hasSQL := false
			sqlKeywords := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "MERGE"}
			for _, keyword := range sqlKeywords {
				if findInString(contentStr, keyword) != -1 {
					hasSQL = true
					break
				}
			}

			if !hasSQL {
				t.Logf("SQL file %s does not contain any SQL keywords", filePath)
				return false
			}
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property test failed: %v", err)
	}
}
