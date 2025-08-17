---
trigger: always_on
---

# Error Handling Rules

## Error Categories
<error_types>
- **Schema Errors**: Invalid JSON Schema, parsing failures
- **Generation Errors**: Type generation failures, constraint violations
- **LLM Errors**: Connection failures, timeout, invalid responses
- **Validation Errors**: Schema compliance, cross-field rule violations
- **System Errors**: File I/O, memory, network issues
</error_types>

## Error Handling Patterns
<error_patterns>
- Use structured errors with error codes and context
- Implement graceful degradation for LLM failures
- Retry with exponential backoff for transient failures
- Circuit breaker pattern for external service calls
- Fail fast for configuration and schema errors
- Continue processing other records on individual failures
</error_patterns>

## Error Recovery
<recovery_strategies>
- **LLM Failures**: Fallback to deterministic generation
- **Schema Violations**: Apply minimal patching if possible
- **Network Issues**: Use cached responses when available
- **Memory Pressure**: Reduce batch sizes and flush buffers
- **Timeout**: Extend deadline or skip expensive operations
</recovery_strategies>

## Error Reporting
<error_reporting>
- Log all errors with structured context (zerolog)
- Include error codes, timestamps, and correlation IDs
- Aggregate error counts in metrics
- Generate error summaries in manifest files
- Provide actionable error messages to users
- Never expose sensitive data in error messages
</error_reporting>

## Testing Error Conditions
<error_testing>
- Unit tests for all error paths
- Chaos engineering for network failures
- Resource exhaustion testing
- Invalid input fuzzing
- LLM service unavailability scenarios
</error_testing>
