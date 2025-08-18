package population

import (
	"context"
	"testing"
)

func TestPopulationAnalyzer_AnalyzePopulation(t *testing.T) {
	analyzer := NewPopulationAnalyzer(nil)
	ctx := context.Background()

	tests := []struct {
		name        string
		description string
		wantDomain  string
		wantError   bool
	}{
		{
			name:        "hospital scenario",
			description: "100-bed regional hospital",
			wantDomain:  "hospital",
			wantError:   false,
		},
		{
			name:        "bank scenario",
			description: "community bank with 5 branches",
			wantDomain:  "bank",
			wantError:   false,
		},
		{
			name:        "retail scenario",
			description: "retail chain with 10 stores",
			wantDomain:  "retail",
			wantError:   false,
		},
		{
			name:        "ecommerce scenario",
			description: "e-commerce platform with 50K users",
			wantDomain:  "ecommerce",
			wantError:   false,
		},
		{
			name:        "insurance scenario",
			description: "insurance company with 25K policyholders",
			wantDomain:  "insurance",
			wantError:   false,
		},
		{
			name:        "empty description",
			description: "",
			wantError:   true,
		},
		{
			name:        "unknown domain",
			description: "unknown business type",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := analyzer.AnalyzePopulation(ctx, tt.description)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("AnalyzePopulation() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("AnalyzePopulation() unexpected error: %v", err)
				return
			}
			
			if strategy == nil {
				t.Errorf("AnalyzePopulation() returned nil strategy")
				return
			}
			
			if strategy.Scenario.Domain != tt.wantDomain {
				t.Errorf("AnalyzePopulation() domain = %v, want %v", strategy.Scenario.Domain, tt.wantDomain)
			}
			
			// Verify strategy has required fields
			if len(strategy.RecordCounts) == 0 {
				t.Errorf("AnalyzePopulation() no record counts generated")
			}
			
			if len(strategy.Schemas) == 0 {
				t.Errorf("AnalyzePopulation() no schemas generated")
			}
			
			if strategy.Timeline == nil {
				t.Errorf("AnalyzePopulation() no timeline generated")
			}
			
			if strategy.Resources == nil {
				t.Errorf("AnalyzePopulation() no resources generated")
			}
		})
	}
}

func TestPopulationAnalyzer_parseScenario(t *testing.T) {
	analyzer := NewPopulationAnalyzer(nil)
	ctx := context.Background()

	tests := []struct {
		name            string
		description     string
		wantDomain      string
		wantBaseUnit    string
		wantBaseCount   int
		wantLocation    string
		wantError       bool
	}{
		{
			name:          "hospital with location",
			description:   "500-bed regional hospital in Chicago",
			wantDomain:    "hospital",
			wantBaseUnit:  "beds",
			wantBaseCount: 500,
			wantLocation:  "Chicago",
			wantError:     false,
		},
		{
			name:          "bank with branches",
			description:   "community bank with 12 branches",
			wantDomain:    "bank",
			wantBaseUnit:  "branches",
			wantBaseCount: 12,
			wantLocation:  "unknown",
			wantError:     false,
		},
		{
			name:          "retail stores",
			description:   "retail chain with 20 stores",
			wantDomain:    "retail",
			wantBaseUnit:  "stores",
			wantBaseCount: 20,
			wantLocation:  "unknown",
			wantError:     false,
		},
		{
			name:          "ecommerce users",
			description:   "e-commerce platform with 100K users",
			wantDomain:    "ecommerce",
			wantBaseUnit:  "users",
			wantBaseCount: 100000,
			wantLocation:  "unknown",
			wantError:     false,
		},
		{
			name:          "insurance policyholders",
			description:   "insurance company with 25K policyholders",
			wantDomain:    "insurance",
			wantBaseUnit:  "policyholders",
			wantBaseCount: 25000,
			wantLocation:  "unknown",
			wantError:     false,
		},
		{
			name:        "no numbers",
			description: "hospital without numbers",
			wantError:   true,
		},
		{
			name:        "unknown domain",
			description: "unknown business with 10 units",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenario, err := analyzer.parseScenario(ctx, tt.description)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("parseScenario() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("parseScenario() unexpected error: %v", err)
				return
			}
			
			if scenario.Domain != tt.wantDomain {
				t.Errorf("parseScenario() domain = %v, want %v", scenario.Domain, tt.wantDomain)
			}
			
			if scenario.BaseUnit != tt.wantBaseUnit {
				t.Errorf("parseScenario() baseUnit = %v, want %v", scenario.BaseUnit, tt.wantBaseUnit)
			}
			
			if scenario.BaseCount != tt.wantBaseCount {
				t.Errorf("parseScenario() baseCount = %v, want %v", scenario.BaseCount, tt.wantBaseCount)
			}
			
			if scenario.Location != tt.wantLocation {
				t.Errorf("parseScenario() location = %v, want %v", scenario.Location, tt.wantLocation)
			}
		})
	}
}

