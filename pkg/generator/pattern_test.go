package generator

import (
	"math/rand"
	"regexp"
	"testing"
	"time"
)

// TestPatternGeneration_PropertyBased uses property-based testing to verify
// that generated strings always match their expected patterns
func TestPatternGeneration_PropertyBased(t *testing.T) {
	testCases := []struct {
		name    string
		pattern string
		regex   *regexp.Regexp
	}{
		{
			name:    "SKU_Format",
			pattern: "^[A-Z]{2}[0-9]{6}$",
			regex:   regexp.MustCompile("^[A-Z]{2}[0-9]{6}$"),
		},
		{
			name:    "Product_ID_PRD",
			pattern: "^PRD[0-9]{8}$",
			regex:   regexp.MustCompile("^PRD[0-9]{8}$"),
		},
		{
			name:    "Product_ID_PRD_Dash",
			pattern: "^PRD-[0-9]{6}$",
			regex:   regexp.MustCompile("^PRD-[0-9]{6}$"),
		},
		{
			name:    "Warehouse_Format",
			pattern: "^WH[0-9]{3}$",
			regex:   regexp.MustCompile("^WH[0-9]{3}$"),
		},
		{
			name:    "Supplier_Format",
			pattern: "^SUP[0-9]{5}$",
			regex:   regexp.MustCompile("^SUP[0-9]{5}$"),
		},
		{
			name:    "Transaction_ID",
			pattern: "^TXN-[0-9]{10}$",
			regex:   regexp.MustCompile("^TXN-[0-9]{10}$"),
		},
		{
			name:    "Ten_Digit_Number",
			pattern: "^[0-9]{10}$",
			regex:   regexp.MustCompile("^[0-9]{10}$"),
		},
		{
			name:    "Nine_Digit_Number",
			pattern: "^[0-9]{9}$",
			regex:   regexp.MustCompile("^[0-9]{9}$"),
		},
		{
			name:    "ICD10_Format",
			pattern: "^[A-Z][0-9]{2}\\.[0-9]{1,2}$",
			regex:   regexp.MustCompile("^[A-Z][0-9]{2}\\.[0-9]{1,2}$"),
		},
		{
			name:    "Warehouse_Location",
			pattern: "^[A-Z]{2}-[A-Z]{3}-[0-9]{3}$",
			regex:   regexp.MustCompile("^[A-Z]{2}-[A-Z]{3}-[0-9]{3}$"),
		},
		// X12 EDI patterns
		{
			name:    "X12_Purchase_Order",
			pattern: "^PO[0-9]{8}$",
			regex:   regexp.MustCompile("^PO[0-9]{8}$"),
		},
		{
			name:    "X12_Party_ID_Short",
			pattern: "^[A-Z0-9]{2,15}$",
			regex:   regexp.MustCompile("^[A-Z0-9]{2,15}$"),
		},
		{
			name:    "X12_Product_ID",
			pattern: "^[A-Z0-9]{6,20}$",
			regex:   regexp.MustCompile("^[A-Z0-9]{6,20}$"),
		},
		{
			name:    "X12_Manufacturer_Part",
			pattern: "^MPN[A-Z0-9]{8,15}$",
			regex:   regexp.MustCompile("^MPN[A-Z0-9]{8,15}$"),
		},
		{
			name:    "X12_State_Code",
			pattern: "^[A-Z]{2}$",
			regex:   regexp.MustCompile("^[A-Z]{2}$"),
		},
		{
			name:    "X12_ZIP_Code",
			pattern: "^[0-9]{5}(-[0-9]{4})?$",
			regex:   regexp.MustCompile("^[0-9]{5}(-[0-9]{4})?$"),
		},
	}

	generator := NewDeterministicGenerator(12345)

	// Property-based test: for each pattern, generate multiple values
	// and verify they ALL match the expected regex
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with multiple seeds to ensure consistency across different inputs
			for seed := int64(1); seed <= 100; seed++ {
				rng := rand.New(rand.NewSource(seed))
				
				generated, err := generator.generateFromPattern(tc.pattern, rng)
				if err != nil {
					t.Fatalf("Failed to generate pattern for %s: %v", tc.pattern, err)
				}

				if !tc.regex.MatchString(generated) {
					t.Errorf("Generated value '%s' does not match pattern '%s' (seed: %d)", 
						generated, tc.pattern, seed)
				}

				// Additional property: generated strings should be deterministic for same seed
				rng2 := rand.New(rand.NewSource(seed))
				generated2, err := generator.generateFromPattern(tc.pattern, rng2)
				if err != nil {
					t.Fatalf("Failed to generate pattern for %s (second attempt): %v", tc.pattern, err)
				}

				if generated != generated2 {
					t.Errorf("Non-deterministic generation for pattern '%s': '%s' != '%s' (seed: %d)",
						tc.pattern, generated, generated2, seed)
				}
			}
		})
	}
}

