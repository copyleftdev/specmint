package generator

import (
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
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
	// Enhanced pattern generation with specific pattern recognition

	// Handle common e-commerce patterns
	switch pattern {
	case "^[A-Z]{2}[0-9]{6}$":
		// SKU format: 2 uppercase letters + 6 digits
		letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		result := make([]rune, 8)
		result[0] = letters[rng.Intn(len(letters))]
		result[1] = letters[rng.Intn(len(letters))]
		for i := 2; i < 8; i++ {
			result[i] = rune('0' + rng.Intn(10))
		}
		return string(result), nil

	case "^PRD[0-9]{8}$":
		// Product ID format: PRD + 8 digits
		return fmt.Sprintf("PRD%08d", rng.Intn(100000000)), nil

	case "^PRD-[0-9]{6}$":
		// Product ID format: PRD- + 6 digits
		return fmt.Sprintf("PRD-%06d", rng.Intn(1000000)), nil

	case "^WH[0-9]{3}$":
		// Warehouse format: WH + 3 digits
		return fmt.Sprintf("WH%03d", rng.Intn(1000)), nil

	case "^SUP[0-9]{5}$":
		// Supplier format: SUP + 5 digits
		return fmt.Sprintf("SUP%05d", rng.Intn(100000)), nil

	case "^TXN-[0-9]{10}$":
		// Transaction ID format: TXN- + 10 digits
		return fmt.Sprintf("TXN-%010d", rng.Intn(1000000000)), nil

	case "^[0-9]{10}$":
		// 10 digit number (account numbers, NPI)
		return fmt.Sprintf("%010d", rng.Intn(1000000000)), nil

	case "^[0-9]{9}$":
		// 9 digit number (routing numbers)
		return fmt.Sprintf("%09d", rng.Intn(1000000000)), nil

	case "^[0-9]{4}$":
		// 4 digit number (MCC codes)
		return fmt.Sprintf("%04d", rng.Intn(10000)), nil

	case "^[0-9]{5}$":
		// 5 digit number (procedure codes)
		return fmt.Sprintf("%05d", rng.Intn(100000)), nil

	case "^[A-Z][0-9]{2}\\.[0-9]{1,2}$":
		// ICD-10 format: Letter + 2 digits + dot + 1-2 digits
		letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		letter := letters[rng.Intn(len(letters))]
		first := rng.Intn(100)
		second := rng.Intn(100)
		return fmt.Sprintf("%c%02d.%02d", letter, first, second), nil

	case "^[A-Z]{2}-[A-Z]{3}-[0-9]{3}$":
		// Warehouse location format: XX-XXX-000
		letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		result := make([]rune, 9)
		result[0] = letters[rng.Intn(len(letters))]
		result[1] = letters[rng.Intn(len(letters))]
		result[2] = '-'
		result[3] = letters[rng.Intn(len(letters))]
		result[4] = letters[rng.Intn(len(letters))]
		result[5] = letters[rng.Intn(len(letters))]
		result[6] = '-'
		result[7] = rune('0' + rng.Intn(10))
		result[8] = rune('0' + rng.Intn(10))
		result = append(result, rune('0'+rng.Intn(10)))
		return string(result), nil

	// X12 EDI specific patterns
	case "^PO[0-9]{8}$":
		// Purchase Order format: PO + 8 digits
		return fmt.Sprintf("PO%08d", rng.Intn(100000000)), nil

	case "^[A-Z0-9]{2,15}$":
		// Party ID format: 2-15 alphanumeric characters
		length := 2 + rng.Intn(14) // 2-15 characters
		charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		result := make([]rune, length)
		for i := range result {
			result[i] = rune(charset[rng.Intn(len(charset))])
		}
		return string(result), nil

	case "^[A-Z0-9]{6,20}$":
		// Product ID format: 6-20 alphanumeric characters
		length := 6 + rng.Intn(15) // 6-20 characters
		charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		result := make([]rune, length)
		for i := range result {
			result[i] = rune(charset[rng.Intn(len(charset))])
		}
		return string(result), nil

	case "^MPN[A-Z0-9]{8,15}$":
		// Manufacturer Part Number format: MPN + 8-15 alphanumeric
		length := 8 + rng.Intn(8) // 8-15 characters after MPN
		charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		result := "MPN"
		for i := 0; i < length; i++ {
			result += string(charset[rng.Intn(len(charset))])
		}
		return result, nil

	case "^[A-Z]{2}$":
		// 2-letter state/country code
		letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		return fmt.Sprintf("%c%c",
			letters[rng.Intn(len(letters))],
			letters[rng.Intn(len(letters))]), nil

	case "^[0-9]{5}(-[0-9]{4})?$":
		// ZIP code format: 5 digits or ZIP+4
		zip5 := fmt.Sprintf("%05d", rng.Intn(100000))
		if rng.Float32() < 0.3 { // 30% chance of ZIP+4
			zip4 := fmt.Sprintf("%04d", rng.Intn(10000))
			return fmt.Sprintf("%s-%s", zip5, zip4), nil
		}
		return zip5, nil

	// Medical/Pharmacy specific patterns
	case "^RX[0-9]{8}$":
		return fmt.Sprintf("RX%08d", rng.Intn(100000000)), nil
	case "^[0-9]{5}-[0-9]{4}-[0-9]{2}$":
		// NDC code format
		return fmt.Sprintf("%05d-%04d-%02d",
			rng.Intn(100000), rng.Intn(10000), rng.Intn(100)), nil
	case "^[A-Z]{2}[0-9]{7}$":
		// DEA number format
		letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		return fmt.Sprintf("%c%c%07d",
			letters[rng.Intn(26)], letters[rng.Intn(26)], rng.Intn(10000000)), nil
	case "^PA[0-9]{8}$":
		// Prior authorization number
		return fmt.Sprintf("PA%08d", rng.Intn(100000000)), nil
	case "^INS[0-9]{6}$":
		// Insurance ID format
		return fmt.Sprintf("INS%06d", rng.Intn(1000000)), nil

	// Healthcare Claims 837 patterns
	case "^CLM[0-9]{10}$":
		// Claim control number
		return fmt.Sprintf("CLM%010d", rng.Intn(10000000000)), nil
	case "^[A-Z0-9]{8,15}$":
		// Insurance member ID
		chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		length := 8 + rng.Intn(8) // 8-15 characters
		result := make([]byte, length)
		for i := range result {
			result[i] = chars[rng.Intn(len(chars))]
		}
		return string(result), nil
	case "^[0-9]{2}-[0-9]{7}$":
		// Federal Tax ID format
		return fmt.Sprintf("%02d-%07d", rng.Intn(100), rng.Intn(10000000)), nil
	case "^[A-Z0-9]{5,10}$":
		// Payer ID
		chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		length := 5 + rng.Intn(6) // 5-10 characters
		result := make([]byte, length)
		for i := range result {
			result[i] = chars[rng.Intn(len(chars))]
		}
		return string(result), nil
	case "^[A-Z][0-9]{2}\\.[0-9A-Z]{1,4}$":
		// ICD-10 diagnosis code format
		letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		digits := "0123456789"
		alphanumeric := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		suffixLength := 1 + rng.Intn(4) // 1-4 characters
		suffix := make([]byte, suffixLength)
		for i := range suffix {
			suffix[i] = alphanumeric[rng.Intn(len(alphanumeric))]
		}
		return fmt.Sprintf("%c%c%c.%s",
			letters[rng.Intn(26)],
			digits[rng.Intn(10)],
			digits[rng.Intn(10)],
			string(suffix)), nil

	case "^[A-Z0-9]{6,12}$":
		// Insurance group number format: 6-12 alphanumeric
		length := 6 + rng.Intn(7) // 6-12 characters
		charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		result := make([]rune, length)
		for i := range result {
			result[i] = rune(charset[rng.Intn(len(charset))])
		}
		return string(result), nil

	case "^[0-9]{6}$":
		// BIN (Bank Identification Number) format: 6 digits
		return fmt.Sprintf("%06d", rng.Intn(1000000)), nil

	case "^[A-Z0-9]{3,10}$":
		// PCN (Processor Control Number) format: 3-10 alphanumeric
		length := 3 + rng.Intn(8) // 3-10 characters
		charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		result := make([]rune, length)
		for i := range result {
			result[i] = rune(charset[rng.Intn(len(charset))])
		}
		return string(result), nil
	}

	// Fallback: analyze pattern structure
	if strings.Contains(pattern, "[0-9]") && strings.Contains(pattern, "[A-Z]") {
		// Mixed alphanumeric pattern
		return g.generateMixedPattern(pattern, rng)
	}

	if strings.Contains(pattern, "[0-9]") {
		// Numeric pattern - extract length from pattern
		length := g.extractNumericLength(pattern)
		return fmt.Sprintf("%0*d", length, rng.Intn(int(math.Pow(10, float64(length))))), nil
	}

	if strings.Contains(pattern, "[a-zA-Z]") || strings.Contains(pattern, "[A-Z]") {
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

// generateMixedPattern generates strings for mixed alphanumeric patterns
func (g *DeterministicGenerator) generateMixedPattern(pattern string, rng *mathrand.Rand) (string, error) {
	// Simple implementation for mixed patterns
	// This could be enhanced with proper regex parsing
	return g.generateRandomString(8, rng), nil
}

// extractNumericLength extracts the expected length from numeric patterns
func (g *DeterministicGenerator) extractNumericLength(pattern string) int {
	// Extract length from patterns like [0-9]{6} or {10}
	if strings.Contains(pattern, "{") && strings.Contains(pattern, "}") {
		start := strings.Index(pattern, "{") + 1
		end := strings.Index(pattern, "}")
		if end > start {
			if length, err := strconv.Atoi(pattern[start:end]); err == nil {
				return length
			}
		}
	}
	// Default length for numeric patterns
	return 6
}
