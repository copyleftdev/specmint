# Development Workflow

Complete development workflow for SpecMint including code quality checks, testing, and deployment preparation.

## Steps

1. **Pre-Development Setup**
   ```bash
   # Ensure Go environment is ready
   go version
   go mod tidy
   go mod download
   
   # Start Ollama if needed
   pgrep ollama || ollama serve &
   sleep 3
   ```

2. **Code Quality Checks**
   ```bash
   # Format code
   go fmt ./...
   
   # Run linter
   golangci-lint run ./...
   
   # Check for security issues
   gosec ./...
   
   # Verify module dependencies
   go mod verify
   ```

3. **Run Tests**
   ```bash
   # Unit tests with coverage
   go test -v -race -coverprofile=coverage.out ./...
   
   # Generate coverage report
   go tool cover -html=coverage.out -o coverage.html
   
   # Integration tests
   go test -v -tags=integration ./...
   ```

4. **Build and Validate**
   ```bash
   # Build binary
   go build -o bin/specmint ./cmd/specmint
   
   # Test CLI help
   ./bin/specmint --help
   
   # Run doctor command
   ./bin/specmint doctor --full
   ```

5. **Performance Testing**
   ```bash
   # Run benchmarks
   go test -bench=. -benchmem ./...
   
   # Profile memory usage
   go test -memprofile=mem.prof -bench=BenchmarkGenerate ./...
   go tool pprof mem.prof
   ```

6. **Documentation Updates**
   ```bash
   # Generate godoc
   godoc -http=:6060 &
   
   # Update CLI help documentation
   ./bin/specmint generate-docs --output ./docs/cli.md
   
   # Validate markdown files
   markdownlint docs/*.md
   ```

## Success Criteria
- All tests pass with >90% coverage
- No linting errors or security issues
- Binary builds successfully
- Performance benchmarks within targets
- Documentation is up to date