// TestPatternGeneration_EdgeCases tests edge cases and boundary conditions
func TestPatternGeneration_EdgeCases(t *testing.T) {
	generator := NewDeterministicGenerator(12345)
	rng := rand.New(rand.NewSource(42))

	testCases := []struct {
		name     string
		pattern  string
		expected string // expected exact match or empty for regex validation
		regex    *regexp.Regexp
	}{
		{
			name:    "Empty_Pattern",
			pattern: "",
			regex:   regexp.MustCompile(".*"), // should match anything
		},
		{
			name:    "Unknown_Pattern",
			pattern: "^[X-Y]{5}[#]{2}$",
			regex:   regexp.MustCompile(".+"), // should generate something
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generated, err := generator.generateFromPattern(tc.pattern, rng)
			if err != nil {
				t.Fatalf("Failed to generate pattern for %s: %v", tc.pattern, err)
			}

			if tc.expected != "" && generated != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, generated)
			}

			if tc.regex != nil && !tc.regex.MatchString(generated) {
				t.Errorf("Generated value '%s' does not match expected regex for pattern '%s'", 
					generated, tc.pattern)
			}

			// Property: generated strings should not be empty (unless pattern is empty)
			if tc.pattern != "" && generated == "" {
				t.Errorf("Generated empty string for non-empty pattern '%s'", tc.pattern)
			}
		})
	}
}

// TestPatternGeneration_Performance tests that pattern generation is efficient
func TestPatternGeneration_Performance(t *testing.T) {
	generator := NewDeterministicGenerator(12345)
	pattern := "^[A-Z]{2}[0-9]{6}$" // SKU pattern
	
	start := time.Now()
	iterations := 10000
	
	for i := 0; i < iterations; i++ {
		rng := rand.New(rand.NewSource(int64(i)))
		_, err := generator.generateFromPattern(pattern, rng)
		if err != nil {
			t.Fatalf("Pattern generation failed at iteration %d: %v", i, err)
		}
	}
	
	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)
	
	// Performance requirement: should generate patterns in < 1ms on average
	if avgDuration > time.Millisecond {
		t.Errorf("Pattern generation too slow: average %v per generation (expected < 1ms)", avgDuration)
	}
	
	t.Logf("Generated %d patterns in %v (avg: %v per pattern)", iterations, duration, avgDuration)
}

// BenchmarkPatternGeneration benchmarks the pattern generation performance
func BenchmarkPatternGeneration(b *testing.B) {
	generator := NewDeterministicGenerator(12345)
	patterns := []string{
		"^[A-Z]{2}[0-9]{6}$",      // SKU
		"^PRD[0-9]{8}$",           // Product ID
		"^WH[0-9]{3}$",            // Warehouse
		"^[0-9]{10}$",             // Account number
		"^[A-Z][0-9]{2}\\.[0-9]{1,2}$", // ICD-10
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		pattern := patterns[i%len(patterns)]
		rng := rand.New(rand.NewSource(int64(i)))
		_, err := generator.generateFromPattern(pattern, rng)
		if err != nil {
			b.Fatalf("Pattern generation failed: %v", err)
		}
	}
}
