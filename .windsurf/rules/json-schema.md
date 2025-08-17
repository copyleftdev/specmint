---
trigger: always_on
---

# JSON Schema & Data Generation Rules

## Schema Handling
<schema_rules>
- Support JSON Schema Draft 2020-12
- Validate schemas before generation
- Extract constraints and cross-field rules
- Handle nested objects and arrays efficiently
- Support custom extensions (x-llm, x-cross-field-rules)
- Cache parsed schemas for performance
</schema_rules>

## Deterministic Generation
<deterministic_rules>
- Use seeded RNG for reproducible outputs
- Seed derivation: FNV64(base_seed, field_path, record_index)
- Maintain stable key ordering for objects
- Include optional fields with 90% probability
- Generate arrays with uniform length distribution
- Ensure bitwise-identical outputs for same inputs
</deterministic_rules>

## Type-Specific Generation
<type_generation>
- **strings**: enum > examples > format > pattern > synthetic
- **numbers**: respect min/max/multipleOf constraints
- **arrays**: honor minItems/maxItems bounds
- **objects**: process required fields first, then optional
- **booleans**: use Bernoulli(0.5) distribution
- **formats**: implement date, email, uuid, phone generators
</type_generation>

## LLM Field Markers
<llm_field_rules>
- Fields marked with "x-llm": true get LLM enrichment
- Description starting with "llm:" indicates LLM field
- Preserve original schema constraints after enrichment
- Validate LLM outputs against schema
- Fallback to deterministic values on LLM failure
</llm_field_rules>
