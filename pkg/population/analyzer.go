package population

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PopulationAnalyzer analyzes business scenarios and suggests realistic data generation strategies
type PopulationAnalyzer struct {
	templates map[string]*PopulationTemplate
	llmClient LLMClient // For parsing complex scenarios
}

// PopulationTemplate defines realistic ratios and patterns for a business domain
type PopulationTemplate struct {
	Domain      string                    `json:"domain"`
	Description string                    `json:"description"`
	BaseMetrics map[string]MetricRatio    `json:"base_metrics"`
	Schemas     []SchemaRecommendation    `json:"schemas"`
	Relationships []RelationshipRule      `json:"relationships"`
}

// MetricRatio defines realistic ratios for different data types
type MetricRatio struct {
	Name        string  `json:"name"`
	Ratio       float64 `json:"ratio"`        // Records per base unit
	Distribution string `json:"distribution"` // normal, poisson, uniform
	MinValue    int     `json:"min_value"`
	MaxValue    int     `json:"max_value"`
	Description string  `json:"description"`
}

// SchemaRecommendation suggests appropriate schemas for the population
type SchemaRecommendation struct {
	SchemaPath   string   `json:"schema_path"`
	RecordType   string   `json:"record_type"`
	Priority     string   `json:"priority"` // critical, important, optional
	Dependencies []string `json:"dependencies"`
}

// RelationshipRule defines how different record types relate to each other
type RelationshipRule struct {
	ParentType   string  `json:"parent_type"`
	ChildType    string  `json:"child_type"`
	Relationship string  `json:"relationship"` // one-to-many, many-to-many
	Ratio        float64 `json:"ratio"`
	Description  string  `json:"description"`
}

// PopulationScenario represents a parsed business scenario
type PopulationScenario struct {
	Domain      string            `json:"domain"`
	BaseUnit    string            `json:"base_unit"`    // "beds", "branches", "stores"
	BaseCount   int               `json:"base_count"`   // 500, 12, 20
	Location    string            `json:"location"`     // "Chicago", "regional"
	Attributes  map[string]string `json:"attributes"`   // Additional context
	Template    *PopulationTemplate `json:"template"`
}

// GenerationStrategy provides a complete data generation plan
type GenerationStrategy struct {
	Scenario     *PopulationScenario      `json:"scenario"`
	RecordCounts map[string]int           `json:"record_counts"`
	Schemas      []SchemaRecommendation   `json:"schemas"`
	Dependencies []string                 `json:"dependencies"`
	Timeline     *GenerationTimeline      `json:"timeline"`
	Resources    *ResourceEstimate        `json:"resources"`
}

// GenerationTimeline estimates generation time and order
type GenerationTimeline struct {
	EstimatedDuration string   `json:"estimated_duration"`
	Phases           []Phase  `json:"phases"`
}

// Phase represents a generation phase with dependencies
type Phase struct {
	Name         string   `json:"name"`
	RecordTypes  []string `json:"record_types"`
	EstimatedTime string  `json:"estimated_time"`
	Dependencies []string `json:"dependencies"`
}

// ResourceEstimate calculates resource requirements
type ResourceEstimate struct {
	TotalRecords    int    `json:"total_records"`
	EstimatedSize   string `json:"estimated_size"`
	LLMCalls        int    `json:"llm_calls"`
	MemoryRequired  string `json:"memory_required"`
	RecommendedCPUs int    `json:"recommended_cpus"`
}

// LLMClient interface for parsing complex scenarios
type LLMClient interface {
	ParseScenario(ctx context.Context, description string) (*PopulationScenario, error)
}

// NewPopulationAnalyzer creates a new population analyzer with built-in templates
func NewPopulationAnalyzer(llmClient LLMClient) *PopulationAnalyzer {
	analyzer := &PopulationAnalyzer{
		templates: make(map[string]*PopulationTemplate),
		llmClient: llmClient,
	}
	
	// Load built-in templates
	analyzer.loadBuiltinTemplates()
	return analyzer
}

