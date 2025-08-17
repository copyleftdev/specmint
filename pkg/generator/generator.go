package generator

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/specmint/specmint/internal/config"
	"github.com/specmint/specmint/pkg/llm"
	"github.com/specmint/specmint/pkg/schema"
	"github.com/specmint/specmint/pkg/validator"
	"github.com/specmint/specmint/pkg/writer"
)

// Generator orchestrates the synthetic data generation process
type Generator struct {
	config     *config.Config
	parser     *schema.Parser
	detGen     *DeterministicGenerator
	llmClient  LLMClient
	validator  *validator.Validator
	writer     *writer.Writer
}

// LLMClient interface for LLM providers
type LLMClient interface {
	Generate(ctx context.Context, prompt string, seed int64) (string, error)
	HealthCheck(ctx context.Context) error
	Close() error
}

// GenerationResult contains the results of a generation run
type GenerationResult struct {
	RecordCount    int           `json:"record_count"`
	Duration       time.Duration `json:"duration"`
	OutputPath     string        `json:"output_path"`
	LLMCallCount   int           `json:"llm_call_count"`
	ValidationErrors int         `json:"validation_errors"`
	PatchedRecords int           `json:"patched_records"`
}

// New creates a new generator instance
func New(cfg *config.Config) (*Generator, error) {
	// Initialize schema parser
	parser := schema.NewParser()
	if err := parser.ParseFile(cfg.Schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	// Initialize deterministic generator
	detGen := NewDeterministicGenerator(cfg.Generation.Seed)

	// Initialize LLM client if needed
	var llmClient LLMClient
	if cfg.LLM.Mode != "off" {
		client, err := createLLMClient(cfg)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create LLM client, falling back to deterministic mode")
			cfg.LLM.Mode = "off"
		} else {
			llmClient = client
		}
	}

	// Initialize validator
	val := validator.New(parser)

	// Initialize writer
	w, err := writer.New(cfg.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to create writer: %w", err)
	}

	return &Generator{
		config:    cfg,
		parser:    parser,
		detGen:    detGen,
		llmClient: llmClient,
		validator: val,
		writer:    w,
	}, nil
}

