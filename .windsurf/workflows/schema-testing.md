# Schema Validation and Testing

Comprehensive schema validation workflow for SpecMint with domain-specific testing.

## Steps

1. **Schema Validation**
   ```bash
   # Validate all schemas are valid JSON Schema
   for schema in schemas/**/*.json; do
     echo "Validating $schema"
     specmint validate-schema --schema "$schema" || exit 1
   done
   ```

2. **Domain-Specific Tests**
   ```bash
   # Healthcare schema tests
   specmint generate --schema schemas/healthcare/claims.json --count 10 --seed 1 --out test-healthcare
   specmint validate --schema schemas/healthcare/claims.json --dataset test-healthcare/dataset.jsonl
   
   # Fintech schema tests  
   specmint generate --schema schemas/fintech/transactions.json --count 10 --seed 2 --out test-fintech
   specmint validate --schema schemas/fintech/transactions.json --dataset test-fintech/dataset.jsonl
   
   # E-commerce schema tests
   specmint generate --schema schemas/ecommerce/products.json --count 10 --seed 3 --out test-ecommerce
   specmint validate --schema schemas/ecommerce/products.json --dataset test-ecommerce/dataset.jsonl
   ```

3. **Cross-Field Rule Testing**
   ```bash
   # Test business logic constraints
   specmint test-rules --schema schemas/healthcare/claims.json --rules test/rules/healthcare.json
   specmint test-rules --schema schemas/fintech/transactions.json --rules test/rules/fintech.json
   specmint test-rules --schema schemas/ecommerce/products.json --rules test/rules/ecommerce.json
   ```

4. **Edge Case Testing**
   ```bash
   # Test with minimal required fields only
   specmint generate --schema schemas/healthcare/claims.json --count 5 --seed 999 --minimal-fields --out test-minimal
   
   # Test with maximum complexity
   specmint generate --schema schemas/healthcare/claims.json --count 5 --seed 888 --max-complexity --out test-complex
   
   # Test with large arrays
   specmint generate --schema schemas/ecommerce/products.json --count 5 --seed 777 --large-arrays --out test-arrays
   ```

5. **Performance Testing**
   ```bash
   # Test generation speed
   time specmint generate --schema schemas/healthcare/claims.json --count 1000 --seed 123 --out perf-test
   
   # Test memory usage
   /usr/bin/time -v specmint generate --schema schemas/healthcare/claims.json --count 10000 --seed 456 --out memory-test
   ```

## Success Criteria
- All schemas pass validation
- Generated data conforms to schemas
- Cross-field rules are enforced
- Edge cases handled gracefully
- Performance targets met
