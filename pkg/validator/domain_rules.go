package validator

import (
	"fmt"
	"regexp"
	"time"
)

// DomainValidator provides domain-specific validation rules
type DomainValidator struct {
	rules map[string][]ValidationRule
}

// ValidationRule represents a domain-specific validation rule
type ValidationRule struct {
	Name        string
	Description string
	Validator   func(data map[string]interface{}) error
	Severity    string
}

// NewDomainValidator creates a new domain validator with built-in rules
func NewDomainValidator() *DomainValidator {
	dv := &DomainValidator{
		rules: make(map[string][]ValidationRule),
	}
	
	dv.registerHealthcareRules()
	dv.registerFintechRules()
	dv.registerEcommerceRules()
	
	return dv
}

// ValidateDomain validates data against domain-specific rules
func (dv *DomainValidator) ValidateDomain(domain string, data map[string]interface{}) []error {
	var errors []error
	
	if rules, exists := dv.rules[domain]; exists {
		for _, rule := range rules {
			if err := rule.Validator(data); err != nil {
				errors = append(errors, fmt.Errorf("[%s] %s: %v", rule.Severity, rule.Name, err))
			}
		}
	}
	
	return errors
}

// Healthcare domain validation rules
func (dv *DomainValidator) registerHealthcareRules() {
	dv.rules["healthcare"] = []ValidationRule{
		{
			Name:        "icd10_format",
			Description: "Validate ICD-10 diagnosis codes format",
			Severity:    "error",
			Validator: func(data map[string]interface{}) error {
				if clinical, ok := data["clinical_data"].(map[string]interface{}); ok {
					if diagnoses, ok := clinical["diagnoses"].([]interface{}); ok {
						for _, diag := range diagnoses {
							if diagMap, ok := diag.(map[string]interface{}); ok {
								if code, ok := diagMap["icd10_code"].(string); ok {
									if !isValidICD10(code) {
										return fmt.Errorf("invalid ICD-10 code: %s", code)
									}
								}
							}
						}
					}
				}
				return nil
			},
		},
		{
			Name:        "charge_amount_realistic",
			Description: "Ensure realistic charge amounts ($10 - $50,000)",
			Severity:    "error",
			Validator: func(data map[string]interface{}) error {
				if billing, ok := data["billing"].(map[string]interface{}); ok {
					if total, ok := billing["total_charges"].(float64); ok {
						if total < 10 || total > 50000 {
							return fmt.Errorf("unrealistic charge amount: $%.2f", total)
						}
					}
				}
				return nil
			},
		},
		{
			Name:        "date_ordering",
			Description: "DOB <= Service Date <= Submitted Date",
			Severity:    "error",
			Validator: func(data map[string]interface{}) error {
				var dob, serviceDate, submittedDate time.Time
				var err error
				
				// Extract dates
				if patient, ok := data["patient"].(map[string]interface{}); ok {
					if demo, ok := patient["demographics"].(map[string]interface{}); ok {
						if dobStr, ok := demo["date_of_birth"].(string); ok {
							dob, err = time.Parse("2006-01-02", dobStr)
							if err != nil {
								return fmt.Errorf("invalid DOB format: %s", dobStr)
							}
						}
					}
				}
				
				if billing, ok := data["billing"].(map[string]interface{}); ok {
					if serviceStr, ok := billing["service_date"].(string); ok {
						serviceDate, err = time.Parse("2006-01-02", serviceStr)
						if err != nil {
							return fmt.Errorf("invalid service date format: %s", serviceStr)
						}
					}
					if subStr, ok := billing["submitted_date"].(string); ok {
						submittedDate, err = time.Parse("2006-01-02", subStr)
						if err != nil {
							return fmt.Errorf("invalid submitted date format: %s", subStr)
						}
					}
				}
				
				// Validate ordering
				if !dob.IsZero() && !serviceDate.IsZero() && dob.After(serviceDate) {
					return fmt.Errorf("DOB (%s) cannot be after service date (%s)", dob.Format("2006-01-02"), serviceDate.Format("2006-01-02"))
				}
				
				if !serviceDate.IsZero() && !submittedDate.IsZero() && serviceDate.After(submittedDate) {
					return fmt.Errorf("service date (%s) cannot be after submitted date (%s)", serviceDate.Format("2006-01-02"), submittedDate.Format("2006-01-02"))
				}
				
				return nil
			},
		},
		{
			Name:        "npi_format",
			Description: "Validate NPI numbers format (10 digits)",
			Severity:    "error",
			Validator: func(data map[string]interface{}) error {
				if encounter, ok := data["encounter"].(map[string]interface{}); ok {
					if provider, ok := encounter["provider"].(map[string]interface{}); ok {
						if npi, ok := provider["npi"].(string); ok {
							if !isValidNPI(npi) {
								return fmt.Errorf("invalid NPI format: %s", npi)
							}
						}
					}
				}
				return nil
			},
		},
		{
			Name:        "vital_signs_plausible",
			Description: "Blood pressure values should be medically plausible",
			Severity:    "warning",
			Validator: func(data map[string]interface{}) error {
				if clinical, ok := data["clinical_data"].(map[string]interface{}); ok {
					if vitals, ok := clinical["vital_signs"].(map[string]interface{}); ok {
						if bp, ok := vitals["blood_pressure"].(map[string]interface{}); ok {
							if sys, ok := bp["systolic"].(float64); ok {
								if dia, ok := bp["diastolic"].(float64); ok {
									if sys <= dia {
										return fmt.Errorf("systolic (%v) must be greater than diastolic (%v)", sys, dia)
									}
									if (sys - dia) < 20 {
										return fmt.Errorf("pulse pressure too narrow: %v", sys-dia)
									}
								}
							}
						}
					}
				}
				return nil
			},
		},
	}
}

