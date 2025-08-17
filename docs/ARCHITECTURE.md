# SpecMint Architecture

This document provides detailed architectural information for the SpecMint synthetic dataset generator.

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        SpecMint CLI                            │
├─────────────────────────────────────────────────────────────────┤
│  cmd/specmint/                                                  │
│  ├── main.go           # Application entry point               │
│  └── commands.go       # CLI command definitions               │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Core Generation Engine                     │
├─────────────────────────────────────────────────────────────────┤
│  pkg/generator/                                                 │
│  ├── generator.go      # Main generation orchestrator          │
│  └── deterministic.go  # Deterministic value generation        │
└─────────────────────────────────────────────────────────────────┘
                │                           │                    │
                ▼                           ▼                    ▼
┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐
│   Schema Parser     │  │   LLM Integration   │  │   Domain Validator  │
├─────────────────────┤  ├─────────────────────┤  ├─────────────────────┤
│ pkg/schema/         │  │ pkg/llm/            │  │ pkg/validator/      │
│ └── parser.go       │  │ └── ollama.go       │  │ ├── validator.go    │
│                     │  │                     │  │ └── domain_rules.go │
│ • JSON Schema       │  │ • Local Ollama      │  │ • Business Rules    │
│ • Validation        │  │ • Field Enhancement │  │ • Cross-field Valid │
│ • Constraint Parse  │  │ • Fallback Logic    │  │ • Domain Logic      │
└─────────────────────┘  └─────────────────────┘  └─────────────────────┘
                                    │
                                    ▼
                        ┌─────────────────────┐
                        │   Output Writer     │
                        ├─────────────────────┤
                        │ pkg/writer/         │
                        │ └── writer.go       │
                        │                     │
                        │ • JSONL/JSON Output │
                        │ • Manifest Files    │
                        │ • Format Handling   │
                        └─────────────────────┘
```

## 📦 Package Structure

### Core Packages (`pkg/`)

#### `pkg/generator/`
- **Purpose**: Core dataset generation logic
- **Key Components**:
  - `generator.go`: Main orchestrator, coordinates all generation phases
  - `deterministic.go`: Seeded random generation for reproducibility
- **Responsibilities**: Schema-compliant data generation, LLM coordination

#### `pkg/schema/`
- **Purpose**: JSON Schema parsing and validation
- **Key Components**:
  - `parser.go`: Schema parsing, constraint extraction, validation
- **Responsibilities**: Schema compliance, constraint handling, field mapping

#### `pkg/llm/`
- **Purpose**: Local LLM integration for data enhancement
- **Key Components**:
  - `ollama.go`: Ollama client, prompt generation, response handling
- **Responsibilities**: Field-level enhancement, fallback logic, rate limiting

#### `pkg/validator/`
- **Purpose**: Domain-specific business rule validation
- **Key Components**:
  - `validator.go`: Validation orchestrator and framework
  - `domain_rules.go`: Healthcare, fintech, e-commerce rules
- **Responsibilities**: Business logic validation, cross-field consistency

#### `pkg/writer/`
- **Purpose**: Output formatting and file management
- **Key Components**:
  - `writer.go`: Multi-format output, manifest generation
- **Responsibilities**: File I/O, format handling, metadata tracking

### Internal Packages (`internal/`)

#### `internal/config/`
- **Purpose**: Configuration management
- **Responsibilities**: Settings parsing, environment handling, validation

#### `internal/logger/`
- **Purpose**: Structured logging
- **Responsibilities**: Log formatting, level management, output routing

## 🔄 Data Flow

1. **CLI Input** → Command parsing and validation
2. **Schema Loading** → JSON Schema parsing and constraint extraction
3. **Generation Planning** → Record count, seeding, worker allocation
4. **Core Generation** → Deterministic value generation per schema
5. **LLM Enhancement** → Optional field-level enrichment (if enabled)
6. **Domain Validation** → Business rule compliance checking
7. **Output Writing** → JSONL/JSON formatting and file creation
8. **Manifest Creation** → Metadata and generation summary

## 🎯 Design Principles

### Modularity
- Clear separation of concerns
- Pluggable architecture for new domains
- Independent package testing

### Performance
- Streaming output for large datasets
- Configurable worker pools
- Efficient memory usage

### Reliability
- Comprehensive error handling
- Graceful LLM fallback
- Deterministic reproducibility

### Security
- Input validation at all boundaries
- Secure file permissions
- No data persistence beyond output

## 🔌 Extension Points

### Adding New Domains
1. Extend `pkg/validator/domain_rules.go`
2. Add domain-specific validation logic
3. Create test schemas in `test/schemas/`

### Adding Output Formats
1. Extend `pkg/writer/writer.go`
2. Implement format-specific writers
3. Update CLI format options

### Adding LLM Providers
1. Create new client in `pkg/llm/`
2. Implement common interface
3. Add provider selection logic

This architecture ensures maintainability, extensibility, and production readiness while maintaining clean separation of concerns.
