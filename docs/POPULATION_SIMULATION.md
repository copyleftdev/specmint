# Population-Based Simulation Architecture

## Overview

SpecMint's population-based simulation system transforms natural language business scenarios into comprehensive synthetic datasets. This document describes the architecture, algorithms, and design decisions behind this intelligent data generation system.

## Core Components

### 1. Population Analyzer (`pkg/population/analyzer.go`)

The `PopulationAnalyzer` is the central intelligence component that:

- **Parses Business Scenarios**: Extracts domain, scale, location, and attributes from natural language
- **Selects Domain Templates**: Matches scenarios to appropriate business domain templates
- **Calculates Record Counts**: Applies realistic scaling algorithms based on business size
- **Estimates Resources**: Predicts generation time, memory usage, and LLM calls
- **Generates Strategy**: Creates complete execution plan with dependencies

#### Key Methods

```go
func (pa *PopulationAnalyzer) AnalyzePopulation(ctx context.Context, description string) (*GenerationStrategy, error)
func (pa *PopulationAnalyzer) parseScenario(description string) (*PopulationScenario, error)
func (pa *PopulationAnalyzer) calculateRecordCounts(template *PopulationTemplate, baseCount int) map[string]int
func (pa *PopulationAnalyzer) estimateTimeline(recordCounts map[string]int) *GenerationTimeline
```

### 2. Domain Templates (`pkg/population/templates.go`)

Domain-specific templates encode real-world business knowledge:

#### Hospital Template
```go
BaseMetrics: map[string]float64{
    "patients_per_bed":      5.0,   // 5 patients per bed capacity
    "providers_per_bed":     0.2,   // 1 provider per 5 beds
    "claims_per_patient":    1.5,   // 1.5 claims per patient
    "prescriptions_per_patient": 2.4, // 2.4 prescriptions per patient
}
```

#### Bank Template
```go
BaseMetrics: map[string]float64{
    "customers_per_branch":     500,   // 500 customers per branch
    "accounts_per_customer":    1.5,   // 1.5 accounts per customer
    "transactions_per_account": 20,    // 20 transactions per account per month
    "loans_per_branch":         5,     // 5 loans per branch
}
```

### 3. CLI Integration (`cmd/specmint/simulate.go`)

The `simulate` command provides three operational modes:

1. **Analysis Mode**: `--population "description"` (analysis only)
2. **Save Strategy**: `--save-strategy file.json` (save for later execution)
3. **Execute Mode**: `--execute --output ./data` (full generation)

## Algorithms

### Scenario Parsing Algorithm

```go
func parseScenario(description string) (*PopulationScenario, error) {
    // 1. Extract domain keywords (hospital, bank, retail, etc.)
    domain := extractDomain(description)
    
    // 2. Extract base unit and count (e.g., "25 beds", "5 branches")
    baseUnit, count := extractBaseUnit(description)
    
    // 3. Extract location context (optional)
    location := extractLocation(description)
    
    // 4. Extract additional attributes
    attributes := extractAttributes(description)
    
    return &PopulationScenario{
        Domain:     domain,
        BaseUnit:   baseUnit,
        Count:      count,
        Location:   location,
        Attributes: attributes,
    }
}
```

### Record Count Calculation

```go
func calculateRecordCounts(template *PopulationTemplate, baseCount int) map[string]int {
    counts := make(map[string]int)
    
    for recordType, ratio := range template.BaseMetrics {
        // Apply scaling factor based on business size
        scaledRatio := ratio * getScalingFactor(baseCount)
        
        // Calculate record count with realistic variance
        baseRecords := float64(baseCount) * scaledRatio
        variance := baseRecords * 0.1 // 10% variance
        
        // Add realistic randomization
        finalCount := int(baseRecords + (rand.Float64()-0.5)*variance)
        counts[recordType] = max(1, finalCount)
    }
    
    return counts
}
```

### Resource Estimation

```go
func estimateResources(recordCounts map[string]int) *ResourceEstimate {
    totalRecords := sumRecords(recordCounts)
    
    return &ResourceEstimate{
        TotalRecords:    totalRecords,
        EstimatedSizeMB: totalRecords * avgRecordSizeKB / 1024,
        LLMCalls:        totalRecords * llmFieldsPerRecord,
        MemoryRequiredMB: calculateMemoryRequirement(totalRecords),
        RecommendedCPUs: calculateOptimalCPUs(totalRecords),
        EstimatedDuration: calculateDuration(totalRecords),
    }
}
```

## Domain Knowledge Encoding

### Healthcare Domain

