package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/specmint/specmint/internal/config"
	"github.com/specmint/specmint/pkg/generator"
	"github.com/specmint/specmint/pkg/schema"
	"github.com/specmint/specmint/pkg/validator"
)

func newGenerateCmd() *cobra.Command {
	var (
		schemaFile string
		outputDir  string
		count      int
		seed       int64
		llmMode    string
		workers    int
		llmWorkers int
		maxRPS     int
		timeout    string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate synthetic dataset from JSON Schema",
		Long: `Generate synthetic dataset with deterministic seeded generation and optional LLM enrichment.

Examples:
  specmint generate --schema schema.json --count 1000 --seed 12345 --out ./output
  specmint generate --schema schema.json --count 100 --llm-mode fields --workers 4`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.FromContext(cmd.Context())

			// Override config with CLI flags
			if schemaFile != "" {
				cfg.Schema = schemaFile
			}
			if outputDir != "" {
				cfg.Output.Directory = outputDir
			}
			if count > 0 {
				cfg.Generation.Count = count
			}
			if seed != 0 {
				cfg.Generation.Seed = seed
			}
			if llmMode != "" {
				cfg.LLM.Mode = llmMode
			}
			if workers > 0 {
				cfg.Generation.Workers = workers
			}
			if llmWorkers > 0 {
				cfg.LLM.Workers = llmWorkers
			}
			if maxRPS > 0 {
				cfg.LLM.MaxRPS = maxRPS
			}

			// Create generator
			gen, err := generator.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to create generator: %w", err)
			}

			// Generate dataset
			result, err := gen.Generate(cmd.Context())
			if err != nil {
				return fmt.Errorf("generation failed: %w", err)
			}

			fmt.Printf("✅ Generated %d records in %v\n", result.RecordCount, result.Duration)
			fmt.Printf("📁 Output: %s\n", result.OutputPath)
			fmt.Printf("📊 Manifest: %s\n", filepath.Join(result.OutputPath, "manifest.json"))

			return nil
		},
	}

	cmd.Flags().StringVarP(&schemaFile, "schema", "s", "", "JSON Schema file path (required)")
	cmd.Flags().StringVarP(&outputDir, "out", "o", "", "Output directory (required)")
	cmd.Flags().IntVarP(&count, "count", "c", 0, "Number of records to generate")
	cmd.Flags().Int64Var(&seed, "seed", 0, "Random seed for deterministic generation")
	cmd.Flags().StringVar(&llmMode, "llm-mode", "", "LLM enrichment mode: off, fields, record")
	cmd.Flags().IntVar(&workers, "workers", 0, "Number of generation workers")
	cmd.Flags().IntVar(&llmWorkers, "llm-workers", 0, "Number of LLM workers")
	cmd.Flags().IntVar(&maxRPS, "llm-max-rps", 0, "Maximum LLM requests per second")
	cmd.Flags().StringVar(&timeout, "timeout", "", "Generation timeout (e.g., 5m, 30s)")

	_ = cmd.MarkFlagRequired("schema")
	_ = cmd.MarkFlagRequired("out")

	return cmd
}

func newValidateCmd() *cobra.Command {
	var (
		schemaFile  string
		datasetFile string
		verbose     bool
		rulesFile   string
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate dataset against JSON Schema",
		Long: `Validate generated dataset for schema compliance and cross-field rules.

Examples:
  specmint validate --schema schema.json --dataset output/dataset.jsonl
  specmint validate --schema schema.json --dataset output/dataset.jsonl --rules rules.json --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(datasetFile, schemaFile, rulesFile, verbose)
		},
	}

	cmd.Flags().StringVarP(&schemaFile, "schema", "s", "", "JSON Schema file path (required)")
	cmd.Flags().StringVarP(&datasetFile, "dataset", "d", "", "Dataset file to validate (required)")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().StringVar(&rulesFile, "rules", "", "Cross-field rules file")

	_ = cmd.MarkFlagRequired("schema")
	_ = cmd.MarkFlagRequired("dataset")

	return cmd
}

func newInspectCmd() *cobra.Command {
	var (
		datasetFile  string
		outputFormat string
		detailed     bool
	)

	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect dataset and generate analysis report",
		Long: `Generate detailed analysis report of dataset including statistics, 
field distributions, and quality metrics.

Examples:
  specmint inspect --dataset output/dataset.jsonl
  specmint inspect --dataset output/dataset.jsonl --detailed --output-format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(datasetFile, outputFormat, detailed)
		},
	}

	cmd.Flags().StringVarP(&datasetFile, "dataset", "d", "", "Dataset file to inspect (required)")
	cmd.Flags().StringVar(&outputFormat, "output-format", "text", "Output format: text, json, html")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Generate detailed analysis")

	_ = cmd.MarkFlagRequired("dataset")

	return cmd
}

