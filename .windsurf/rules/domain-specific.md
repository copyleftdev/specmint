---
trigger: always_on
---

# Domain-Specific Validation Rules

## Healthcare Domain
<healthcare_rules>
- Validate ICD-10 diagnosis codes format and ranges
- Ensure realistic charge amounts ($10 - $50,000)
- Enforce date ordering: DOB <= Service Date <= Submitted Date
- Use synthetic but medically plausible procedure codes
- Validate NPI numbers format (10 digits)
- Ensure HIPAA compliance - no real patient data
</healthcare_rules>

## Fintech Domain
<fintech_rules>
- Validate ABA routing numbers (9 digits with checksum)
- Enforce transaction limits and approval workflows
- Generate realistic but fake account numbers
- Validate currency codes (ISO 4217)
- Implement risk scoring logic (0-100 scale)
- Large transactions (>$10K) must require approval
</fintech_rules>

## E-commerce Domain
<ecommerce_rules>
- Validate SKU formats (e.g., AB123456)
- Ensure price-inventory consistency
- Generate realistic product names and descriptions
- Validate warehouse location codes
- Implement rating distribution logic
- High-value items should have lower inventory
</ecommerce_rules>

## Cross-Field Validation
<cross_field_rules>
- Always validate relationships between related fields
- Implement domain-specific business rules
- Use minimal patching to fix constraint violations
- Log validation failures for analysis
- Maintain referential integrity across records
</cross_field_rules>
