# CI/CD Pipeline

Continuous integration and deployment pipeline for SpecMint with automated testing and validation.

## Steps

1. **Environment Setup**
   ```bash
   # Install dependencies
   go mod download
   
   # Install tools
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
   
   # Setup Ollama in CI
   curl -fsSL https://ollama.ai/install.sh | sh
   ollama serve &
   sleep 10
   ollama pull qwen2.5:latest
   ```

2. **Code Quality Gates**
   ```bash
   # Format check
   test -z "$(go fmt ./...)"
   
   # Lint check
   golangci-lint run --timeout=5m ./...
   
   # Security scan
   gosec -quiet ./...
   
   # Dependency check
   go mod verify
   go mod tidy
   git diff --exit-code go.mod go.sum
   ```

3. **Test Suite**
   ```bash
   # Unit tests
   go test -v -race -coverprofile=coverage.out ./...
   
   # Coverage threshold check
   go tool cover -func=coverage.out | grep total | awk "{if(\$3+0 < 90) exit 1}"
   
   # Integration tests
   go test -v -tags=integration -timeout=10m ./...
   
   # Golden dataset tests
   /test-golden-datasets
   ```

4. **Build Artifacts**
   ```bash
   # Build for multiple platforms
   GOOS=linux GOARCH=amd64 go build -o bin/specmint-linux-amd64 ./cmd/specmint
   GOOS=darwin GOARCH=amd64 go build -o bin/specmint-darwin-amd64 ./cmd/specmint
   GOOS=windows GOARCH=amd64 go build -o bin/specmint-windows-amd64.exe ./cmd/specmint
   
   # Generate checksums
   sha256sum bin/* > bin/checksums.txt
   ```

5. **Performance Validation**
   ```bash
   # Benchmark tests
   go test -bench=. -benchmem -count=3 ./... > benchmark.txt
   
   # Performance regression check
   benchcmp baseline.txt benchmark.txt
   ```

6. **Release Preparation**
   ```bash
   # Tag version
   git tag -a v$(cat VERSION) -m "Release v$(cat VERSION)"
   
   # Generate changelog
   git log --oneline --since="$(git describe --tags --abbrev=0 HEAD^)" > CHANGELOG.md
   
   # Package release
   tar -czf specmint-$(cat VERSION)-linux-amd64.tar.gz -C bin specmint-linux-amd64
   ```

## Success Criteria
- All quality gates pass
- Test coverage >90%
- All golden datasets validate
- Performance benchmarks within thresholds
- Artifacts build successfully for all platforms