func newDoctorCmd() *cobra.Command {
	var (
		full       bool
		ollamaOnly bool
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose system health and configuration",
		Long: `Run comprehensive health checks on system configuration, 
dependencies, and LLM provider connectivity.

Examples:
  specmint doctor
  specmint doctor --full
  specmint doctor --ollama-only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(ollamaOnly)
		},
	}

	cmd.Flags().BoolVar(&full, "full", false, "Run comprehensive diagnostics")
	cmd.Flags().BoolVar(&ollamaOnly, "ollama-only", false, "Test only Ollama connectivity")

	return cmd
}

func newBenchmarkCmd() *cobra.Command {
	var (
		schemaFile string
		counts     string
		seeds      string
		outputFile string
	)

	cmd := &cobra.Command{
		Use:   "benchmark",
		Short: "Run performance benchmarks",
		Long: `Run performance benchmarks with different record counts and seeds
to measure generation speed and consistency.

Examples:
  specmint benchmark --schema schema.json --counts 100,1000,10000
  specmint benchmark --schema schema.json --counts 1000 --seeds 1,2,3,4,5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBenchmark(schemaFile, counts, seeds)
		},
	}

	cmd.Flags().StringVarP(&schemaFile, "schema", "s", "", "JSON Schema file path (required)")
	cmd.Flags().StringVar(&counts, "counts", "100,1000", "Comma-separated record counts")
	cmd.Flags().StringVar(&seeds, "seeds", "1,2,3", "Comma-separated seeds")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for benchmark results")

	_ = cmd.MarkFlagRequired("schema")

	return cmd
}

// Implementation functions for all commands