func TestPopulationAnalyzer_calculateRecordCounts(t *testing.T) {
	analyzer := NewPopulationAnalyzer(nil)
	ctx := context.Background()

	// Create test scenarios
	smallScenario, err := analyzer.parseScenario(ctx, "10-bed hospital")
	if err != nil {
		t.Fatalf("Failed to parse small scenario: %v", err)
	}
	
	largeScenario, err := analyzer.parseScenario(ctx, "1000-bed hospital")
	if err != nil {
		t.Fatalf("Failed to parse large scenario: %v", err)
	}

	// Find hospital template
	template, err := analyzer.findTemplate("hospital")
	if err != nil {
		t.Fatalf("Failed to find hospital template: %v", err)
	}
	
	smallScenario.Template = template
	largeScenario.Template = template

	// Calculate record counts
	smallCounts := analyzer.calculateRecordCounts(smallScenario)
	largeCounts := analyzer.calculateRecordCounts(largeScenario)

	// Verify all expected record types are present
	expectedTypes := []string{"patients", "providers", "claims", "prescriptions", "procedures", "lab_results"}
	for _, recordType := range expectedTypes {
		if count, exists := smallCounts[recordType]; !exists {
			t.Errorf("calculateRecordCounts() missing record type: %s", recordType)
		} else if count <= 0 {
			t.Errorf("calculateRecordCounts() invalid count for %s: %d", recordType, count)
		}
	}

	// Verify scaling relationships
	for recordType := range smallCounts {
		if smallCounts[recordType] >= largeCounts[recordType] {
			t.Errorf("calculateRecordCounts() scaling issue for %s: small=%d, large=%d", 
				recordType, smallCounts[recordType], largeCounts[recordType])
		}
	}
}

func TestPopulationAnalyzer_estimateTimeline(t *testing.T) {
	analyzer := NewPopulationAnalyzer(nil)

	recordCounts := map[string]int{
		"patients":      100,
		"providers":     20,
		"claims":        150,
		"prescriptions": 240,
		"procedures":    50,
		"lab_results":   400,
	}

	timeline := analyzer.estimateTimeline(recordCounts)

	if timeline == nil {
		t.Errorf("estimateTimeline() returned nil")
		return
	}

	if timeline.EstimatedDuration == "" {
		t.Errorf("estimateTimeline() missing estimated duration")
	}

	if len(timeline.Phases) == 0 {
		t.Errorf("estimateTimeline() no phases generated")
	}

	// Verify phases have required fields
	for i, phase := range timeline.Phases {
		if phase.Name == "" {
			t.Errorf("estimateTimeline() phase %d missing name", i)
		}
		if phase.EstimatedTime == "" {
			t.Errorf("estimateTimeline() phase %d missing estimated time", i)
		}
		if len(phase.RecordTypes) == 0 {
			t.Errorf("estimateTimeline() phase %d has no record types", i)
		}
	}
}

