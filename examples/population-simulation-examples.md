# Population-Based Simulation Examples

This guide demonstrates SpecMint's population-based simulation capabilities with real-world business scenarios.

## Healthcare Examples

### Small Regional Hospital
```bash
# 25-bed community hospital
./bin/specmint simulate --population "25-bed community hospital" --execute --output ./hospital-25bed

# Expected output:
# - 125 patients
# - 25 providers  
# - 188 claims
# - 300 prescriptions
# - 63 procedures
# - 500 lab_results
```

### Large Medical Center
```bash
# 500-bed regional medical center
./bin/specmint simulate --population "500-bed regional medical center" --execute --output ./hospital-500bed

# Expected output:
# - 2,500 patients
# - 500 providers
# - 3,750 claims
# - 6,000 prescriptions
# - 1,250 procedures
# - 10,000 lab_results
```

### Specialty Clinic
```bash
# Cardiology practice with 10 providers
./bin/specmint simulate --population "cardiology clinic with 10 providers" --execute --output ./cardiology-clinic

# Expected output:
# - 500 patients
# - 10 providers
# - 750 claims (cardiology-focused)
# - 600 prescriptions
# - 250 procedures
# - 1,000 lab_results
```

## Banking & Financial Services

### Community Bank
```bash
# Small community bank
./bin/specmint simulate --population "community bank with 3 branches" --execute --output ./community-bank

# Expected output:
# - 1,500 customers
# - 2,250 accounts
# - 45,000 transactions
# - 150 loans
# - 300 credit_cards
```

### Regional Bank
```bash
# Mid-size regional bank
./bin/specmint simulate --population "regional bank with 25 branches" --execute --output ./regional-bank

# Expected output:
# - 12,500 customers
# - 18,750 accounts
# - 375,000 transactions
# - 1,250 loans
# - 2,500 credit_cards
```

### Credit Union
```bash
# Member-owned credit union
./bin/specmint simulate --population "credit union with 50K members" --execute --output ./credit-union

# Expected output:
# - 50,000 customers
# - 75,000 accounts
# - 1,500,000 transactions
# - 5,000 loans
# - 10,000 credit_cards
```

## E-commerce & Retail

### Startup E-commerce
```bash
# Small online retailer
./bin/specmint simulate --population "e-commerce startup with 5K users" --execute --output ./ecommerce-startup

# Expected output:
# - 5,000 users
# - 500 products
# - 2,500 orders
# - 1,250 reviews
# - 12,500 cart_sessions
```

### Established E-commerce Platform
```bash
# Large e-commerce platform
./bin/specmint simulate --population "e-commerce platform with 100K users" --execute --output ./ecommerce-large

# Expected output:
# - 100,000 users
# - 10,000 products
# - 50,000 orders
# - 25,000 reviews
# - 250,000 cart_sessions
```

### Retail Chain
```bash
# Physical retail stores
./bin/specmint simulate --population "retail chain with 15 stores" --execute --output ./retail-chain

# Expected output:
# - 7,500 products
# - 22,500 customers
# - 45,000 orders
# - 15,000 employees
# - 112,500 inventory_records
```

## Insurance

### Regional Insurance Company
```bash
# Mid-size insurance provider
./bin/specmint simulate --population "insurance company with 25K policyholders" --execute --output ./insurance-regional

# Expected output:
# - 25,000 policyholders
# - 30,000 policies
# - 12,500 claims
# - 2,500 agents
# - 75,000 payments
```

## Advanced Usage

### Analysis Only (No Generation)
```bash
# Analyze scenario without generating data
./bin/specmint simulate --population "100-bed hospital"

# Shows:
# - Recommended record counts
# - Schema dependencies
# - Resource estimates
# - Generation timeline
```

### Save Strategy for Later
```bash
# Save analysis to file for later execution
./bin/specmint simulate --population "retail chain with 20 stores" --save-strategy ./retail-strategy.json

# Execute saved strategy
./bin/specmint simulate --load-strategy ./retail-strategy.json --execute --output ./retail-data
```

### Custom Output Directory
```bash
# Specify custom output location
./bin/specmint simulate --population "community bank with 5 branches" --execute --output /data/bank-simulation
```

### Debug Mode
```bash
# Enable detailed logging
./bin/specmint simulate --population "10-bed hospital" --execute --output ./debug-test --debug
```

## Understanding Output

Each simulation creates:

### Directory Structure
```
output-directory/
├── simulation-manifest.json    # Complete strategy and metadata
├── claims/
│   ├── dataset.jsonl          # Generated records
│   └── manifest.json          # Generation metadata
├── prescriptions/
│   ├── dataset.jsonl
│   └── manifest.json
└── [other-data-types]/
    ├── dataset.jsonl
    └── manifest.json
```

### Simulation Manifest
```json
{
  "scenario": "25-bed community hospital",
  "domain": "hospital",
  "base_unit": "25 beds",
  "total_records": 1201,
  "schemas_generated": 6,
  "execution_time": "45.2s",
  "llm_calls": 240,
  "generation_strategy": {
    "record_counts": {
      "patients": 125,
      "providers": 25,
      "claims": 188,
      "prescriptions": 300,
      "procedures": 63,
      "lab_results": 500
    }
  }
}
```

## Tips for Optimal Results

### Choose Realistic Scenarios
- Use business-relevant scale descriptors
- Include location context when relevant
- Specify the primary business metric (beds, branches, users, stores)

### Performance Considerations
- Start with smaller scenarios for testing
- Large simulations (>100K records) may take several minutes
- Monitor system resources during generation

### Data Quality
- LLM enrichment improves realism but increases generation time
- Use `--seed` parameter for reproducible datasets
- Validate generated data with built-in validation tools

### Troubleshooting
- Use `--debug` flag for detailed logging
- Check `simulation-manifest.json` for execution details
- Missing schemas will create placeholder files