**Business Logic**:
- Hospitals scale by bed count
- Patient volume correlates with bed capacity
- Provider-to-patient ratios follow industry standards
- Claims generation follows realistic medical patterns

**Realistic Ratios**:
- 5 patients per bed (annual capacity)
- 1 provider per 5 beds
- 1.5 claims per patient
- 2.4 prescriptions per patient
- 0.5 procedures per patient
- 4 lab results per patient

### Banking Domain

**Business Logic**:
- Banks scale by branch count
- Customer acquisition follows geographic patterns
- Account products have realistic adoption rates
- Transaction volumes reflect real banking behavior

**Realistic Ratios**:
- 500 customers per branch
- 1.5 accounts per customer
- 20 transactions per account monthly
- 5 loans per branch
- 10 credit cards per branch

### E-commerce Domain

**Business Logic**:
- Platforms scale by user count
- Product catalogs grow with user base
- Order patterns follow e-commerce metrics
- Review rates match industry standards

**Realistic Ratios**:
- 0.1 products per user (catalog scaling)
- 0.5 orders per user monthly
- 0.25 reviews per user
- 2.5 cart sessions per user

## Performance Optimizations

### Quick Test Mode

For development and demonstration, record count ratios are reduced:

```go
// Production ratios
"patients_per_bed": 5.0

// Quick test ratios (10x reduction)
"patients_per_bed": 0.5
```

This enables:
- Fast iteration during development
- Quick demonstrations
- Reduced resource consumption
- Faster CI/CD pipeline testing

### Memory Management

- **Streaming Generation**: Records generated and written incrementally
- **Connection Pooling**: Reuse HTTP connections for LLM calls
- **Batch Processing**: Group similar operations for efficiency
- **Resource Monitoring**: Track memory usage and adjust batch sizes

### LLM Optimization

- **Circuit Breaker**: Prevent cascade failures during LLM unavailability
- **Fallback Strategy**: Graceful degradation to deterministic generation
- **Rate Limiting**: Respect LLM service limits
- **Caching**: Reuse common prompts and responses

## Integration Points

### Generator Integration

The simulation system integrates with SpecMint's existing generator:

```go
func executeGenerationStrategy(strategy *GenerationStrategy, outputDir string) error {
    for _, phase := range strategy.Phases {
        for _, recordType := range phase.RecordTypes {
            // Call existing generator with calculated parameters
            err := generator.Generate(GenerateConfig{
                Schema:    recordType.SchemaPath,
                Count:     recordType.Count,
                Output:    filepath.Join(outputDir, recordType.Name),
                Seed:      strategy.Seed,
                LLMMode:   "fields",
            })
        }
    }
}
```

### Schema Discovery

The system automatically discovers and validates schemas:

```go
func findSchemaPath(domain, recordType string) string {
    searchPaths := []string{
        fmt.Sprintf("test/schemas/%s/%s.json", domain, recordType),
        fmt.Sprintf("schemas/%s/%s.json", domain, recordType),
        fmt.Sprintf("test/schemas/%s/%s-*.json", domain, recordType),
    }
    
    for _, path := range searchPaths {
        if fileExists(path) {
            return path
        }
    }
    
    // Create placeholder if not found
    return createPlaceholderSchema(domain, recordType)
}
```

## Future Enhancements

### Advanced Scenario Parsing

- **LLM-Powered Parsing**: Use local LLM for complex scenario understanding
- **Multi-Domain Scenarios**: Support mixed business contexts
- **Temporal Modeling**: Add time-based growth patterns
- **Geographic Scaling**: Location-based adjustment factors

### Enhanced Templates

- **Industry Variants**: Specialized templates for sub-domains
- **Regulatory Compliance**: Built-in compliance rule enforcement
- **Custom Templates**: User-defined domain templates
- **Template Versioning**: Evolution and backward compatibility

### Intelligent Optimization

- **Adaptive Scaling**: Machine learning-based ratio optimization
- **Performance Prediction**: Accurate resource estimation
- **Quality Metrics**: Automated realism scoring
- **Feedback Loop**: Learn from generated dataset quality

## Testing Strategy

### Unit Tests
- Template validation and loading
- Scenario parsing accuracy
- Record count calculation algorithms
- Resource estimation accuracy

### Integration Tests
- End-to-end simulation execution
- Schema discovery and fallback
- Generator integration
- Output validation

### Performance Tests
- Large-scale simulation benchmarks
- Memory usage profiling
- LLM integration performance
- Concurrent execution testing

### Domain Tests
- Business logic validation
- Realistic ratio verification
- Cross-field relationship testing
- Industry standard compliance
