package generator

import (
	"fmt"
	"hash/fnv"
	"math"
	"strings"
	"time"

	"github.com/specmint/specmint/pkg/schema"
	mathrand "math/rand"
)

// DeterministicGenerator generates values using seeded RNG for reproducibility
type DeterministicGenerator struct {
	baseSeed int64
	rng      *mathrand.Rand
}

// NewDeterministicGenerator creates a new deterministic generator
func NewDeterministicGenerator(seed int64) *DeterministicGenerator {
	return &DeterministicGenerator{
		baseSeed: seed,
		rng:      mathrand.New(mathrand.NewSource(seed)),
	}
}

// GenerateValue generates a deterministic value for a schema node
func (g *DeterministicGenerator) GenerateValue(node *schema.SchemaNode, recordIndex int) (interface{}, error) {
	// Create seed for this specific field and record
	seed := g.deriveSeed(node.Path, recordIndex)
	rng := mathrand.New(mathrand.NewSource(seed))

	return g.generateValue(node, rng)
}

// deriveSeed creates a deterministic seed based on path and record index
func (g *DeterministicGenerator) deriveSeed(path string, recordIndex int) int64 {
	h := fnv.New64a()
	_, err := h.Write([]byte(path))
	if err != nil {
		return 0
	}
	_, err = h.Write([]byte{byte(recordIndex), byte(recordIndex >> 8), byte(recordIndex >> 16), byte(recordIndex >> 24)})
	if err != nil {
		return 0
	}
	pathHash := int64(h.Sum64() & 0x7FFFFFFFFFFFFFFF) // Ensure positive

	return g.baseSeed ^ pathHash
}

// generateValue generates a value based on the schema node type and constraints
func (g *DeterministicGenerator) generateValue(node *schema.SchemaNode, rng *mathrand.Rand) (interface{}, error) {
	// Handle enum values first
	if len(node.Enum) > 0 {
		idx := rng.Intn(len(node.Enum))
		return node.Enum[idx], nil
	}

	// Handle examples if available
	if len(node.Examples) > 0 && rng.Float64() < 0.7 { // 70% chance to use examples
		idx := rng.Intn(len(node.Examples))
		return node.Examples[idx], nil
	}

	// Generate based on type
	switch node.Type {
	case "string":
		return g.generateString(node, rng)
	case "integer":
		return g.generateInteger(node, rng)
	case "number":
		return g.generateNumber(node, rng)
	case "boolean":
		return rng.Float64() < 0.5, nil
	case "array":
		return g.generateArray(node, rng)
	case "object":
		return g.generateObject(node, rng)
	case "null":
		return nil, nil
	default:
		return g.generateString(node, rng) // Default to string
	}
}

// generateString generates string values with format and pattern constraints
func (g *DeterministicGenerator) generateString(node *schema.SchemaNode, rng *mathrand.Rand) (string, error) {
	// Handle specific formats
	switch node.Format {
	case "email":
		return g.generateEmail(rng), nil
	case "uuid":
		return g.generateUUID(rng), nil
	case "date":
		return g.generateDate(rng), nil
	case "date-time":
		return g.generateDateTime(rng), nil
	case "uri":
		return g.generateURI(rng), nil
	case "phone":
		return g.generatePhone(rng), nil
	}

	// Handle pattern constraint
	if node.Pattern != "" {
		return g.generateFromPattern(node.Pattern, rng)
	}

	// Generate based on length constraints
	minLen := 5
	maxLen := 20

	if node.MinLength != nil {
		minLen = *node.MinLength
	}
	if node.MaxLength != nil {
		maxLen = *node.MaxLength
		if maxLen < minLen {
			maxLen = minLen
		}
	}

	length := minLen + rng.Intn(maxLen-minLen+1)
	return g.generateRandomString(length, rng), nil
}

// generateInteger generates integer values with min/max constraints
func (g *DeterministicGenerator) generateInteger(node *schema.SchemaNode, rng *mathrand.Rand) (int64, error) {
	min := int64(0)
	max := int64(1000)

	if node.Minimum != nil {
		min = int64(*node.Minimum)
	}
	if node.Maximum != nil {
		max = int64(*node.Maximum)
	}

	if max < min {
		max = min
	}

	value := min + rng.Int63n(max-min+1)

	// Apply multipleOf constraint
	if node.MultipleOf != nil {
		multiple := int64(*node.MultipleOf)
		if multiple > 0 {
			value = (value / multiple) * multiple
		}
	}

	return value, nil
}

// generateNumber generates float values with min/max constraints
func (g *DeterministicGenerator) generateNumber(node *schema.SchemaNode, rng *mathrand.Rand) (float64, error) {
	min := 0.0
	max := 1000.0

	if node.Minimum != nil {
		min = *node.Minimum
	}
	if node.Maximum != nil {
		max = *node.Maximum
	}

	if max < min {
		max = min
	}

	value := min + rng.Float64()*(max-min)

	// Apply multipleOf constraint
	if node.MultipleOf != nil && *node.MultipleOf > 0 {
		value = math.Round(value/(*node.MultipleOf)) * (*node.MultipleOf)
	}

	return value, nil
}