// Fintech domain validation rules
func (dv *DomainValidator) registerFintechRules() {
	dv.rules["fintech"] = []ValidationRule{
		{
			Name:        "routing_number_checksum",
			Description: "Validate ABA routing numbers (9 digits with checksum)",
			Severity:    "error",
			Validator: func(data map[string]interface{}) error {
				if account, ok := data["account"].(map[string]interface{}); ok {
					if routing, ok := account["routing_number"].(string); ok {
						if !isValidRoutingNumber(routing) {
							return fmt.Errorf("invalid routing number: %s", routing)
						}
					}
				}
				return nil
			},
		},
		{
			Name:        "currency_code_iso",
			Description: "Validate currency codes (ISO 4217)",
			Severity:    "error",
			Validator: func(data map[string]interface{}) error {
				if txn, ok := data["transaction_details"].(map[string]interface{}); ok {
					if currency, ok := txn["currency"].(string); ok {
						if !isValidCurrencyCode(currency) {
							return fmt.Errorf("invalid currency code: %s", currency)
						}
					}
				}
				return nil
			},
		},
		{
			Name:        "risk_score_range",
			Description: "Risk scoring logic (0-100 scale)",
			Severity:    "error",
			Validator: func(data map[string]interface{}) error {
				if risk, ok := data["risk_assessment"].(map[string]interface{}); ok {
					if score, ok := risk["risk_score"].(float64); ok {
						if score < 0 || score > 100 {
							return fmt.Errorf("risk score out of range: %v", score)
						}
					}
				}
				return nil
			},
		},
		{
			Name:        "large_transaction_approval",
			Description: "Large transactions (>$10K) must require approval",
			Severity:    "error",
			Validator: func(data map[string]interface{}) error {
				if txn, ok := data["transaction_details"].(map[string]interface{}); ok {
					if amount, ok := txn["amount"].(float64); ok {
						if amount > 10000 {
							if risk, ok := data["risk_assessment"].(map[string]interface{}); ok {
								if status, ok := risk["approval_status"].(string); ok {
									if status != "manual_review" && status != "approved" {
										return fmt.Errorf("large transaction ($%.2f) requires approval, got status: %s", amount, status)
									}
								}
							}
						}
					}
				}
				return nil
			},
		},
	}
}

