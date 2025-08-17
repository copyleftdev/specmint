---
trigger: always_on
---

# SpecMint Core Architecture Rules

## Project Overview
- **Project**: SpecMint - High-performance synthetic dataset generator
- **Language**: Go (primary), with JSON Schema validation
- **Architecture**: CLI + library with modular components
- **Local LLM**: Ollama integration (qwen2.5:latest) at localhost:11434
- **Domains**: Healthcare, Fintech, E-commerce synthetic data

## Code Quality Standards
<coding_guidelines>
- Use Go 1.21+ with modules
- Follow standard Go project layout (cmd/, pkg/, internal/)
- Implement comprehensive error handling with structured errors
- Use context.Context for cancellation and timeouts
- Prefer composition over inheritance
- Write self-documenting code with clear variable names
- Add godoc comments for all public functions and types
</coding_guidelines>

## Performance Requirements
<performance_standards>
- Target: 10k deterministic records <30s
- Memory: <200MB steady-state
- Ollama calls: <2s p95 latency
- Schema validation: 500+ records/sec
- Concurrent processing with worker pools
- Use connection pooling for HTTP clients
</performance_standards>

## Testing Standards
<testing_requirements>
- Maintain >90% test coverage
- Include golden dataset tests for deterministic validation
- Property-based testing for schema generation
- Benchmark tests for performance validation
- Edge case testing for complex schemas
- Integration tests with local Ollama
</testing_requirements>

## Security & Privacy
<security_guidelines>
- No real PII/PHI in synthetic data
- Validate and sanitize all LLM outputs
- Use environment variables for API keys
- Implement rate limiting and circuit breakers
- Support offline-only mode
- Redact sensitive information in logs
</security_guidelines>
