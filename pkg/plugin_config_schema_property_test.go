package pkg

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Feature: compile-time-plugins, Property 4: Configuration Default Application**
// **Validates: Requirements 4.1, 4.2**
func TestProperty_ConfigurationDefaultApplication(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("defaults are applied when user config is empty", prop.ForAll(
		func(schema map[string]interface{}) bool {
			// Merge with empty user config
			merged := MergeConfigWithDefaults(nil, schema)

			// Verify all defaults from schema are present in merged config
			for key, schemaValue := range schema {
				schemaField, ok := parseSchemaField(schemaValue)
				if !ok {
					continue
				}

				if schemaField.Default != nil {
					mergedValue, exists := merged[key]
					if !exists {
						return false
					}
					// Check that the default value is present
					if !valuesEqual(mergedValue, schemaField.Default) {
						return false
					}
				}
			}

			return true
		},
		genConfigSchema(),
	))

	properties.Property("defaults are applied for missing fields", prop.ForAll(
		func(schema map[string]interface{}, userConfig map[string]interface{}) bool {
			merged := MergeConfigWithDefaults(userConfig, schema)

			// Check that defaults are applied for fields not in user config
			for key, schemaValue := range schema {
				schemaField, ok := parseSchemaField(schemaValue)
				if !ok {
					continue
				}

				if schemaField.Default != nil {
					// If user didn't provide this field, default should be present
					if _, userHasKey := userConfig[key]; !userHasKey {
						mergedValue, exists := merged[key]
						if !exists {
							return false
						}
						if !valuesEqual(mergedValue, schemaField.Default) {
							return false
						}
					}
				}
			}

			return true
		},
		genConfigSchema(),
		genUserConfig(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: compile-time-plugins, Property 5: Configuration Merging**
// **Validates: Requirements 4.3**
func TestProperty_ConfigurationMerging(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("user config overrides defaults", prop.ForAll(
		func(schema map[string]interface{}, userConfig map[string]interface{}) bool {
			merged := MergeConfigWithDefaults(userConfig, schema)

			// All user values should be present and unchanged
			for key, userValue := range userConfig {
				mergedValue, exists := merged[key]
				if !exists {
					return false
				}

				// For non-nested values, they should match exactly
				// For nested objects, we need to check recursively
				if schemaValue, hasSchema := schema[key]; hasSchema {
					schemaField, ok := parseSchemaField(schemaValue)
					if ok && schemaField.Type == ConfigTypeObject {
						// Nested object - will be merged recursively
						continue
					}
				}

				// For simple values, check equality
				if !valuesEqual(mergedValue, userValue) {
					return false
				}
			}

			return true
		},
		genConfigSchema(),
		genUserConfig(),
	))

	properties.Property("schema defaults fill missing user fields", prop.ForAll(
		func(schema map[string]interface{}) bool {
			// Create partial user config (only some fields)
			userConfig := make(map[string]interface{})
			count := 0
			for key := range schema {
				if count%2 == 0 { // Only include every other field
					userConfig[key] = "user-value"
				}
				count++
			}

			merged := MergeConfigWithDefaults(userConfig, schema)

			// Check that defaults are present for missing fields
			for key, schemaValue := range schema {
				schemaField, ok := parseSchemaField(schemaValue)
				if !ok {
					continue
				}

				if _, userHasKey := userConfig[key]; !userHasKey {
					// User didn't provide this field
					if schemaField.Default != nil {
						mergedValue, exists := merged[key]
						if !exists {
							return false
						}
						if !valuesEqual(mergedValue, schemaField.Default) {
							return false
						}
					}
				}
			}

			return true
		},
		genConfigSchema(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: compile-time-plugins, Property 6: Required Configuration Validation**
// **Validates: Requirements 4.4**
func TestProperty_RequiredConfigurationValidation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("missing required fields cause validation error", prop.ForAll(
		func(schema map[string]interface{}) bool {
			// Find required fields in schema
			requiredFields := make([]string, 0)
			for key, schemaValue := range schema {
				schemaField, ok := parseSchemaField(schemaValue)
				if ok && schemaField.Required {
					requiredFields = append(requiredFields, key)
				}
			}

			if len(requiredFields) == 0 {
				// No required fields, validation should pass
				return true
			}

			// Create config missing the first required field
			config := make(map[string]interface{})
			for key := range schema {
				if key != requiredFields[0] {
					config[key] = "some-value"
				}
			}

			// Validation should fail
			err := ValidateRequiredFields(config, schema)
			return err != nil
		},
		genConfigSchemaWithRequired(),
	))

	properties.Property("present required fields pass validation", prop.ForAll(
		func(schema map[string]interface{}) bool {
			// Create config with all required fields present
			config := make(map[string]interface{})
			for key, schemaValue := range schema {
				schemaField, ok := parseSchemaField(schemaValue)
				if ok && schemaField.Required {
					config[key] = "required-value"
				}
			}

			// Validation should pass
			err := ValidateRequiredFields(config, schema)
			return err == nil
		},
		genConfigSchemaWithRequired(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Feature: compile-time-plugins, Property 7: Nested Configuration Defaults**
// **Validates: Requirements 4.5**
func TestProperty_NestedConfigurationDefaults(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("nested defaults are applied recursively", prop.ForAll(
		func(nestedSchema map[string]interface{}) bool {
			// Create a schema with nested object
			schema := map[string]interface{}{
				"nested": map[string]interface{}{
					"type":       "object",
					"properties": nestedSchema,
				},
			}

			// Merge with empty user config
			merged := MergeConfigWithDefaults(nil, schema)

			// Check that nested defaults are applied
			nestedMerged, ok := merged["nested"].(map[string]interface{})
			if !ok {
				return false
			}

			for key, schemaValue := range nestedSchema {
				schemaField, ok := parseSchemaField(schemaValue)
				if !ok {
					continue
				}

				if schemaField.Default != nil {
					value, exists := nestedMerged[key]
					if !exists {
						return false
					}
					if !valuesEqual(value, schemaField.Default) {
						return false
					}
				}
			}

			return true
		},
		genConfigSchema(),
	))

	properties.Property("nested user values override nested defaults", prop.ForAll(
		func(nestedSchema map[string]interface{}) bool {
			// Create a schema with nested object
			schema := map[string]interface{}{
				"nested": map[string]interface{}{
					"type":       "object",
					"properties": nestedSchema,
				},
			}

			// Create user config with nested values
			userNested := make(map[string]interface{})
			for key := range nestedSchema {
				userNested[key] = "user-nested-value"
			}
			userConfig := map[string]interface{}{
				"nested": userNested,
			}

			// Merge
			merged := MergeConfigWithDefaults(userConfig, schema)

			// Check that user values are present
			nestedMerged, ok := merged["nested"].(map[string]interface{})
			if !ok {
				return false
			}

			for key := range nestedSchema {
				value, exists := nestedMerged[key]
				if !exists {
					return false
				}
				if value != "user-nested-value" {
					return false
				}
			}

			return true
		},
		genConfigSchema(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Generator for configuration schemas
func genConfigSchema() gopter.Gen {
	return gen.SliceOfN(5, gen.AlphaString().SuchThat(func(s string) bool { return s != "" })).
		Map(func(keys []string) map[string]interface{} {
			schema := make(map[string]interface{})
			for i, key := range keys {
				switch i % 4 {
				case 0:
					// String field with default
					schema[key] = map[string]interface{}{
						"type":    "string",
						"default": "default-value",
					}
				case 1:
					// Int field with default
					schema[key] = map[string]interface{}{
						"type":    "int",
						"default": 42,
					}
				case 2:
					// Bool field with default
					schema[key] = map[string]interface{}{
						"type":    "bool",
						"default": true,
					}
				case 3:
					// Field without default
					schema[key] = map[string]interface{}{
						"type": "string",
					}
				}
			}
			return schema
		}).SuchThat(func(v interface{}) bool {
		m := v.(map[string]interface{})
		return len(m) > 0
	})
}

// Generator for configuration schemas with required fields
func genConfigSchemaWithRequired() gopter.Gen {
	return gen.SliceOfN(5, gen.AlphaString().SuchThat(func(s string) bool { return s != "" })).
		Map(func(keys []string) map[string]interface{} {
			schema := make(map[string]interface{})
			for i, key := range keys {
				switch i % 3 {
				case 0:
					// Required string field
					schema[key] = map[string]interface{}{
						"type":     "string",
						"required": true,
					}
				case 1:
					// Optional string field with default
					schema[key] = map[string]interface{}{
						"type":    "string",
						"default": "default-value",
					}
				case 2:
					// Required int field
					schema[key] = map[string]interface{}{
						"type":     "int",
						"required": true,
					}
				}
			}
			return schema
		}).SuchThat(func(v interface{}) bool {
		// Ensure at least one required field exists
		schema := v.(map[string]interface{})
		for _, schemaValue := range schema {
			if schemaMap, ok := schemaValue.(map[string]interface{}); ok {
				if required, ok := schemaMap["required"].(bool); ok && required {
					return true
				}
			}
		}
		return false
	})
}

// Generator for user configuration
func genUserConfig() gopter.Gen {
	return gen.SliceOfN(5, gen.AlphaString().SuchThat(func(s string) bool { return s != "" })).
		Map(func(keys []string) map[string]interface{} {
			config := make(map[string]interface{})
			for i, key := range keys {
				switch i % 4 {
				case 0:
					config[key] = "user-string"
				case 1:
					config[key] = 123
				case 2:
					config[key] = true
				case 3:
					config[key] = false
				}
			}
			return config
		})
}

// valuesEqual compares two values for equality
func valuesEqual(a, b interface{}) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// For maps, compare recursively
	if aMap, ok := a.(map[string]interface{}); ok {
		if bMap, ok := b.(map[string]interface{}); ok {
			if len(aMap) != len(bMap) {
				return false
			}
			for key, aValue := range aMap {
				bValue, exists := bMap[key]
				if !exists {
					return false
				}
				if !valuesEqual(aValue, bValue) {
					return false
				}
			}
			return true
		}
		return false
	}

	// For simple types, use direct comparison
	return a == b
}