// AnalyzePopulation analyzes a business scenario and returns a generation strategy
func (pa *PopulationAnalyzer) AnalyzePopulation(ctx context.Context, description string) (*GenerationStrategy, error) {
	// Parse the scenario description
	scenario, err := pa.parseScenario(ctx, description)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scenario: %w", err)
	}
	
	// Find matching template
	template, err := pa.findTemplate(scenario.Domain)
	if err != nil {
		return nil, fmt.Errorf("no template found for domain %s: %w", scenario.Domain, err)
	}
	
	scenario.Template = template
	
	// Calculate realistic record counts
	recordCounts := pa.calculateRecordCounts(scenario)
	
	// Generate timeline and resource estimates
	timeline := pa.estimateTimeline(recordCounts)
	resources := pa.estimateResources(recordCounts)
	
	strategy := &GenerationStrategy{
		Scenario:     scenario,
		RecordCounts: recordCounts,
		Schemas:      template.Schemas,
		Dependencies: pa.calculateDependencies(template),
		Timeline:     timeline,
		Resources:    resources,
	}
	
	return strategy, nil
}

// parseScenario extracts key information from the scenario description
func (pa *PopulationAnalyzer) parseScenario(ctx context.Context, description string) (*PopulationScenario, error) {
	// Try pattern matching first for common formats
	scenario := pa.parseWithPatterns(description)
	if scenario != nil {
		return scenario, nil
	}
	
	// Fall back to LLM parsing for complex scenarios
	if pa.llmClient != nil {
		return pa.llmClient.ParseScenario(ctx, description)
	}
	
	return nil, fmt.Errorf("unable to parse scenario: %s", description)
}

// parseWithPatterns uses regex patterns to extract scenario information
func (pa *PopulationAnalyzer) parseWithPatterns(description string) *PopulationScenario {
	patterns := map[string]*regexp.Regexp{
		"hospital":  regexp.MustCompile(`(\d+)-bed\s+.*hospital`),
		"bank":      regexp.MustCompile(`bank.*with\s+(\d+)\s+branches`),
		"retail":    regexp.MustCompile(`(\d+)\s+stores?`),
		"ecommerce": regexp.MustCompile(`(\d+[KM]?)\s+.*users?`),
		"insurance": regexp.MustCompile(`(\d+[KM]?)\s+.*policyholders?`),
	}
	
	for domain, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(description); len(matches) > 1 {
			count, err := pa.parseCount(matches[1])
			if err != nil {
				continue
			}
			
			return &PopulationScenario{
				Domain:    domain,
				BaseUnit:  pa.getBaseUnit(domain),
				BaseCount: count,
				Location:  pa.extractLocation(description),
				Attributes: pa.extractAttributes(description),
			}
		}
	}
	
	return nil
}

// parseCount handles counts with K/M suffixes
func (pa *PopulationAnalyzer) parseCount(countStr string) (int, error) {
	countStr = strings.ToUpper(countStr)
	
	if strings.HasSuffix(countStr, "K") {
		base, err := strconv.Atoi(strings.TrimSuffix(countStr, "K"))
		if err != nil {
			return 0, err
		}
		return base * 1000, nil
	}
	
	if strings.HasSuffix(countStr, "M") {
		base, err := strconv.Atoi(strings.TrimSuffix(countStr, "M"))
		if err != nil {
			return 0, err
		}
		return base * 1000000, nil
	}
	
	return strconv.Atoi(countStr)
}

// getBaseUnit returns the base unit for a domain
func (pa *PopulationAnalyzer) getBaseUnit(domain string) string {
	units := map[string]string{
		"hospital":  "beds",
		"bank":      "branches",
		"retail":    "stores",
		"ecommerce": "users",
		"insurance": "policyholders",
	}
	return units[domain]
}