// generateArray generates array values with item constraints
func (g *DeterministicGenerator) generateArray(node *schema.SchemaNode, rng *mathrand.Rand) ([]interface{}, error) {
	if node.Items == nil {
		return []interface{}{}, nil
	}

	minItems := 1
	maxItems := 5

	if node.MinItems != nil {
		minItems = *node.MinItems
	}
	if node.MaxItems != nil {
		maxItems = *node.MaxItems
		if maxItems < minItems {
			maxItems = minItems
		}
	}

	length := minItems + rng.Intn(maxItems-minItems+1)
	result := make([]interface{}, length)

	for i := 0; i < length; i++ {
		// Create unique seed for each array item
		itemSeed := g.deriveSeed(fmt.Sprintf("%s[%d]", node.Path, i), 0)
		itemRng := mathrand.New(mathrand.NewSource(itemSeed))

		value, err := g.generateValue(node.Items, itemRng)
		if err != nil {
			return nil, fmt.Errorf("failed to generate array item %d: %w", i, err)
		}
		result[i] = value
	}

	return result, nil
}

// generateObject generates object values with property constraints
func (g *DeterministicGenerator) generateObject(node *schema.SchemaNode, rng *mathrand.Rand) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if node.Properties == nil {
		return result, nil
	}

	// Generate required fields first
	for _, propName := range node.Required {
		if prop, exists := node.Properties[propName]; exists {
			value, err := g.generateValue(prop, rng)
			if err != nil {
				return nil, fmt.Errorf("failed to generate required property %s: %w", propName, err)
			}
			result[propName] = value
		}
	}

	// Generate optional fields with probability
	requiredMap := make(map[string]bool)
	for _, req := range node.Required {
		requiredMap[req] = true
	}

	for propName, prop := range node.Properties {
		if !requiredMap[propName] {
			// Use field-specific probability
			if rng.Float64() < prop.OptionalProb {
				value, err := g.generateValue(prop, rng)
				if err != nil {
					return nil, fmt.Errorf("failed to generate optional property %s: %w", propName, err)
				}
				result[propName] = value
			}
		}
	}

	return result, nil
}

// Format-specific generators

func (g *DeterministicGenerator) generateEmail(rng *mathrand.Rand) string {
	domains := []string{"example.com", "test.org", "sample.net", "demo.co"}
	names := []string{"user", "test", "demo", "sample", "john", "jane", "admin"}

	name := names[rng.Intn(len(names))]
	domain := domains[rng.Intn(len(domains))]
	suffix := rng.Intn(1000)

	return fmt.Sprintf("%s%d@%s", name, suffix, domain)
}

func (g *DeterministicGenerator) generateUUID(rng *mathrand.Rand) string {
	b := make([]byte, 16)
	_, err := rng.Read(b)
	if err != nil {
		return ""
	}

	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func (g *DeterministicGenerator) generateDate(rng *mathrand.Rand) string {
	// Generate date within last 5 years
	now := time.Now()
	start := now.AddDate(-5, 0, 0)
	days := int(now.Sub(start).Hours() / 24)

	randomDays := rng.Intn(days)
	date := start.AddDate(0, 0, randomDays)

	return date.Format("2006-01-02")
}

func (g *DeterministicGenerator) generateDateTime(rng *mathrand.Rand) string {
	// Generate datetime within last year
	now := time.Now()
	start := now.AddDate(-1, 0, 0)
	duration := now.Sub(start)

	randomDuration := time.Duration(rng.Int63n(int64(duration)))
	dateTime := start.Add(randomDuration)

	return dateTime.Format(time.RFC3339)
}

func (g *DeterministicGenerator) generateURI(rng *mathrand.Rand) string {
	schemes := []string{"http", "https"}
	hosts := []string{"example.com", "test.org", "api.sample.net"}
	paths := []string{"/api/v1", "/data", "/users", "/items"}

	scheme := schemes[rng.Intn(len(schemes))]
	host := hosts[rng.Intn(len(hosts))]
	path := paths[rng.Intn(len(paths))]
	id := rng.Intn(10000)

	return fmt.Sprintf("%s://%s%s/%d", scheme, host, path, id)
}

func (g *DeterministicGenerator) generatePhone(rng *mathrand.Rand) string {
	// Generate US phone number format
	area := 200 + rng.Intn(800)
	exchange := 200 + rng.Intn(800)
	number := rng.Intn(10000)

	return fmt.Sprintf("(%03d) %03d-%04d", area, exchange, number)
}

func (g *DeterministicGenerator) generateFromPattern(pattern string, rng *mathrand.Rand) (string, error) {
	// Simple pattern generation - could be enhanced with proper regex generation
	// For now, generate a string that might match common patterns

	if strings.Contains(pattern, "[0-9]") {
		// Numeric pattern
		return fmt.Sprintf("%d", rng.Intn(1000000)), nil
	}

	if strings.Contains(pattern, "[a-zA-Z]") {
		// Alphabetic pattern
		return g.generateRandomString(8, rng), nil
	}

	// Default to random string
	return g.generateRandomString(10, rng), nil
}

func (g *DeterministicGenerator) generateRandomString(length int, rng *mathrand.Rand) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	for i := range result {
		result[i] = charset[rng.Intn(len(charset))]
	}

	return string(result)
}