// E-commerce domain validation rules
func (dv *DomainValidator) registerEcommerceRules() {
	dv.rules["ecommerce"] = []ValidationRule{
		{
			Name:        "sku_format",
			Description: "Validate SKU formats (e.g., AB123456)",
			Severity:    "error",
			Validator: func(data map[string]interface{}) error {
				if sku, ok := data["sku"].(string); ok {
					if !isValidSKU(sku) {
						return fmt.Errorf("invalid SKU format: %s", sku)
					}
				}
				return nil
			},
		},
		{
			Name:        "price_inventory_consistency",
			Description: "Ensure price-inventory consistency",
			Severity:    "warning",
			Validator: func(data map[string]interface{}) error {
				if pricing, ok := data["pricing"].(map[string]interface{}); ok {
					if inventory, ok := data["inventory"].(map[string]interface{}); ok {
						if basePrice, ok := pricing["base_price"].(float64); ok {
							if stock, ok := inventory["stock_quantity"].(float64); ok {
								// High-value items should have lower inventory
								if basePrice > 1000 && stock > 1000 {
									return fmt.Errorf("high-value item ($%.2f) should have lower inventory (%v)", basePrice, stock)
								}
							}
						}
					}
				}
				return nil
			},
		},
		{
			Name:        "sale_price_validation",
			Description: "Sale price must be less than base price",
			Severity:    "error",
			Validator: func(data map[string]interface{}) error {
				if pricing, ok := data["pricing"].(map[string]interface{}); ok {
					if basePrice, ok := pricing["base_price"].(float64); ok {
						if salePrice, ok := pricing["sale_price"].(float64); ok {
							if salePrice >= basePrice {
								return fmt.Errorf("sale price ($%.2f) must be less than base price ($%.2f)", salePrice, basePrice)
							}
						}
					}
				}
				return nil
			},
		},
		{
			Name:        "warehouse_location_format",
			Description: "Validate warehouse location codes",
			Severity:    "error",
			Validator: func(data map[string]interface{}) error {
				if inventory, ok := data["inventory"].(map[string]interface{}); ok {
					if location, ok := inventory["warehouse_location"].(string); ok {
						if !isValidWarehouseLocation(location) {
							return fmt.Errorf("invalid warehouse location format: %s", location)
						}
					}
				}
				return nil
			},
		},
	}
}

// Helper validation functions
func isValidICD10(code string) bool {
	// ICD-10 format: Letter followed by 2 digits, optionally followed by decimal and 1-4 more digits/X
	pattern := `^[A-Z][0-9]{2}(\.[0-9X]{1,4})?$`
	matched, _ := regexp.MatchString(pattern, code)
	return matched
}

func isValidNPI(npi string) bool {
	// NPI must be exactly 10 digits
	if len(npi) != 10 {
		return false
	}
	for _, char := range npi {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func isValidRoutingNumber(routing string) bool {
	// ABA routing number must be exactly 9 digits
	if len(routing) != 9 {
		return false
	}
	
	// Check if all characters are digits
	for _, char := range routing {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	// Simple checksum validation (ABA algorithm)
	sum := 0
	for i, char := range routing {
		digit := int(char - '0')
		weight := []int{3, 7, 1, 3, 7, 1, 3, 7, 1}[i]
		sum += digit * weight
	}
	
	return sum%10 == 0
}

func isValidCurrencyCode(currency string) bool {
	// ISO 4217 currency codes
	validCurrencies := map[string]bool{
		"USD": true, "EUR": true, "GBP": true, "CAD": true, "AUD": true,
		"JPY": true, "CHF": true, "CNY": true, "INR": true, "BRL": true,
	}
	return validCurrencies[currency]
}

func isValidSKU(sku string) bool {
	// SKU format: 2 letters followed by 6 digits
	pattern := `^[A-Z]{2}[0-9]{6}$`
	matched, _ := regexp.MatchString(pattern, sku)
	return matched
}

func isValidWarehouseLocation(location string) bool {
	// Warehouse location format: XX-XXX-999
	pattern := `^[A-Z]{2}-[A-Z]{3}-[0-9]{3}$`
	matched, _ := regexp.MatchString(pattern, location)
	return matched
}
