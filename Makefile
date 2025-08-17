# SpecMint - Synthetic Dataset Generator
# Comprehensive Makefile for build, test, lint, and deployment

.PHONY: help build test lint clean install deps security audit format vet staticcheck gosec nancy vulncheck docker run-tests coverage bench

# Default target
.DEFAULT_GOAL := help

# Build variables
BINARY_NAME := specmint
BUILD_DIR := bin
OUTPUT_DIR := output
MAIN_PATH := ./cmd/specmint
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go variables
GO := go
GOFMT := gofmt
GOLINT := golangci-lint
GOVET := $(GO) vet
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOMOD := $(GO) mod
GOINSTALL := $(GO) install

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_TIME) -s -w"
BUILD_FLAGS := -trimpath $(LDFLAGS)

# Test flags
TEST_FLAGS := -v -race -coverprofile=coverage.out
BENCH_FLAGS := -bench=. -benchmem

# Linting tools
LINTING_TOOLS := \
	github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
	github.com/securecodewarrior/nancy@latest \
	github.com/securecodewarrior/gosec/v2/cmd/gosec@latest \
	honnef.co/go/tools/cmd/staticcheck@latest

## help: Show this help message
help:
	@echo "SpecMint - Synthetic Dataset Generator"
	@echo "====================================="
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build: clean
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Built $(BUILD_DIR)/$(BINARY_NAME)"

## build-all: Build for multiple platforms
build-all: clean
	@echo "🔨 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ Built binaries for multiple platforms"

## install: Install the binary to GOPATH/bin
install:
	@echo "📦 Installing $(BINARY_NAME)..."
	$(GOINSTALL) $(BUILD_FLAGS) $(MAIN_PATH)
	@echo "✅ Installed $(BINARY_NAME)"

## deps: Download and tidy dependencies
deps:
	@echo "📥 Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies updated"

## deps-upgrade: Upgrade all dependencies
deps-upgrade:
	@echo "⬆️  Upgrading dependencies..."
	$(GO) get -u ./...
	$(GOMOD) tidy
	@echo "✅ Dependencies upgraded"

## test: Run all tests
test:
	@echo "🧪 Running tests..."
	$(GOTEST) $(TEST_FLAGS) ./...
	@echo "✅ Tests completed"

## test-short: Run short tests only
test-short:
	@echo "🧪 Running short tests..."
	$(GOTEST) -short $(TEST_FLAGS) ./...
	@echo "✅ Short tests completed"

## coverage: Run tests with coverage report
coverage: test
	@echo "📊 Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

## bench: Run benchmarks
bench:
	@echo "⚡ Running benchmarks..."
	$(GOTEST) $(BENCH_FLAGS) ./...
	@echo "✅ Benchmarks completed"

## lint: Run comprehensive linting
lint: install-lint-tools
	@echo "🔍 Running comprehensive linting..."
	$(GOLINT) run --config .golangci.yml
	@echo "✅ Linting completed"

## format: Format Go code
format:
	@echo "🎨 Formatting code..."
	$(GOFMT) -s -w .
	@echo "✅ Code formatted"

## vet: Run go vet
vet:
	@echo "🔍 Running go vet..."
	$(GOVET) ./...
	@echo "✅ Vet completed"

## staticcheck: Run staticcheck
staticcheck:
	@echo "🔍 Running staticcheck..."
	staticcheck ./...
	@echo "✅ Staticcheck completed"

## security: Run security audit
security: install-security-tools
	@echo "🔒 Running security audit..."
	gosec -fmt json -out gosec-report.json ./...
	@echo "✅ Security audit completed (report: gosec-report.json)"

## vulncheck: Check for known vulnerabilities
vulncheck:
	@echo "🛡️  Checking for vulnerabilities..."
	$(GO) install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...
	@echo "✅ Vulnerability check completed"

## nancy: Run Nancy dependency scanner
nancy:
	@echo "🔍 Running Nancy dependency scanner..."
	$(GO) list -json -deps ./... | nancy sleuth
	@echo "✅ Nancy scan completed"

## audit: Run full security audit
audit: vulncheck nancy security
	@echo "✅ Full security audit completed"

## install-lint-tools: Install linting tools
install-lint-tools:
	@echo "🔧 Installing linting tools..."
	@for tool in $(LINTING_TOOLS); do \
		echo "Installing $$tool..."; \
		$(GOINSTALL) $$tool; \
	done
	@echo "✅ Linting tools installed"

## install-security-tools: Install security tools
install-security-tools:
	@echo "🔧 Installing security tools..."
	$(GOINSTALL) github.com/securego/gosec/v2/cmd/gosec@latest
	$(GOINSTALL) golang.org/x/vuln/cmd/govulncheck@latest
	$(GOINSTALL) github.com/sonatype-nexus-community/nancy@latest
	@echo "✅ Security tools installed"

## run-tests: Run the golden test suite
run-tests: build
	@echo "🧪 Running golden test suite..."
	./test/golden-test-suite.sh
	@echo "✅ Golden test suite completed"

## doctor: Run system diagnostics
doctor: build
	@echo "🏥 Running system diagnostics..."
	./$(BUILD_DIR)/$(BINARY_NAME) doctor
	@echo "✅ System diagnostics completed"

## clean: Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(OUTPUT_DIR)
	rm -f coverage.out coverage.html
	rm -f gosec-report.json
	@echo "✅ Clean completed"

## clean-all: Clean everything including dependencies
clean-all: clean
	@echo "🧹 Cleaning everything..."
	$(GOMOD) clean -cache
	@echo "✅ Deep clean completed"

## docker-build: Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t specmint:$(VERSION) .
	docker tag specmint:$(VERSION) specmint:latest
	@echo "✅ Docker image built"

## docker-run: Run Docker container
docker-run: docker-build
	@echo "🐳 Running Docker container..."
	docker run --rm -it specmint:latest doctor
	@echo "✅ Docker container test completed"

## release: Build release artifacts
release: clean build-all test lint audit
	@echo "🚀 Creating release artifacts..."
	@mkdir -p release
	@cp $(BUILD_DIR)/* release/
	@tar -czf release/$(BINARY_NAME)-$(VERSION)-checksums.txt.gz -C release .
	@echo "✅ Release artifacts created in release/"

## ci: Run full CI pipeline
ci: deps format vet lint test audit build run-tests
	@echo "✅ CI pipeline completed successfully"

## dev: Development setup
dev: deps install-lint-tools install-security-tools build
	@echo "✅ Development environment ready"

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Built:   $(BUILD_TIME)"

# Generate completion scripts
## completion-bash: Generate bash completion
completion-bash: build
	./$(BUILD_DIR)/$(BINARY_NAME) completion bash > specmint-completion.bash

## completion-zsh: Generate zsh completion
completion-zsh: build
	./$(BUILD_DIR)/$(BINARY_NAME) completion zsh > specmint-completion.zsh

## completion-fish: Generate fish completion
completion-fish: build
	./$(BUILD_DIR)/$(BINARY_NAME) completion fish > specmint-completion.fish