// extractLocation extracts location information from description
func (pa *PopulationAnalyzer) extractLocation(description string) string {
	// Simple location extraction - could be enhanced with NLP
	locations := []string{"Chicago", "New York", "Los Angeles", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose"}
	
	for _, location := range locations {
		if strings.Contains(strings.ToLower(description), strings.ToLower(location)) {
			return location
		}
	}
	
	if strings.Contains(strings.ToLower(description), "regional") {
		return "regional"
	}
	
	return "unknown"
}

// extractAttributes extracts additional attributes from description
func (pa *PopulationAnalyzer) extractAttributes(description string) map[string]string {
	attributes := make(map[string]string)
	
	// Extract common attributes
	if strings.Contains(strings.ToLower(description), "community") {
		attributes["type"] = "community"
	}
	if strings.Contains(strings.ToLower(description), "regional") {
		attributes["scale"] = "regional"
	}
	if strings.Contains(strings.ToLower(description), "academic") {
		attributes["type"] = "academic"
	}
	
	return attributes
}

// findTemplate finds the best matching template for a domain
func (pa *PopulationAnalyzer) findTemplate(domain string) (*PopulationTemplate, error) {
	template, exists := pa.templates[domain]
	if !exists {
		return nil, fmt.Errorf("no template found for domain: %s", domain)
	}
	return template, nil
}

// calculateRecordCounts calculates realistic record counts based on the scenario and template
func (pa *PopulationAnalyzer) calculateRecordCounts(scenario *PopulationScenario) map[string]int {
	counts := make(map[string]int)
	
	for metricName, metric := range scenario.Template.BaseMetrics {
		baseCount := float64(scenario.BaseCount)
		recordCount := int(baseCount * metric.Ratio)
		
		// Apply min/max constraints
		if recordCount < metric.MinValue {
			recordCount = metric.MinValue
		}
		if metric.MaxValue > 0 && recordCount > metric.MaxValue {
			recordCount = metric.MaxValue
		}
		
		counts[metricName] = recordCount
	}
	
	return counts
}

// estimateTimeline estimates generation timeline based on record counts
func (pa *PopulationAnalyzer) estimateTimeline(recordCounts map[string]int) *GenerationTimeline {
	totalRecords := 0
	for _, count := range recordCounts {
		totalRecords += count
	}
	
	// Rough estimation: 1000 records per second for deterministic, 10 records per second with LLM
	estimatedSeconds := totalRecords / 500 // Conservative estimate
	
	return &GenerationTimeline{
		EstimatedDuration: fmt.Sprintf("%d seconds", estimatedSeconds),
		Phases: []Phase{
			{
				Name:          "Core Data",
				RecordTypes:   []string{"patients", "providers", "facilities"},
				EstimatedTime: fmt.Sprintf("%d seconds", estimatedSeconds/3),
				Dependencies:  []string{},
			},
			{
				Name:          "Transactional Data",
				RecordTypes:   []string{"claims", "prescriptions", "procedures"},
				EstimatedTime: fmt.Sprintf("%d seconds", estimatedSeconds*2/3),
				Dependencies:  []string{"Core Data"},
			},
		},
	}
}

// estimateResources estimates resource requirements
func (pa *PopulationAnalyzer) estimateResources(recordCounts map[string]int) *ResourceEstimate {
	totalRecords := 0
	for _, count := range recordCounts {
		totalRecords += count
	}
	
	// Rough estimates
	estimatedSizeMB := totalRecords * 2 / 1000 // ~2KB per record average
	llmCalls := totalRecords / 5 // Assume 20% of records use LLM
	memoryMB := totalRecords / 1000 + 100 // Base memory + record overhead
	
	return &ResourceEstimate{
		TotalRecords:    totalRecords,
		EstimatedSize:   fmt.Sprintf("%d MB", estimatedSizeMB),
		LLMCalls:        llmCalls,
		MemoryRequired:  fmt.Sprintf("%d MB", memoryMB),
		RecommendedCPUs: 4,
	}
}

// calculateDependencies calculates schema dependencies
func (pa *PopulationAnalyzer) calculateDependencies(template *PopulationTemplate) []string {
	var deps []string
	for _, schema := range template.Schemas {
		deps = append(deps, schema.Dependencies...)
	}
	return deps
}

// loadBuiltinTemplates loads the built-in population templates
func (pa *PopulationAnalyzer) loadBuiltinTemplates() {
	// Load templates from templates.go
	pa.templates["hospital"] = GetHospitalTemplate()
	pa.templates["bank"] = GetBankTemplate()
	pa.templates["retail"] = GetRetailTemplate()
	pa.templates["ecommerce"] = GetEcommerceTemplate()
	pa.templates["insurance"] = GetInsuranceTemplate()
}
