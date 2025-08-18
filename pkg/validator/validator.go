package validator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/specmint/specmint/pkg/schema"
)

// Validator handles record validation and patching
type Validator struct {
	parser *schema.Parser
	rules  []schema.CrossFieldRule
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string      `json:"field"`
	Rule    string      `json:"rule"`
	Message string      `json:"message"`
	Value   interface{} `json:"value"`
}

// New creates a new validator instance
func New(parser *schema.Parser) *Validator {
	rootNode, _ := parser.GetRootNode()
	rules := parser.GetCrossFieldRules(rootNode)

	return &Validator{
		parser: parser,
		rules:  rules,
	}
}

// ValidateRecord validates a record against the schema and cross-field rules
func (v *Validator) ValidateRecord(data map[string]interface{}) []string {
	var errors []string

	// Schema validation
	if err := v.parser.Validate(data); err != nil {
		errors = append(errors, fmt.Sprintf("Schema validation failed: %s", err.Error()))
	}

	// Cross-field rule validation
	for _, rule := range v.rules {
		if err := v.validateCrossFieldRule(data, rule); err != nil {
			errors = append(errors, fmt.Sprintf("Cross-field rule '%s' failed: %s", rule.Name, err.Error()))
		}
	}

	return errors
}

// PatchRecord attempts to fix validation errors in a record
func (v *Validator) PatchRecord(data map[string]interface{}, errors []string) (map[string]interface{}, error) {
	patched := make(map[string]interface{})
	for k, v := range data {
		patched[k] = v
	}

	// Apply patches for cross-field rule violations
	for _, rule := range v.rules {
		if rule.Patch != nil && v.ruleViolated(errors, rule.Name) {
			if err := v.applyPatch(patched, rule.Patch); err != nil {
				return nil, fmt.Errorf("failed to apply patch for rule %s: %w", rule.Name, err)
			}
		}
	}

	return patched, nil
}

// validateCrossFieldRule validates a single cross-field rule
func (v *Validator) validateCrossFieldRule(data map[string]interface{}, rule schema.CrossFieldRule) error {
	switch rule.Rule {
	case "date_ordering":
		return v.validateDateOrdering(data, rule.Fields)
	case "amount_range":
		return v.validateAmountRange(data, rule.Fields)
	case "comparison":
		return v.validateComparison(data, rule.Fields, rule.Constraint)
	case "conditional_required":
		return v.validateConditionalRequired(data, rule.Fields)
	case "mutual_exclusion":
		return v.validateMutualExclusion(data, rule.Fields)
	case "sum_constraint":
		return v.validateSumConstraint(data, rule.Fields)
	default:
		return fmt.Errorf("unknown rule type: %s", rule.Rule)
	}
}

// Domain-specific validation rules

func (v *Validator) validateDateOrdering(data map[string]interface{}, fields []string) error {
	if len(fields) < 2 {
		return fmt.Errorf("date_ordering requires at least 2 fields")
	}

	dates := make([]string, len(fields))
	for i, field := range fields {
		if val, exists := data[field]; exists {
			if dateStr, ok := val.(string); ok {
				dates[i] = dateStr
			} else {
				return fmt.Errorf("field %s is not a string date", field)
			}
		}
	}

	// Check ordering (simplified - assumes ISO date format)
	for i := 1; i < len(dates); i++ {
		if dates[i-1] > dates[i] {
			return fmt.Errorf("date ordering violation: %s (%s) should be <= %s (%s)",
				fields[i-1], dates[i-1], fields[i], dates[i])
		}
	}

	return nil
}

func (v *Validator) validateAmountRange(data map[string]interface{}, fields []string) error {
	if len(fields) != 3 { // amount, min, max
		return fmt.Errorf("amount_range requires exactly 3 fields: amount, min, max")
	}

	amount := v.getNumericValue(data, fields[0])
	min := v.getNumericValue(data, fields[1])
	max := v.getNumericValue(data, fields[2])

	if amount < min || amount > max {
		return fmt.Errorf("amount %f is outside range [%f, %f]", amount, min, max)
	}

	return nil
}

func (v *Validator) validateConditionalRequired(data map[string]interface{}, fields []string) error {
	if len(fields) != 2 {
		return fmt.Errorf("conditional_required requires exactly 2 fields: condition, required")
	}

	conditionField := fields[0]
	requiredField := fields[1]

	if condVal, exists := data[conditionField]; exists {
		// If condition field has a truthy value, required field must exist
		if v.isTruthy(condVal) {
			if _, reqExists := data[requiredField]; !reqExists {
				return fmt.Errorf("field %s is required when %s is present", requiredField, conditionField)
			}
		}
	}

	return nil
}

func (v *Validator) validateMutualExclusion(data map[string]interface{}, fields []string) error {
	presentCount := 0
	var presentFields []string

	for _, field := range fields {
		if _, exists := data[field]; exists {
			presentCount++
			presentFields = append(presentFields, field)
		}
	}

	if presentCount > 1 {
		return fmt.Errorf("mutually exclusive fields present: %s", strings.Join(presentFields, ", "))
	}

	return nil
}

func (v *Validator) validateSumConstraint(data map[string]interface{}, fields []string) error {
	if len(fields) < 3 {
		return fmt.Errorf("sum_constraint requires at least 3 fields: field1, field2, ..., target_sum")
	}

	targetField := fields[len(fields)-1]
	sumFields := fields[:len(fields)-1]

	var sum float64
	for _, field := range sumFields {
		sum += v.getNumericValue(data, field)
	}

	target := v.getNumericValue(data, targetField)
	tolerance := 0.01 // Allow small floating point differences

	if abs(sum-target) > tolerance {
		return fmt.Errorf("sum constraint violation: sum of %s = %f, expected %f",
			strings.Join(sumFields, "+"), sum, target)
	}

	return nil
}