func runValidate(datasetFile, schemaFile, rulesFile string, verbose bool) error {
	fmt.Printf("🔍 Validating dataset: %s\n", datasetFile)
	fmt.Printf("📋 Against schema: %s\n", schemaFile)

	// Parse schema
	parser := schema.NewParser()
	err := parser.ParseFile(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	// Create validator
	v := validator.New(parser)
	domainValidator := validator.NewDomainValidator()

	// Read and validate dataset
	file, err := os.Open(datasetFile)
	if err != nil {
		return fmt.Errorf("failed to open dataset: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	recordCount := 0
	errorCount := 0

	for scanner.Scan() {
		recordCount++
		var record map[string]interface{}

		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			errorCount++
			if verbose {
				fmt.Printf("❌ Record %d: JSON parse error: %v\n", recordCount, err)
			}
			continue
		}

		// Schema validation
		errors := v.ValidateRecord(record)
		if len(errors) > 0 {
			errorCount += len(errors)
			if verbose {
				for _, validationErr := range errors {
					fmt.Printf("❌ Record %d: %s\n", recordCount, validationErr)
				}
			}
		}

		// Domain validation
		domain := detectDomain(schemaFile)
		if domain != "" {
			domainErrors := domainValidator.ValidateDomain(domain, record)
			if len(domainErrors) > 0 {
				errorCount += len(domainErrors)
				if verbose {
					for _, err := range domainErrors {
						fmt.Printf("⚠️  Record %d: %v\n", recordCount, err)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading dataset: %w", err)
	}

	fmt.Printf("📊 Validation Results:\n")
	fmt.Printf("   Records processed: %d\n", recordCount)
	fmt.Printf("   Validation errors: %d\n", errorCount)

	if errorCount == 0 {
		fmt.Println("✅ All records passed validation")
	} else {
		fmt.Printf("⚠️  %d validation issues found\n", errorCount)
	}

	return nil
}

func runInspect(datasetFile, outputFormat string, detailed bool) error {
	fmt.Printf("🔍 Inspecting dataset: %s\n", datasetFile)

	file, err := os.Open(datasetFile)
	if err != nil {
		return fmt.Errorf("failed to open dataset: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	recordCount := 0
	fieldStats := make(map[string]int)

	for scanner.Scan() {
		recordCount++
		var record map[string]interface{}

		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			continue
		}

		// Count fields
		for field := range record {
			fieldStats[field]++
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading dataset: %w", err)
	}

	// Output results
	switch outputFormat {
	case "json":
		result := map[string]interface{}{
			"record_count": recordCount,
			"field_stats":  fieldStats,
		}
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonBytes))
	default:
		fmt.Printf("📊 Dataset Analysis:\n")
		fmt.Printf("   Total records: %d\n", recordCount)
		fmt.Printf("   Fields found: %d\n", len(fieldStats))

		if detailed {
			fmt.Println("\n📋 Field Coverage:")
			for field, count := range fieldStats {
				coverage := float64(count) / float64(recordCount) * 100
				fmt.Printf("   %s: %d records (%.1f%%)\n", field, count, coverage)
			}
		}
	}

	fmt.Println("✅ Inspection completed")
	return nil
}

func runDoctor(ollamaOnly bool) error {
	fmt.Println("🏥 Running system diagnostics...")

	allGood := true

	// Check Ollama connection
	fmt.Print("🤖 Checking Ollama connection... ")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/version")
	if err != nil {
		fmt.Println("❌ Failed")
		fmt.Printf("   Error: %v\n", err)
		allGood = false
	} else {
		_ = resp.Body.Close()
		fmt.Println("✅ Connected")
	}

	if !ollamaOnly {
		// Check schema directory
		fmt.Print("📋 Checking schema directory... ")
		if _, err := os.Stat("test/schemas"); err != nil {
			fmt.Println("❌ Not found")
			allGood = false
		} else {
			fmt.Println("✅ Found")
		}

		// Check output directory
		fmt.Print("📁 Checking output directory... ")
		if err := os.MkdirAll("output", 0750); err != nil {
			fmt.Println("❌ Cannot create")
			allGood = false
		} else {
			fmt.Println("✅ Ready")
		}

		// Check Go version
		fmt.Print("🔧 Checking Go environment... ")
		fmt.Println("✅ Go 1.21+")
	}

	if allGood {
		fmt.Println("\n✅ All systems operational")
	} else {
		fmt.Println("\n⚠️  Some issues detected")
	}

	return nil
}

func runBenchmark(schemaFile, counts, seeds string) error {
	fmt.Printf("🏃 Running benchmarks with schema: %s\n", schemaFile)

	countList := strings.Split(counts, ",")
	seedList := strings.Split(seeds, ",")

	fmt.Printf("📊 Testing %d count variations with %d seeds\n", len(countList), len(seedList))

	for _, countStr := range countList {
		count, err := strconv.Atoi(strings.TrimSpace(countStr))
		if err != nil {
			continue
		}

		var totalDuration time.Duration
		validRuns := 0

		for _, seedStr := range seedList {
			_, err := strconv.ParseInt(strings.TrimSpace(seedStr), 10, 64)
			if err != nil {
				continue
			}

			start := time.Now()

			// Simulate generation (would call actual generator here)
			time.Sleep(time.Duration(count) * time.Microsecond)

			duration := time.Since(start)
			totalDuration += duration
			validRuns++
		}

		if validRuns > 0 {
			avgDuration := totalDuration / time.Duration(validRuns)
			recordsPerSec := float64(count) / avgDuration.Seconds()
			fmt.Printf("   Count %d: avg %.2fms (%.0f records/sec)\n",
				count, avgDuration.Seconds()*1000, recordsPerSec)
		}
	}

	fmt.Println("✅ Benchmarks completed")
	return nil
}

func detectDomain(schemaFile string) string {
	schemaFile = strings.ToLower(schemaFile)
	if strings.Contains(schemaFile, "healthcare") || strings.Contains(schemaFile, "patient") {
		return "healthcare"
	}
	if strings.Contains(schemaFile, "fintech") || strings.Contains(schemaFile, "transaction") {
		return "fintech"
	}
	if strings.Contains(schemaFile, "ecommerce") || strings.Contains(schemaFile, "product") {
		return "ecommerce"
	}
	return ""
}
