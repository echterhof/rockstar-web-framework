package pkg

import (
	"fmt"
)

// ConfigSchemaType represents the type of a configuration field
type ConfigSchemaType string

const (
	ConfigTypeString   ConfigSchemaType = "string"
	ConfigTypeInt      ConfigSchemaType = "int"
	ConfigTypeBool     ConfigSchemaType = "bool"
	ConfigTypeFloat    ConfigSchemaType = "float"
	ConfigTypeDuration ConfigSchemaType = "duration"
	ConfigTypeObject   ConfigSchemaType = "object"
	ConfigTypeArray    ConfigSchemaType = "array"
)

// ConfigSchemaField represents a single field in a configuration schema
type ConfigSchemaField struct {
	Type        ConfigSchemaType       `json:"type" yaml:"type"`
	Default     interface{}            `json:"default,omitempty" yaml:"default,omitempty"`
	Required    bool                   `json:"required,omitempty" yaml:"required,omitempty"`
	Description string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty" yaml:"properties,omitempty"` // For nested objects
}

// MergeConfigWithDefaults merges user configuration with schema defaults
// User values take precedence over defaults
func MergeConfigWithDefaults(userConfig map[string]interface{}, schema map[string]interface{}) map[string]interface{} {
	if schema == nil {
		if userConfig == nil {
			return make(map[string]interface{})
		}
		return userConfig
	}

	result := make(map[string]interface{})

	// First, apply defaults from schema
	for key, schemaValue := range schema {
		schemaField, ok := parseSchemaField(schemaValue)
		if !ok {
			continue
		}

		// Apply default value if present
		if schemaField.Default != nil {
			result[key] = schemaField.Default
		}

		// Handle nested objects
		if schemaField.Type == ConfigTypeObject && schemaField.Properties != nil {
			var nestedUser map[string]interface{}
			if userConfig != nil {
				if nu, ok := userConfig[key].(map[string]interface{}); ok {
					nestedUser = nu
				}
			}
			result[key] = MergeConfigWithDefaults(nestedUser, schemaField.Properties)
		}
	}

	// Then, override with user values
	if userConfig != nil {
		for key, value := range userConfig {
			// Check if this is a nested object in the schema
			if schemaValue, hasSchema := schema[key]; hasSchema {
				schemaField, ok := parseSchemaField(schemaValue)
				if ok && schemaField.Type == ConfigTypeObject && schemaField.Properties != nil {
					// Recursively merge nested objects
					if valueMap, ok := value.(map[string]interface{}); ok {
						result[key] = MergeConfigWithDefaults(valueMap, schemaField.Properties)
						continue
					}
				}
			}
			// For non-nested values or when no schema, just use the user value
			result[key] = value
		}
	}

	return result
}

// ValidateRequiredFields checks that all required fields are present in the configuration
func ValidateRequiredFields(config map[string]interface{}, schema map[string]interface{}) error {
	if schema == nil {
		return nil
	}

	for key, schemaValue := range schema {
		schemaField, ok := parseSchemaField(schemaValue)
		if !ok {
			continue
		}

		// Check if required field is missing
		if schemaField.Required {
			value, exists := config[key]
			if !exists || value == nil {
				return fmt.Errorf("required configuration field '%s' is missing", key)
			}
		}

		// Recursively validate nested objects
		if schemaField.Type == ConfigTypeObject && schemaField.Properties != nil {
			if value, exists := config[key]; exists {
				if nestedConfig, ok := value.(map[string]interface{}); ok {
					if err := ValidateRequiredFields(nestedConfig, schemaField.Properties); err != nil {
						return fmt.Errorf("in field '%s': %w", key, err)
					}
				}
			}
		}
	}

	return nil
}

// parseSchemaField attempts to parse a schema value into a ConfigSchemaField
func parseSchemaField(schemaValue interface{}) (ConfigSchemaField, bool) {
	field := ConfigSchemaField{}

	schemaMap, ok := schemaValue.(map[string]interface{})
	if !ok {
		return field, false
	}

	// Parse type
	if typeVal, ok := schemaMap["type"]; ok {
		if typeStr, ok := typeVal.(string); ok {
			field.Type = ConfigSchemaType(typeStr)
		}
	}

	// Parse default
	if defaultVal, ok := schemaMap["default"]; ok {
		field.Default = defaultVal
	}

	// Parse required
	if requiredVal, ok := schemaMap["required"]; ok {
		if requiredBool, ok := requiredVal.(bool); ok {
			field.Required = requiredBool
		}
	}

	// Parse description
	if descVal, ok := schemaMap["description"]; ok {
		if descStr, ok := descVal.(string); ok {
			field.Description = descStr
		}
	}

	// Parse properties (for nested objects)
	if propsVal, ok := schemaMap["properties"]; ok {
		if propsMap, ok := propsVal.(map[string]interface{}); ok {
			field.Properties = propsMap
		}
	}

	return field, true
}
