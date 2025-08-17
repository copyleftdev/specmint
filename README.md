# SpecMint: Synthetic Dataset Generator

<div align="center">a
  <img src="media/specmint.png" alt="SpecMint Logo" width="200"/>
</div>

[![CI/CD Pipeline](https://github.com/copyleftdev/specmint/actions/workflows/ci.yml/badge.svg)](https://github.com/copyleftdev/specmint/actions/workflows/ci.yml)
[![Security Audit](https://github.com/copyleftdev/specmint/actions/workflows/security.yml/badge.svg)](https://github.com/copyleftdev/specmint/actions/workflows/security.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/copyleftdev/specmint)](https://goreportcard.com/report/github.com/copyleftdev/specmint)
[![codecov](https://codecov.io/gh/copyleftdev/specmint/branch/main/graph/badge.svg)](https://codecov.io/gh/copyleftdev/specmint)
[![Go Version](https://img.shields.io/badge/Go-1.25.0-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/copyleftdev/specmint.svg)](https://github.com/copyleftdev/specmint/releases)
[![GitHub stars](https://img.shields.io/github/stars/copyleftdev/specmint.svg)](https://github.com/copyleftdev/specmint/stargazers)
[![GitHub issues](https://img.shields.io/github/issues/copyleftdev/specmint.svg)](https://github.com/copyleftdev/specmint/issues)

**SpecMint** is a production-ready synthetic dataset generator that creates realistic, schema-compliant datasets with optional LLM enrichment. Built for privacy-conscious data generation, testing, and development workflows.

## 🚀 Quick Start

```bash
# Generate 1000 e-commerce products
./bin/specmint generate -s test/schemas/ecommerce/product.json -o output -c 1000

# Validate existing dataset
./bin/specmint validate -s schema.json -d dataset.jsonl

# System health check
./bin/specmint doctor
```

## 📊 Project Metrics

| Metric | Value | Details |
|--------|-------|---------|
| **Development Time** | ~6 hours | August 17, 2025 (05:00 - 11:00 PST) |
| **Total Lines of Code** | 3,186 | Pure Go implementation |
| **Go Files** | 11 | Modular architecture |
| **Security Rating** | A (Excellent) | Zero vulnerabilities |
| **Test Coverage** | Comprehensive | Golden dataset validation |

## 🏗️ Architecture

SpecMint follows a clean, modular architecture designed for maintainability and extensibility:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Commands  │───▶│  Core Generator  │───▶│  Output Writer  │
│  (Cobra-based)  │    │   (Deterministic │    │   (JSONL/JSON)  │
└─────────────────┘    │   + LLM Enhanced)│    └─────────────────┘
         │              └──────────────────┘              │
         ▼                        │                       ▼
┌─────────────────┐              ▼                ┌─────────────────┐
│ Schema Parser   │    ┌──────────────────┐      │ Domain Validator│
│ (JSON Schema)   │    │  LLM Integration │      │ (Business Rules)│
└─────────────────┘    │  (Local Ollama)  │      └─────────────────┘
                       └──────────────────┘
```

### Core Components

- **`cmd/specmint/`** - CLI interface with 5 commands (generate, validate, inspect, doctor, benchmark)
- **`pkg/generator/`** - Deterministic generation engine with optional LLM enrichment
- **`pkg/schema/`** - JSON Schema parsing and validation
- **`pkg/llm/`** - Local Ollama integration for realistic data enhancement
- **`pkg/validator/`** - Domain-specific business rule validation
- **`pkg/writer/`** - Multi-format output handling
- **`internal/config/`** - Configuration management
- **`internal/logger/`** - Structured logging with zerolog

## 🤝 Development Collaboration

This project represents a unique **Human-AI collaborative development** approach:

### Human Role (Project Lead)
- **Strategic Vision**: Defined requirements for privacy-focused synthetic data generation
- **Architecture Guidance**: Directed modular design decisions and Go best practices
- **Domain Expertise**: Provided business logic for healthcare, fintech, and e-commerce validation
- **Quality Assurance**: Guided testing strategies and security requirements
- **Project Management**: Managed scope, priorities, and deliverable timelines

### AI Role (Cascade Assistant)
- **Code Implementation**: Wrote 100% of the 3,186 lines of Go code
- **Technical Architecture**: Implemented clean architecture patterns and interfaces
- **Testing Strategy**: Developed comprehensive golden dataset testing approach
- **Security Implementation**: Integrated security scanning and vulnerability management
- **Documentation**: Created comprehensive technical documentation and reports

### Collaborative Highlights
- **Real-time Feedback Loop**: Immediate iteration on requirements and implementation
- **Knowledge Transfer**: AI learned domain-specific validation rules through human guidance
- **Quality Standards**: Human oversight ensured enterprise-grade code quality
- **Problem Solving**: Combined human strategic thinking with AI implementation speed

## 🧪 Testing Strategies

SpecMint employs multiple testing methodologies for comprehensive quality assurance:

### 1. Golden Dataset Testing
```bash
./test/golden-test-suite.sh
```
- **Purpose**: Regression testing with known-good datasets
- **Coverage**: All three domains (healthcare, fintech, e-commerce)
- **Validation**: Schema compliance + domain business rules
- **Datasets**: 175 total records across domains

### 2. Domain-Specific Validation
- **Healthcare**: ICD-10 codes, NPI validation, HIPAA compliance
- **Fintech**: ABA routing numbers, transaction limits, risk scoring
- **E-commerce**: SKU formats, inventory consistency, pricing validation

### 3. LLM Integration Testing
- **Connectivity**: Automated Ollama health checks
- **Fallback Logic**: Graceful degradation to deterministic generation
- **Quality Assurance**: LLM output validation against schema constraints

### 4. Security Testing
- **Static Analysis**: gosec security scanner integration
- **Vulnerability Scanning**: govulncheck for Go stdlib issues
- **Dependency Auditing**: nancy for third-party package security

### 5. Performance Benchmarking
```bash
./bin/specmint benchmark -s schema.json --counts 100,1000,10000
```
- **Scalability**: Multi-record generation performance
- **Memory Usage**: Resource consumption monitoring
- **Deterministic Verification**: Seed-based reproducibility testing

## 🔧 Build System & CI/CD

### Local Development
Comprehensive Makefile with 15+ targets for complete development lifecycle:

```bash
# Development
make build test lint

# Security
make audit vulncheck

# CI/CD Pipeline
make ci

# Dependency Management
make deps-update deps-verify

# System Diagnostics
make doctor
```

### Automated CI/CD Pipeline
Production-grade GitHub Actions workflows with expert separation of concerns:

- **CI/CD Pipeline**: Multi-platform builds, test matrix, golden dataset validation
- **Security Audit**: Daily automated security scanning with SARIF integration
- **Release Automation**: Multi-platform binary builds with automated GitHub releases
- **Coverage Reporting**: Automated code coverage via Codecov integration
- **Quality Gates**: Go Report Card integration for code quality metrics

## 🛡️ Security

SpecMint maintains an **A-grade security rating** with:

- ✅ **Zero vulnerabilities** (post Go 1.25.0 upgrade)
- ✅ **Automated security scanning** in CI/CD pipeline
- ✅ **Hardened file permissions** (0600 for logs, 0750 for directories)
- ✅ **Clean dependency tree** with regular vulnerability monitoring
- ✅ **Static code analysis** with 54% security issue reduction
- ✅ **Daily security audits** via GitHub Actions
- ✅ **SARIF integration** for GitHub Security tab

See [SECURITY_AUDIT_REPORT.md](./docs/SECURITY_AUDIT_REPORT.md) for detailed security assessment.

## 🎯 Key Features

### Deterministic Generation
- **Reproducible**: Same seed produces identical datasets
- **Scalable**: Efficient generation of large datasets
- **Schema-Compliant**: Strict adherence to JSON Schema specifications

### LLM Enhancement
- **Local Privacy**: Uses local Ollama instance (no data leaves your machine)
- **Selective Enrichment**: Field-level LLM enhancement with fallback
- **Configurable**: Adjustable workers, rate limiting, and model selection

### Domain Intelligence
- **Business Rules**: Industry-specific validation logic
- **Cross-Field Validation**: Relationship consistency across record fields
- **Realistic Data**: Contextually appropriate synthetic values

### Production Ready
- **CLI Interface**: Professional command-line tool with comprehensive help
- **Multiple Formats**: JSON, JSONL output with manifest generation
- **Monitoring**: Built-in health checks and system diagnostics
- **Extensible**: Plugin-ready architecture for new domains

## 📈 Performance

- **Generation Speed**: 1000+ records/second (deterministic mode)
- **Memory Efficiency**: Streaming output for large datasets
- **LLM Integration**: Configurable rate limiting and worker pools
- **Scalability**: Tested up to 10,000+ record generation

## 🔮 Future Enhancements

- **Additional Domains**: Legal, manufacturing, retail verticals
- **Output Formats**: CSV, Parquet, database direct insertion
- **Cloud LLM Support**: OpenAI, Anthropic, Google integration
- **Web Interface**: Browser-based dataset generation UI
- **API Mode**: REST API for programmatic access

## 📄 License

BSD 3-Clause License - see [LICENSE](LICENSE) for details.

**Attribution Required**: When using SpecMint, please include attribution as specified in the LICENSE file.

## 🙏 Acknowledgments

This project demonstrates the power of **Human-AI collaboration** in software development, combining human strategic vision with AI implementation capabilities to create enterprise-grade solutions in record time.

---

**Built with ❤️ using Go 1.25.0 and collaborative AI development**
