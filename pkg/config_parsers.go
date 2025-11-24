package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// parseINI parses INI format configuration
func parseINI(data []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var currentSection string
	sectionData := make(map[string]interface{})

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			// Save previous section if exists
			if currentSection != "" {
				result[currentSection] = sectionData
				sectionData = make(map[string]interface{})
			}

			currentSection = strings.Trim(line, "[]")
			continue
		}

		// Parse key-value pair
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, "\"'")

		if currentSection == "" {
			// Root level key
			result[key] = parseValue(value)
		} else {
			// Section level key
			sectionData[key] = parseValue(value)
		}
	}

	// Save last section
	if currentSection != "" {
		result[currentSection] = sectionData
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning INI: %w", err)
	}

	return result, nil
}

// parseTOML parses TOML format configuration
func parseTOML(data []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var currentSection string
	var currentTable map[string]interface{}
	rootTable := result

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for table header [section]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			tableName := strings.Trim(line, "[]")

			// Handle nested tables with dots
			if strings.Contains(tableName, ".") {
				parts := strings.Split(tableName, ".")
				current := rootTable

				for i, part := range parts {
					if i == len(parts)-1 {
						// Last part - create new table
						newTable := make(map[string]interface{})
						current[part] = newTable
						currentTable = newTable
						currentSection = tableName
					} else {
						// Intermediate part - navigate or create
						if existing, ok := current[part]; ok {
							if existingMap, ok := existing.(map[string]interface{}); ok {
								current = existingMap
							} else {
								return nil, fmt.Errorf("invalid nested table structure")
							}
						} else {
							newTable := make(map[string]interface{})
							current[part] = newTable
							current = newTable
						}
					}
				}
			} else {
				// Simple table
				newTable := make(map[string]interface{})
				rootTable[tableName] = newTable
				currentTable = newTable
				currentSection = tableName
			}
			continue
		}

		// Parse key-value pair
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Parse TOML value
		parsedValue := parseTOMLValue(value)

		if currentSection == "" {
			// Root level
			rootTable[key] = parsedValue
		} else {
			// In a table
			currentTable[key] = parsedValue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning TOML: %w", err)
	}

	return result, nil
}

// parseTOMLValue parses a TOML value with type detection
func parseTOMLValue(value string) interface{} {
	value = strings.TrimSpace(value)

	// String (quoted)
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return strings.Trim(value, "\"'")
	}

	// Array
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		arrayStr := strings.Trim(value, "[]")
		if arrayStr == "" {
			return []interface{}{}
		}

		items := strings.Split(arrayStr, ",")
		result := make([]interface{}, len(items))
		for i, item := range items {
			result[i] = parseTOMLValue(strings.TrimSpace(item))
		}
		return result
	}

	// Boolean
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// Use generic parser for numbers
	return parseValue(value)
}

// parseYAML parses YAML format configuration (simplified parser)
func parseYAML(data []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var stack []map[string]interface{}
	var indentStack []int
	stack = append(stack, result)
	indentStack = append(indentStack, -1)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines and comments
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Calculate indentation
		indent := len(line) - len(strings.TrimLeft(line, " "))

		// Pop stack if we've decreased indentation
		for len(indentStack) > 1 && indent <= indentStack[len(indentStack)-1] {
			stack = stack[:len(stack)-1]
			indentStack = indentStack[:len(indentStack)-1]
		}

		// Parse key-value or key only
		if strings.Contains(trimmed, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			key := strings.TrimSpace(parts[0])
			valueStr := ""
			if len(parts) > 1 {
				valueStr = strings.TrimSpace(parts[1])
			}

			current := stack[len(stack)-1]

			if valueStr == "" {
				// This is a nested object
				newMap := make(map[string]interface{})
				current[key] = newMap
				stack = append(stack, newMap)
				indentStack = append(indentStack, indent)
			} else {
				// This is a key-value pair
				current[key] = parseYAMLValue(valueStr)
			}
		} else if strings.HasPrefix(trimmed, "-") {
			// Array item
			item := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))

			// For simplicity, we'll treat arrays as comma-separated values
			// A full YAML parser would be more complex
			current := stack[len(stack)-1]

			// Find or create array for last key
			// This is a simplified approach
			if item != "" {
				// Store as individual items with numeric keys
				arrayKey := fmt.Sprintf("item_%d", len(current))
				current[arrayKey] = parseYAMLValue(item)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning YAML: %w", err)
	}

	return result, nil
}

// parseYAMLValue parses a YAML value
func parseYAMLValue(value string) interface{} {
	value = strings.TrimSpace(value)

	// Remove quotes if present
	if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
		(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return strings.Trim(value, "\"'")
	}

	// Boolean
	if value == "true" || value == "yes" || value == "on" {
		return true
	}
	if value == "false" || value == "no" || value == "off" {
		return false
	}

	// Null
	if value == "null" || value == "~" {
		return nil
	}

	// Use generic parser for numbers
	return parseValue(value)
}