// applyPatch applies a patch rule to fix a validation error
func (v *Validator) applyPatch(data map[string]interface{}, patch *schema.PatchRule) error {
	switch patch.Strategy {
	case "set_value":
		data[patch.Target] = patch.Value
	case "adjust_field":
		return v.adjustField(data, patch)
	case "remove_field":
		delete(data, patch.Target)
	default:
		return fmt.Errorf("unknown patch strategy: %s", patch.Strategy)
	}
	return nil
}

func (v *Validator) adjustField(data map[string]interface{}, patch *schema.PatchRule) error {
	currentVal := v.getNumericValue(data, patch.Target)

	if adjustment, ok := patch.Params["adjustment"].(float64); ok {
		data[patch.Target] = currentVal + adjustment
	} else if factor, ok := patch.Params["factor"].(float64); ok {
		data[patch.Target] = currentVal * factor
	} else {
		return fmt.Errorf("adjust_field requires 'adjustment' or 'factor' parameter")
	}

	return nil
}

// Helper functions

func (v *Validator) getNumericValue(data map[string]interface{}, field string) float64 {
	if val, exists := data[field]; exists {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return 0.0
}

func (v *Validator) isTruthy(val interface{}) bool {
	if val == nil {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int, int64, float64:
		return !reflect.ValueOf(v).IsZero()
	default:
		return true
	}
}

func (v *Validator) ruleViolated(errors []string, ruleName string) bool {
	for _, err := range errors {
		if strings.Contains(err, fmt.Sprintf("rule '%s'", ruleName)) {
			return true
		}
	}
	return false
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// validateComparison validates comparison constraints between fields
func (v *Validator) validateComparison(data map[string]interface{}, fields []string, constraint string) error {
	if len(fields) < 2 {
		return fmt.Errorf("comparison rule requires at least 2 fields")
	}

	// Parse constraint to extract left side, operator, and right side
	constraint = strings.TrimSpace(constraint)
	
	var leftSide, operator, rightSide string
	
	// Find the operator
	if strings.Contains(constraint, ">=") {
		parts := strings.Split(constraint, ">=")
		if len(parts) != 2 {
			return fmt.Errorf("invalid constraint format: %s", constraint)
		}
		leftSide = strings.TrimSpace(parts[0])
		operator = ">="
		rightSide = strings.TrimSpace(parts[1])
	} else if strings.Contains(constraint, "<=") {
		parts := strings.Split(constraint, "<=")
		if len(parts) != 2 {
			return fmt.Errorf("invalid constraint format: %s", constraint)
		}
		leftSide = strings.TrimSpace(parts[0])
		operator = "<="
		rightSide = strings.TrimSpace(parts[1])
	} else if strings.Contains(constraint, ">") {
		parts := strings.Split(constraint, ">")
		if len(parts) != 2 {
			return fmt.Errorf("invalid constraint format: %s", constraint)
		}
		leftSide = strings.TrimSpace(parts[0])
		operator = ">"
		rightSide = strings.TrimSpace(parts[1])
	} else if strings.Contains(constraint, "<") {
		parts := strings.Split(constraint, "<")
		if len(parts) != 2 {
			return fmt.Errorf("invalid constraint format: %s", constraint)
		}
		leftSide = strings.TrimSpace(parts[0])
		operator = "<"
		rightSide = strings.TrimSpace(parts[1])
	} else {
		return fmt.Errorf("unsupported comparison operator in constraint: %s", constraint)
	}
	
	// Evaluate left side
	leftValue := v.evaluateExpression(data, leftSide)
	
	// Evaluate right side
	rightValue := v.evaluateExpression(data, rightSide)
	
	// Apply comparison
	switch operator {
	case ">=":
		if leftValue < rightValue {
			return fmt.Errorf("constraint violation: %s (%.6f) should be >= %s (%.6f)", leftSide, leftValue, rightSide, rightValue)
		}
	case "<=":
		if leftValue > rightValue {
			return fmt.Errorf("constraint violation: %s (%.6f) should be <= %s (%.6f)", leftSide, leftValue, rightSide, rightValue)
		}
	case ">":
		if leftValue <= rightValue {
			return fmt.Errorf("constraint violation: %s (%.6f) should be > %s (%.6f)", leftSide, leftValue, rightSide, rightValue)
		}
	case "<":
		if leftValue >= rightValue {
			return fmt.Errorf("constraint violation: %s (%.6f) should be < %s (%.6f)", leftSide, leftValue, rightSide, rightValue)
		}
	}

	return nil
}

// evaluateExpression evaluates a mathematical expression with field references
func (v *Validator) evaluateExpression(data map[string]interface{}, expr string) float64 {
	expr = strings.TrimSpace(expr)
	
	// Handle simple field reference
	if !strings.Contains(expr, "+") && !strings.Contains(expr, "-") && !strings.Contains(expr, "*") && !strings.Contains(expr, "/") {
		return v.getNumericValue(data, expr)
	}
	
	// Handle addition (most common case for medical constraints)
	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		result := 0.0
		for _, part := range parts {
			part = strings.TrimSpace(part)
			result += v.getNumericValue(data, part)
		}
		return result
	}
	
	// Handle subtraction
	if strings.Contains(expr, "-") {
		parts := strings.Split(expr, "-")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			return v.getNumericValue(data, left) - v.getNumericValue(data, right)
		}
	}
	
	// Handle multiplication
	if strings.Contains(expr, "*") {
		parts := strings.Split(expr, "*")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			return v.getNumericValue(data, left) * v.getNumericValue(data, right)
		}
	}
	
	// Fallback: treat as field name
	return v.getNumericValue(data, expr)
}
