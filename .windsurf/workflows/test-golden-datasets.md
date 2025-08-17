# Test Golden Datasets

Validate SpecMint against golden datasets for all supported domains with comprehensive regression testing.

## Steps

1. **Setup Test Environment**
   ```bash
   # Ensure Ollama is running with qwen2.5:latest
   ollama serve &
   ollama pull qwen2.5:latest
   
   # Create test output directory
   mkdir -p ./test-results/golden-$(date +%Y%m%d-%H%M%S)
   ```

2. **Run Healthcare Golden Tests**
   ```bash
   # Generate healthcare dataset
   specmint generate \
     --schema ./test/golden/healthcare/claims_schema.json \
     --count 100 \
     --seed 12345 \
     --out ./test-results/golden-$(date +%Y%m%d-%H%M%S)/healthcare \
     --llm-mode fields
   
   # Compare against golden reference
   diff -u ./test/golden/healthcare/expected_output.jsonl \
           ./test-results/golden-$(date +%Y%m%d-%H%M%S)/healthcare/dataset.jsonl
   ```

3. **Run Fintech Golden Tests**
   ```bash
   # Generate fintech dataset  
   specmint generate \
     --schema ./test/golden/fintech/transactions_schema.json \
     --count 100 \
     --seed 67890 \
     --out ./test-results/golden-$(date +%Y%m%d-%H%M%S)/fintech \
     --llm-mode record
   
   # Validate business rules
   specmint validate \
     --schema ./test/golden/fintech/transactions_schema.json \
     --dataset ./test-results/golden-$(date +%Y%m%d-%H%M%S)/fintech/dataset.jsonl \
     --rules ./test/golden/fintech/business_rules.json
   ```

4. **Run E-commerce Golden Tests**
   ```bash
   # Generate e-commerce dataset
   specmint generate \
     --schema ./test/golden/ecommerce/products_schema.json \
     --count 100 \
     --seed 11111 \
     --out ./test-results/golden-$(date +%Y%m%d-%H%M%S)/ecommerce \
     --llm-mode off
   
   # Check deterministic consistency
   specmint generate \
     --schema ./test/golden/ecommerce/products_schema.json \
     --count 100 \
     --seed 11111 \
     --out ./test-results/golden-$(date +%Y%m%d-%H%M%S)/ecommerce-verify \
     --llm-mode off
   
   # Verify identical outputs
   diff ./test-results/golden-$(date +%Y%m%d-%H%M%S)/ecommerce/dataset.jsonl \
        ./test-results/golden-$(date +%Y%m%d-%H%M%S)/ecommerce-verify/dataset.jsonl
   ```

5. **Run Edge Case Tests**
   ```bash
   # Test complex nested schemas
   for schema in ./test/edge-cases/*.json; do
     echo "Testing $(basename $schema)"
     specmint generate \
       --schema "$schema" \
       --count 10 \
       --seed 99999 \
       --out ./test-results/golden-$(date +%Y%m%d-%H%M%S)/edge-$(basename $schema .json) \
       --llm-mode fields \
       --timeout 60s
   done
   ```

6. **Performance Benchmarks**
   ```bash
   # Run performance tests
   specmint benchmark \
     --schema ./test/golden/healthcare/claims_schema.json \
     --counts 100,1000,10000 \
     --seeds 1,2,3,4,5 \
     --output ./test-results/golden-$(date +%Y%m%d-%H%M%S)/benchmarks.json
   ```

7. **Generate Test Report**
   ```bash
   # Aggregate all test results
   specmint test-report \
     --results-dir ./test-results/golden-$(date +%Y%m%d-%H%M%S) \
     --golden-dir ./test/golden \
     --output ./test-results/golden-$(date +%Y%m%d-%H%M%S)/report.html
   ```

## Success Criteria
- All golden datasets match expected outputs
- No schema validation failures
- Performance benchmarks within targets
- Edge cases handled gracefully
- Deterministic outputs are identical across runs
- LLM enrichment produces valid, realistic data