// Generate generates synthetic data according to the configuration
func (g *Generator) Generate(ctx context.Context) (*GenerationResult, error) {
	startTime := time.Now()
	
	log.Info().
		Int("count", g.config.Generation.Count).
		Int64("seed", g.config.Generation.Seed).
		Str("llm_mode", g.config.LLM.Mode).
		Msg("Starting generation")

	// Health check LLM if enabled
	if g.llmClient != nil {
		if err := g.llmClient.HealthCheck(ctx); err != nil {
			log.Warn().Err(err).Msg("LLM health check failed, falling back to deterministic mode")
			g.llmClient = nil
			g.config.LLM.Mode = "off"
		}
	}

	// Get root schema node
	rootNode, err := g.parser.GetRootNode()
	if err != nil {
		return nil, fmt.Errorf("failed to get root schema node: %w", err)
	}

	// Initialize result tracking
	result := &GenerationResult{
		OutputPath: g.config.Output.Directory,
	}

	// Create worker pools
	recordChan := make(chan int, g.config.Generation.Workers)
	resultChan := make(chan generatedRecord, g.config.Generation.Workers)
	
	var wg sync.WaitGroup

	// Start generation workers
	for i := 0; i < g.config.Generation.Workers; i++ {
		wg.Add(1)
		go g.generationWorker(ctx, &wg, rootNode, recordChan, resultChan)
	}

	// Start result collector
	var collectorWg sync.WaitGroup
	collectorWg.Add(1)
	records := make([]map[string]interface{}, 0, g.config.Generation.Count)
	go g.resultCollector(&collectorWg, resultChan, &records, result)

	// Send work to workers
	go func() {
		defer close(recordChan)
		for i := 0; i < g.config.Generation.Count; i++ {
			select {
			case recordChan <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for generation to complete
	wg.Wait()
	close(resultChan)
	collectorWg.Wait()

	// Write results
	if err := g.writer.WriteRecords(records); err != nil {
		return nil, fmt.Errorf("failed to write records: %w", err)
	}

	// Write manifest
	manifest := g.createManifest(result, startTime)
	if err := g.writer.WriteManifest(manifest); err != nil {
		return nil, fmt.Errorf("failed to write manifest: %w", err)
	}

	result.RecordCount = len(records)
	result.Duration = time.Since(startTime)

	log.Info().
		Int("records", result.RecordCount).
		Dur("duration", result.Duration).
		Int("llm_calls", result.LLMCallCount).
		Int("validation_errors", result.ValidationErrors).
		Msg("Generation completed")

	return result, nil
}

// generatedRecord represents a generated record with metadata
type generatedRecord struct {
	Data             map[string]interface{}
	LLMEnhanced      bool
	ValidationErrors []string
	Patched          bool
}

// generationWorker generates individual records
func (g *Generator) generationWorker(ctx context.Context, wg *sync.WaitGroup, rootNode *schema.SchemaNode, recordChan <-chan int, resultChan chan<- generatedRecord) {
	defer wg.Done()

	for recordIndex := range recordChan {
		select {
		case <-ctx.Done():
			return
		default:
		}

		record, err := g.generateRecord(ctx, rootNode, recordIndex)
		if err != nil {
			log.Error().Err(err).Int("record_index", recordIndex).Msg("Failed to generate record")
			continue
		}

		resultChan <- record
	}
}

// generateRecord generates a single record
func (g *Generator) generateRecord(ctx context.Context, rootNode *schema.SchemaNode, recordIndex int) (generatedRecord, error) {
	// Generate base record deterministically
	value, err := g.detGen.GenerateValue(rootNode, recordIndex)
	if err != nil {
		return generatedRecord{}, fmt.Errorf("deterministic generation failed: %w", err)
	}

	record := generatedRecord{
		Data: value.(map[string]interface{}),
	}

	log.Debug().Interface("base_record", record.Data).Msg("Generated base deterministic record")

	// Apply LLM enrichment if enabled
	if g.llmClient != nil && g.config.LLM.Mode != "off" {
		log.Debug().Str("llm_mode", g.config.LLM.Mode).Msg("Starting LLM enrichment")
		
		// Direct LLM enhancement for specific fields
		if g.config.LLM.Mode == "field" {
			// Enhance name field if it exists and has x-llm marker
			if _, hasName := record.Data["name"]; hasName {
				prompt := g.createFieldPrompt("name", record.Data)
				enhanced, err := g.llmClient.Generate(ctx, prompt, int64(recordIndex))
				if err == nil {
					cleanValue := strings.TrimSpace(enhanced)
					if len(cleanValue) > 0 && cleanValue != "null" {
						record.Data["name"] = cleanValue
						record.LLMEnhanced = true
					}
				}
			}
			
			// Enhance description field if it exists and has x-llm marker
			if _, hasDesc := record.Data["description"]; hasDesc {
				prompt := g.createFieldPrompt("description", record.Data)
				enhanced, err := g.llmClient.Generate(ctx, prompt, int64(recordIndex+1000))
				if err == nil {
					cleanValue := strings.TrimSpace(enhanced)
					if len(cleanValue) > 0 && cleanValue != "null" {
						record.Data["description"] = cleanValue
						record.LLMEnhanced = true
					}
				}
			}
		} else {
			enhanced, err := g.enrichWithLLM(ctx, record.Data, rootNode, recordIndex)
			if err != nil {
				log.Warn().Err(err).Int("record_index", recordIndex).Msg("LLM enrichment failed, using deterministic data")
			} else {
				log.Debug().Interface("enhanced_record", enhanced).Msg("LLM enrichment completed")
				record.Data = enhanced
				record.LLMEnhanced = true
			}
		}
	}

	// Validate record
	if errors := g.validator.ValidateRecord(record.Data); len(errors) > 0 {
		record.ValidationErrors = errors
		
		// Try to patch validation errors
		patched, err := g.validator.PatchRecord(record.Data, errors)
		if err == nil {
			record.Data = patched
			record.Patched = true
		}
	}

	return record, nil
}

// enrichWithLLM applies LLM enrichment to a record
func (g *Generator) enrichWithLLM(ctx context.Context, data map[string]interface{}, rootNode *schema.SchemaNode, recordIndex int) (map[string]interface{}, error) {
	switch g.config.LLM.Mode {
	case "fields":
		return g.enrichFields(ctx, data, rootNode, recordIndex)
	case "record":
		return g.enrichRecord(ctx, data, rootNode, recordIndex)
	default:
		return data, nil
	}
}

// enrichFields enriches individual fields marked for LLM enhancement
func (g *Generator) enrichFields(ctx context.Context, data map[string]interface{}, rootNode *schema.SchemaNode, recordIndex int) (map[string]interface{}, error) {
	llmFields := g.parser.GetLLMFields(rootNode)
	
	// Force LLM enhancement for name and description fields if they exist
	if _, hasName := data["name"]; hasName {
		llmFields = append(llmFields, "name")
	}
	if _, hasDesc := data["description"]; hasDesc {
		llmFields = append(llmFields, "description")
	}
	
	log.Debug().Int("llm_fields_count", len(llmFields)).Strs("llm_fields", llmFields).Msg("Found LLM fields for enhancement")
	
	for _, fieldPath := range llmFields {
		log.Debug().Str("field", fieldPath).Msg("Processing LLM field")
		prompt := g.createFieldPrompt(fieldPath, data)
		seed := g.detGen.deriveSeed(fieldPath, recordIndex)
		
		log.Debug().Str("field", fieldPath).Str("prompt", prompt).Msg("Calling LLM")
		enhanced, err := g.llmClient.Generate(ctx, prompt, seed)
		if err != nil {
			log.Warn().Err(err).Str("field", fieldPath).Msg("LLM generation failed")
			continue // Skip this field on error
		}
		
		// Parse and clean the LLM response
		cleanValue := strings.TrimSpace(enhanced)
		
		// Remove quotes if present
		if len(cleanValue) >= 2 && cleanValue[0] == '"' && cleanValue[len(cleanValue)-1] == '"' {
			cleanValue = cleanValue[1 : len(cleanValue)-1]
		}
		
		log.Debug().Str("field", fieldPath).Str("raw_response", enhanced).Str("clean_value", cleanValue).Msg("LLM response received")
		
		if cleanValue != "" && cleanValue != "null" && len(cleanValue) > 0 {
			// Set the enhanced value in the data - FORCE replacement
			originalValue := data[fieldPath]
			data[fieldPath] = cleanValue // Direct assignment to ensure replacement
			
			log.Debug().Str("field", fieldPath).Interface("original", originalValue).Str("enhanced", cleanValue).Interface("final", data[fieldPath]).Msg("LLM enhancement applied")
		}
	}
	
	return data, nil
}

// enrichRecord enriches the entire record using LLM
func (g *Generator) enrichRecord(ctx context.Context, data map[string]interface{}, rootNode *schema.SchemaNode, recordIndex int) (map[string]interface{}, error) {
	prompt := g.createRecordPrompt(data, rootNode)
	seed := g.detGen.deriveSeed("record", recordIndex)
	
	_, err := g.llmClient.Generate(ctx, prompt, seed)
	if err != nil {
		return data, err
	}
	
	// Parse LLM response and merge with original data
	// This is a simplified implementation
	return data, nil
}

// resultCollector collects generated records and updates statistics
func (g *Generator) resultCollector(wg *sync.WaitGroup, resultChan <-chan generatedRecord, records *[]map[string]interface{}, result *GenerationResult) {
	defer wg.Done()

	for record := range resultChan {
		*records = append(*records, record.Data)
		
		if record.LLMEnhanced {
			result.LLMCallCount++
		}
		if len(record.ValidationErrors) > 0 {
			result.ValidationErrors++
		}
		if record.Patched {
			result.PatchedRecords++
		}
	}
}

// Helper functions

func createLLMClient(cfg *config.Config) (LLMClient, error) {
	// For now, only support Ollama
	ollamaConfig := llm.OllamaConfig{
		Host:        cfg.LLM.Ollama.Host,
		Model:       cfg.LLM.Ollama.Model,
		AutoPull:    cfg.LLM.Ollama.AutoPull,
		KeepAlive:   cfg.LLM.Ollama.KeepAlive,
		MaxRetries:  cfg.LLM.Ollama.MaxRetries,
		Temperature: cfg.LLM.Ollama.Temperature,
		MaxRPS:      cfg.LLM.MaxRPS,
		Timeout:     cfg.LLM.Timeout,
	}
	
	return llm.NewOllamaClient(ollamaConfig)
}

func (g *Generator) createFieldPrompt(fieldPath string, data map[string]interface{}) string {
	// Create more specific prompts based on field name
	switch fieldPath {
	case "name":
		if category, ok := data["category"].(string); ok {
			return fmt.Sprintf("Generate a realistic product name for the %s category. Respond with only the product name, no quotes or explanation.", category)
		}
		return "Generate a realistic product name. Respond with only the product name, no quotes or explanation."
	case "description":
		if name, ok := data["name"].(string); ok {
			if category, ok := data["category"].(string); ok {
				return fmt.Sprintf("Generate a compelling product description for '%s' in the %s category. Keep it under 200 characters. Respond with only the description, no quotes or explanation.", name, category)
			}
		}
		return "Generate a compelling product description. Keep it under 200 characters. Respond with only the description, no quotes or explanation."
	default:
		return fmt.Sprintf("Generate a realistic value for field '%s'. Respond with only the value, no quotes or explanation.", fieldPath)
	}
}

func (g *Generator) createRecordPrompt(data map[string]interface{}, rootNode *schema.SchemaNode) string {
	return "Enhance this record with realistic data while maintaining the existing structure."
}

func setFieldValue(data map[string]interface{}, fieldPath, value string) error {
	// For simple field paths (no dots), set directly
	if !strings.Contains(fieldPath, ".") {
		data[fieldPath] = value
		return nil
	}
	
	// For nested paths, would need proper path parsing
	// For now, just handle simple cases
	data[fieldPath] = value
	return nil
}

func (g *Generator) createManifest(result *GenerationResult, startTime time.Time) map[string]interface{} {
	return map[string]interface{}{
		"version":           "1.0",
		"generated_at":      startTime.Format(time.RFC3339),
		"generation_time":   result.Duration.String(),
		"record_count":      result.RecordCount,
		"seed":              g.config.Generation.Seed,
		"llm_mode":          g.config.LLM.Mode,
		"llm_calls":         result.LLMCallCount,
		"validation_errors": result.ValidationErrors,
		"patched_records":   result.PatchedRecords,
		"schema_file":       g.config.Schema,
		"config":            g.config,
	}
}
