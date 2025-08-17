package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sony/gobreaker"
	"golang.org/x/time/rate"
)

// OllamaClient handles communication with local Ollama instance
type OllamaClient struct {
	baseURL     string
	model       string
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	breaker     *gobreaker.CircuitBreaker
	pool        *connectionPool
	config      OllamaConfig
}

// OllamaConfig holds Ollama-specific configuration
type OllamaConfig struct {
	Host        string
	Model       string
	AutoPull    bool
	KeepAlive   time.Duration
	MaxRetries  int
	Temperature float32
	MaxRPS      int
	Timeout     time.Duration
	MaxConns    int
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// ModelInfo represents model information from Ollama
type ModelInfo struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}

// ModelsResponse represents the response from /api/tags
type ModelsResponse struct {
	Models []ModelInfo `json:"models"`
}

// connectionPool manages HTTP connections to Ollama
type connectionPool struct {
	client    *http.Client
	maxConns  int
	semaphore chan struct{}
}

// NewOllamaClient creates a new Ollama client with health checks and connection pooling
func NewOllamaClient(config OllamaConfig) (*OllamaClient, error) {
	if config.Host == "" {
		config.Host = "http://localhost:11434"
	}
	if config.Model == "" {
		config.Model = "qwen2.5:latest"
	}
	if config.MaxRPS <= 0 {
		config.MaxRPS = 3
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxConns <= 0 {
		config.MaxConns = 4
	}

	// Create HTTP client with timeout and keep-alive
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        config.MaxConns,
			MaxIdleConnsPerHost: config.MaxConns,
			IdleConnTimeout:     config.KeepAlive,
			DisableKeepAlives:   false,
		},
	}

	// Create connection pool
	pool := &connectionPool{
		client:    httpClient,
		maxConns:  config.MaxConns,
		semaphore: make(chan struct{}, config.MaxConns),
	}

	// Initialize semaphore
	for i := 0; i < config.MaxConns; i++ {
		pool.semaphore <- struct{}{}
	}

	// Create rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(config.MaxRPS), config.MaxRPS)

	// Create circuit breaker
	breaker := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "ollama",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 2
		},
	})

	client := &OllamaClient{
		baseURL:     strings.TrimSuffix(config.Host, "/"),
		model:       config.Model,
		httpClient:  httpClient,
		rateLimiter: rateLimiter,
		breaker:     breaker,
		pool:        pool,
		config:      config,
	}

	return client, nil
}

// HealthCheck verifies Ollama connectivity and model availability
func (c *OllamaClient) HealthCheck(ctx context.Context) error {
	// Check if Ollama is running
	if err := c.ping(ctx); err != nil {
		return fmt.Errorf("ollama ping failed: %w", err)
	}

	// Check if model is available
	available, err := c.isModelAvailable(ctx)
	if err != nil {
		return fmt.Errorf("failed to check model availability: %w", err)
	}

	if !available {
		if c.config.AutoPull {
			if err := c.pullModel(ctx); err != nil {
				return fmt.Errorf("failed to pull model %s: %w", c.model, err)
			}
		} else {
			return fmt.Errorf("model %s not available and auto-pull disabled", c.model)
		}
	}

	return nil
}

// Generate generates text using Ollama with the given prompt and seed
func (c *OllamaClient) Generate(ctx context.Context, prompt string, seed int64) (string, error) {
	// Skip LLM calls in CI environment
	if os.Getenv("SKIP_OLLAMA_TESTS") == "true" {
		log.Debug().Msg("Skipping Ollama call in CI environment")
		return "", fmt.Errorf("ollama disabled in CI environment")
	}
	// Wait for rate limit
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Use circuit breaker
	result, err := c.breaker.Execute(func() (interface{}, error) {
		return c.generateWithRetry(ctx, prompt, seed)
	})

	if err != nil {
		return "", fmt.Errorf("generation failed: %w", err)
	}

	return result.(string), nil
}

// generateWithRetry performs the actual generation with retry logic
func (c *OllamaClient) generateWithRetry(ctx context.Context, prompt string, seed int64) (string, error) {
	var lastErr error

	for attempt := 0; attempt < c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt*attempt) * time.Second
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff):
			}
		}

		response, err := c.doGenerate(ctx, prompt, seed)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			break
		}
	}

	return "", fmt.Errorf("generation failed after %d attempts: %w", c.config.MaxRetries, lastErr)
}

// doGenerate performs a single generation request
func (c *OllamaClient) doGenerate(ctx context.Context, prompt string, seed int64) (string, error) {
	// Acquire connection from pool
	select {
	case <-c.pool.semaphore:
		defer func() { c.pool.semaphore <- struct{}{} }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// Prepare request
	options := map[string]interface{}{
		"temperature": c.config.Temperature,
		"seed":        seed,
	}

	req := OllamaRequest{
		Model:   c.model,
		Prompt:  prompt,
		Stream:  false,
		Options: options,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.pool.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("ollama error: %s", ollamaResp.Error)
	}

	return strings.TrimSpace(ollamaResp.Response), nil
}

// ping checks if Ollama is responding
func (c *OllamaClient) ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// isModelAvailable checks if the configured model is available
func (c *OllamaClient) isModelAvailable(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return false, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var modelsResp ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return false, err
	}

	for _, model := range modelsResp.Models {
		if model.Name == c.model {
			return true, nil
		}
	}

	return false, nil
}

// pullModel pulls the configured model if not available
func (c *OllamaClient) pullModel(ctx context.Context) error {
	pullReq := map[string]string{"name": c.model}
	reqBody, err := json.Marshal(pullReq)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/pull", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Wait for pull to complete (simplified - in production would stream progress)
	time.Sleep(10 * time.Second)

	return nil
}

// Close closes the client and releases resources
func (c *OllamaClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// GetStats returns client statistics
func (c *OllamaClient) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"model":           c.model,
		"base_url":        c.baseURL,
		"max_rps":         c.config.MaxRPS,
		"max_connections": c.config.MaxConns,
		"breaker_state":   c.breaker.State().String(),
	}
}