func TestPopulationAnalyzer_estimateResources(t *testing.T) {
	analyzer := NewPopulationAnalyzer(nil)

	recordCounts := map[string]int{
		"patients":      100,
		"providers":     20,
		"claims":        150,
		"prescriptions": 240,
		"procedures":    50,
		"lab_results":   400,
	}

	resources := analyzer.estimateResources(recordCounts)

	if resources == nil {
		t.Errorf("estimateResources() returned nil")
		return
	}

	expectedTotal := 100 + 20 + 150 + 240 + 50 + 400
	if resources.TotalRecords != expectedTotal {
		t.Errorf("estimateResources() totalRecords = %d, want %d", resources.TotalRecords, expectedTotal)
	}

	if resources.EstimatedSize == "" {
		t.Errorf("estimateResources() missing estimated size")
	}

	if resources.LLMCalls <= 0 {
		t.Errorf("estimateResources() invalid LLM calls: %d", resources.LLMCalls)
	}

	if resources.MemoryRequired == "" {
		t.Errorf("estimateResources() missing memory required")
	}

	if resources.RecommendedCPUs <= 0 {
		t.Errorf("estimateResources() invalid CPU count: %d", resources.RecommendedCPUs)
	}
}

func TestPopulationAnalyzer_SchemaRecommendations(t *testing.T) {
	analyzer := NewPopulationAnalyzer(nil)
	ctx := context.Background()
	
	// Test full analysis to get schema recommendations
	strategy, err := analyzer.AnalyzePopulation(ctx, "100-bed hospital")
	if err != nil {
		t.Fatalf("AnalyzePopulation() failed: %v", err)
	}

	if len(strategy.Schemas) == 0 {
		t.Errorf("AnalyzePopulation() no schemas generated")
		return
	}

	// Verify schema fields
	for _, schema := range strategy.Schemas {
		if schema.RecordType == "" {
			t.Errorf("Schema missing record type")
		}
		if schema.SchemaPath == "" {
			t.Errorf("Schema missing path")
		}
		if schema.Priority == "" {
			t.Errorf("Schema missing priority")
		}
	}
}

// Benchmark tests
func BenchmarkAnalyzePopulation(b *testing.B) {
	analyzer := NewPopulationAnalyzer(nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzePopulation(ctx, "100-bed regional hospital")
		if err != nil {
			b.Fatalf("AnalyzePopulation() error: %v", err)
		}
	}
}

func BenchmarkParseScenario(b *testing.B) {
	analyzer := NewPopulationAnalyzer(nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.parseScenario(ctx, "500-bed regional hospital in Chicago")
		if err != nil {
			b.Fatalf("parseScenario() error: %v", err)
		}
	}
}

func BenchmarkCalculateRecordCounts(b *testing.B) {
	analyzer := NewPopulationAnalyzer(nil)
	ctx := context.Background()
	
	scenario, err := analyzer.parseScenario(ctx, "100-bed hospital")
	if err != nil {
		b.Fatalf("Failed to parse scenario: %v", err)
	}
	
	template, err := analyzer.findTemplate("hospital")
	if err != nil {
		b.Fatalf("Failed to find template: %v", err)
	}
	scenario.Template = template

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.calculateRecordCounts(scenario)
	}
}

// Test edge cases
func TestPopulationAnalyzer_EdgeCases(t *testing.T) {
	analyzer := NewPopulationAnalyzer(nil)
	ctx := context.Background()

	// Test very small numbers
	strategy, err := analyzer.AnalyzePopulation(ctx, "1-bed hospital")
	if err != nil {
		t.Errorf("AnalyzePopulation() failed for small number: %v", err)
	}
	if strategy != nil {
		for recordType, count := range strategy.RecordCounts {
			if count <= 0 {
				t.Errorf("AnalyzePopulation() invalid count for %s: %d", recordType, count)
			}
		}
	}

	// Test very large numbers
	strategy, err = analyzer.AnalyzePopulation(ctx, "10000-bed hospital")
	if err != nil {
		t.Errorf("AnalyzePopulation() failed for large number: %v", err)
	}
	if strategy != nil && strategy.Resources.TotalRecords <= 0 {
		t.Errorf("AnalyzePopulation() invalid total records for large scenario")
	}

	// Test different number formats
	testCases := []string{
		"5K users e-commerce platform",
		"50K users e-commerce platform", 
		"100K users e-commerce platform",
		"1M users e-commerce platform",
	}

	for _, testCase := range testCases {
		_, err := analyzer.AnalyzePopulation(ctx, testCase)
		if err != nil {
			t.Errorf("AnalyzePopulation() failed for %s: %v", testCase, err)
		}
	}
}
