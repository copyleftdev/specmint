package main

import (
	"context"
	"os"

	"github.com/specmint/specmint/internal/config"
	"github.com/specmint/specmint/internal/logger"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	ctx := context.Background()
	
	// Initialize logger
	log := logger.New()
	
	rootCmd := &cobra.Command{
		Use:   "specmint",
		Short: "High-performance synthetic dataset generator",
		Long: `SpecMint generates synthetic datasets from JSON Schema with deterministic 
seeded generation and optional LLM enrichment via local Ollama or cloud providers.`,
		Version: version,
	}

	// Add global flags
	var configFile string
	var debug bool
	
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is specmint.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	
	// Initialize config
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configFile)
		if err != nil {
			return err
		}
		
		if debug {
			cfg.Debug = true
		}
		
		// Update logger level
		if cfg.Debug {
			log = logger.WithLevel("debug")
		}
		
		// Store config in context
		ctx = config.WithContext(ctx, cfg)
		cmd.SetContext(ctx)
		
		return nil
	}

	// Add subcommands
	rootCmd.AddCommand(
		newGenerateCmd(),
		newValidateCmd(),
		newInspectCmd(),
		newDoctorCmd(),
		newBenchmarkCmd(),
	)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Error().Err(err).Msg("Command failed")
		os.Exit(1)
	}
}
