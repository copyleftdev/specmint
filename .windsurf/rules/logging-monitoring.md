---
trigger: always_on
---

# Logging and Monitoring Rules

## Structured Logging
<logging_standards>
- Use zerolog for structured JSON logging
- Include correlation IDs for request tracing
- Log levels: DEBUG, INFO, WARN, ERROR, FATAL
- Never log sensitive data (API keys, PII)
- Include context: operation, duration, record_count
- Use consistent field names across components
</logging_standards>

## Log Categories
<log_categories>
- **Generation**: Record counts, processing time, success rates
- **LLM**: Request/response times, token usage, model info
- **Validation**: Schema compliance, rule violations, patch operations
- **Performance**: Memory usage, CPU utilization, throughput
- **Errors**: Failure modes, recovery actions, impact assessment
</log_categories>

## Metrics Collection
<metrics_standards>
- Use Prometheus metrics with consistent naming
- Counter: specmint_records_generated_total
- Histogram: specmint_generation_duration_seconds
- Gauge: specmint_memory_usage_bytes
- Counter: specmint_llm_requests_total{provider,model,status}
- Histogram: specmint_llm_response_duration_seconds
</metrics_standards>

## Observability Requirements
<observability_rules>
- Export metrics on /metrics endpoint
- Include health check endpoint /health
- Generate performance reports in manifest
- Track resource utilization trends
- Monitor LLM provider availability
- Alert on error rate thresholds
</observability_rules>

## Production Monitoring
<monitoring_setup>
- Dashboard panels for key metrics
- Alerting on SLA violations
- Log aggregation and search
- Performance trend analysis
- Capacity planning metrics
- Cost tracking for cloud LLM usage
</monitoring_setup>

## Debug and Troubleshooting
<debug_guidelines>
- Enable verbose logging with --debug flag
- Include stack traces for unexpected errors
- Provide diagnostic commands (doctor, inspect)
- Generate detailed reports for support
- Maintain debug symbols in binaries
</debug_guidelines>
