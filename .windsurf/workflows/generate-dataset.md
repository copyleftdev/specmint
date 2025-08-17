# Generate Synthetic Dataset

Generate a synthetic dataset using SpecMint with comprehensive validation and monitoring.

## Steps

1. **Validate Environment**
   - Check Ollama is running: `curl -s http://localhost:11434/api/tags`
   - Verify qwen2.5:latest model is available
   - Validate schema file exists and is valid JSON Schema

2. **Run Health Check**
   ```bash
   specmint doctor --full
   ```
   - Verify all providers are healthy
   - Check system resources (memory, disk space)
   - Validate configuration files

3. **Generate Dataset**
   ```bash
   specmint generate \
     --schema ./schemas/{domain}/{schema_name}.json \
     --count {record_count} \
     --seed {seed_value} \
     --out ./output/{domain}_{timestamp} \
     --llm-mode {fields|record|off} \
     --workers 6 \
     --llm_workers 3 \
     --llm.max_rps 3
   ```

4. **Validate Output**
   ```bash
   specmint validate \
     --schema ./schemas/{domain}/{schema_name}.json \
     --dataset ./output/{domain}_{timestamp}/dataset.jsonl \
     --verbose
   ```

5. **Generate Report**
   ```bash
   specmint inspect \
     --dataset ./output/{domain}_{timestamp}/dataset.jsonl \
     --detailed \
     --output-format json
   ```

6. **Archive Results**
   - Copy manifest.json to results archive
   - Generate performance summary
   - Update golden datasets if this is a reference run

## Success Criteria
- 100% schema compliance
- All cross-field rules pass
- Performance targets met
- No LLM fallbacks (unless expected)
- Manifest contains complete metadata
