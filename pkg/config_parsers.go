package pkg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// parseJSON parses JSON format configuration
func parseJSON(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return result, nil
}

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

// parseTOML parses TOML format configuration using github.com/BurntSushi/toml
func parseTOML(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := toml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse TOML: %w", err)
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

// parseYAML parses YAML format configuration using yaml.v3
func parseYAML(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
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
