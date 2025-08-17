package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type contextKey string

const configKey contextKey = "config"

// Config represents the complete SpecMint configuration
type Config struct {
	Debug      bool       `yaml:"debug" json:"debug"`
	Schema     string     `yaml:"schema" json:"schema"`
	Generation Generation `yaml:"generation" json:"generation"`
	LLM        LLM        `yaml:"llm" json:"llm"`
	Output     Output     `yaml:"output" json:"output"`
	Logging    Logging    `yaml:"logging" json:"logging"`
	Metrics    Metrics    `yaml:"metrics" json:"metrics"`
}

type Generation struct {
	Count   int   `yaml:"count" json:"count"`
	Seed    int64 `yaml:"seed" json:"seed"`
	Workers int   `yaml:"workers" json:"workers"`
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
}

type LLM struct {
	Mode       string            `yaml:"mode" json:"mode"` // off, fields, record
	Provider   string            `yaml:"provider" json:"provider"` // auto, ollama, openai, anthropic
	Workers    int               `yaml:"workers" json:"workers"`
	MaxRPS     int               `yaml:"max_rps" json:"max_rps"`
	Timeout    time.Duration     `yaml:"timeout" json:"timeout"`
	Ollama     OllamaConfig      `yaml:"ollama" json:"ollama"`
	OpenAI     OpenAIConfig      `yaml:"openai" json:"openai"`
	Anthropic  AnthropicConfig   `yaml:"anthropic" json:"anthropic"`
	Budget     BudgetConfig      `yaml:"budget" json:"budget"`
}

type OllamaConfig struct {
	Host         string        `yaml:"host" json:"host"`
	Model        string        `yaml:"model" json:"model"`
	AutoPull     bool          `yaml:"auto_pull" json:"auto_pull"`
	KeepAlive    time.Duration `yaml:"keep_alive" json:"keep_alive"`
	MaxRetries   int           `yaml:"max_retries" json:"max_retries"`
	Temperature  float32       `yaml:"temperature" json:"temperature"`
}

type OpenAIConfig struct {
	APIKey      string  `yaml:"api_key" json:"api_key"`
	Model       string  `yaml:"model" json:"model"`
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`
	Temperature float32 `yaml:"temperature" json:"temperature"`
}

type AnthropicConfig struct {
	APIKey      string  `yaml:"api_key" json:"api_key"`
	Model       string  `yaml:"model" json:"model"`
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`
	Temperature float32 `yaml:"temperature" json:"temperature"`
}

type BudgetConfig struct {
	MaxCostUSD     float64 `yaml:"max_cost_usd" json:"max_cost_usd"`
	WarnThreshold  float64 `yaml:"warn_threshold" json:"warn_threshold"`
	TrackingEnabled bool   `yaml:"tracking_enabled" json:"tracking_enabled"`
}

type Output struct {
	Directory string `yaml:"directory" json:"directory"`
	Format    string `yaml:"format" json:"format"` // jsonl, json
	Manifest  bool   `yaml:"manifest" json:"manifest"`
	Compress  bool   `yaml:"compress" json:"compress"`
}

type Logging struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"` // json, text
	File   string `yaml:"file" json:"file"`
}

type Metrics struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Port    int    `yaml:"port" json:"port"`
	Path    string `yaml:"path" json:"path"`
}

// Default returns a configuration with sensible defaults
func Default() *Config {
	return &Config{
		Debug: false,
		Generation: Generation{
			Count:   100,
			Seed:    time.Now().UnixNano(),
			Workers: 4,
			Timeout: 5 * time.Minute,
		},
		LLM: LLM{
			Mode:     "off",
			Provider: "auto",
			Workers:  2,
			MaxRPS:   3,
			Timeout:  30 * time.Second,
			Ollama: OllamaConfig{
				Host:        "http://localhost:11434",
				Model:       "qwen2.5:latest",
				AutoPull:    true,
				KeepAlive:   5 * time.Minute,
				MaxRetries:  3,
				Temperature: 0.1,
			},
			OpenAI: OpenAIConfig{
				Model:       "gpt-4o-mini",
				MaxTokens:   1000,
				Temperature: 0.1,
			},
			Anthropic: AnthropicConfig{
				Model:       "claude-3-haiku-20240307",
				MaxTokens:   1000,
				Temperature: 0.1,
			},
			Budget: BudgetConfig{
				MaxCostUSD:      10.0,
				WarnThreshold:   0.8,
				TrackingEnabled: true,
			},
		},
		Output: Output{
			Directory: "./output",
			Format:    "jsonl",
			Manifest:  true,
			Compress:  false,
		},
		Logging: Logging{
			Level:  "info",
			Format: "json",
		},
		Metrics: Metrics{
			Enabled: true,
			Port:    9090,
			Path:    "/metrics",
		},
	}
}

// Load configuration from file with environment variable overrides
func Load(configFile string) (*Config, error) {
	cfg := Default()

	// Load from file if specified
	if configFile != "" {
		if err := loadFromFile(cfg, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	} else {
		// Try default locations
		for _, path := range []string{"specmint.yaml", "specmint.yml", ".specmint.yaml"} {
			if _, err := os.Stat(path); err == nil {
				if err := loadFromFile(cfg, path); err != nil {
					return nil, fmt.Errorf("failed to load config file %s: %w", path, err)
				}
				break
			}
		}
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func loadFromFile(cfg *Config, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cfg)
}

func applyEnvOverrides(cfg *Config) {
	if val := os.Getenv("SPECMINT_DEBUG"); val == "true" {
		cfg.Debug = true
	}
	if val := os.Getenv("SPECMINT_SCHEMA"); val != "" {
		cfg.Schema = val
	}
	if val := os.Getenv("SPECMINT_OUT"); val != "" {
		cfg.Output.Directory = val
	}
	if val := os.Getenv("SPECMINT_SEED"); val != "" {
		if seed, err := time.Parse(time.RFC3339, val); err == nil {
			cfg.Generation.Seed = seed.UnixNano()
		}
	}
	if val := os.Getenv("OLLAMA_HOST"); val != "" {
		cfg.LLM.Ollama.Host = val
	}
	if val := os.Getenv("OPENAI_API_KEY"); val != "" {
		cfg.LLM.OpenAI.APIKey = val
	}
	if val := os.Getenv("ANTHROPIC_API_KEY"); val != "" {
		cfg.LLM.Anthropic.APIKey = val
	}
}

// Validate checks configuration for consistency and required values
func (c *Config) Validate() error {
	if c.Generation.Count <= 0 {
		return fmt.Errorf("generation count must be positive")
	}
	if c.Generation.Workers <= 0 {
		c.Generation.Workers = 4
	}
	if c.LLM.Workers <= 0 {
		c.LLM.Workers = 2
	}
	if c.LLM.MaxRPS <= 0 {
		c.LLM.MaxRPS = 3
	}
	if c.Output.Directory == "" {
		return fmt.Errorf("output directory is required")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(c.Output.Directory, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

// WithContext stores the config in context
func WithContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, configKey, cfg)
}

// FromContext retrieves the config from context
func FromContext(ctx context.Context) *Config {
	if cfg, ok := ctx.Value(configKey).(*Config); ok {
		return cfg
	}
	return Default()
}
