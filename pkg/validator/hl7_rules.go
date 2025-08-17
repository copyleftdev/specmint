package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// HL7ValidationRules implements healthcare interoperability validation
type HL7ValidationRules struct {
	icd10Regex   *regexp.Regexp
	cptRegex     *regexp.Regexp
	npiRegex     *regexp.Regexp
	ssnRegex     *regexp.Regexp
	phoneRegex   *regexp.Regexp
	zipRegex     *regexp.Regexp
	hl7DateRegex *regexp.Regexp
}

// NewHL7ValidationRules creates a new HL7 validator
func NewHL7ValidationRules() *HL7ValidationRules {
	return &HL7ValidationRules{
		icd10Regex:   regexp.MustCompile(`^[A-TV-Z]\d{2}(\.\d{1,3})?$`),
		cptRegex:     regexp.MustCompile(`^\d{5}$`),
		npiRegex:     regexp.MustCompile(`^\d{10}$`),
		ssnRegex:     regexp.MustCompile(`^\d{9}$`),
		phoneRegex:   regexp.MustCompile(`^\d{10}$`),
		zipRegex:     regexp.MustCompile(`^\d{5}(-\d{4})?$`),
		hl7DateRegex: regexp.MustCompile(`^\d{8,14}$`),
	}
}

// ValidateICD10Code validates ICD-10-CM diagnosis codes
func (h *HL7ValidationRules) ValidateICD10Code(code string) error {
	if !h.icd10Regex.MatchString(code) {
		return fmt.Errorf("invalid ICD-10 code format: %s", code)
	}

	// Validate specific ICD-10 business rules
	if strings.HasPrefix(code, "U") {
		return fmt.Errorf("U codes are reserved for WHO use: %s", code)
	}

	// External cause codes cannot be primary diagnosis
	if strings.HasPrefix(code, "V") || strings.HasPrefix(code, "W") || 
	   strings.HasPrefix(code, "X") || strings.HasPrefix(code, "Y") {
		// This would need context to validate properly
	}

	return nil
}

// ValidateCPTCode validates CPT procedure codes
func (h *HL7ValidationRules) ValidateCPTCode(code string) error {
	if !h.cptRegex.MatchString(code) {
		return fmt.Errorf("invalid CPT code format: %s", code)
	}

	codeNum, err := strconv.Atoi(code)
	if err != nil {
		return fmt.Errorf("CPT code must be numeric: %s", code)
	}

	// Validate CPT code ranges
	if codeNum >= 100 && codeNum <= 99499 {
		// Category I codes
		return nil
	} else if strings.HasSuffix(code, "F") {
		// Category II codes (performance measurement)
		return nil
	} else if strings.HasSuffix(code, "T") {
		// Category III codes (emerging technology)
		return nil
	}

	return fmt.Errorf("CPT code outside valid ranges: %s", code)
}

// ValidateNPI validates National Provider Identifier
func (h *HL7ValidationRules) ValidateNPI(npi string) error {
	if !h.npiRegex.MatchString(npi) {
		return fmt.Errorf("invalid NPI format: %s", npi)
	}

	// Luhn algorithm check for NPI
	return h.validateLuhnChecksum(npi)
}

// ValidateSSN validates Social Security Number
func (h *HL7ValidationRules) ValidateSSN(ssn string) error {
	if !h.ssnRegex.MatchString(ssn) {
		return fmt.Errorf("invalid SSN format: %s", ssn)
	}

	// Check for invalid SSN patterns
	if ssn == "000000000" || ssn == "123456789" {
		return fmt.Errorf("invalid SSN pattern: %s", ssn)
	}

	// Area number cannot be 000, 666, or 900-999
	area := ssn[:3]
	if area == "000" || area == "666" || (area >= "900" && area <= "999") {
		return fmt.Errorf("invalid SSN area number: %s", area)
	}

	return nil
}

// ValidateHL7DateTime validates HL7 date/time formats
func (h *HL7ValidationRules) ValidateHL7DateTime(dateTime string) error {
	if !h.hl7DateRegex.MatchString(dateTime) {
		return fmt.Errorf("invalid HL7 date/time format: %s", dateTime)
	}

	// Parse based on length
	var layout string
	switch len(dateTime) {
	case 8: // YYYYMMDD
		layout = "20060102"
	case 12: // YYYYMMDDHHMM
		layout = "200601021504"
	case 14: // YYYYMMDDHHMMSS
		layout = "20060102150405"
	default:
		return fmt.Errorf("unsupported HL7 date/time length: %d", len(dateTime))
	}

	_, err := time.Parse(layout, dateTime)
	if err != nil {
		return fmt.Errorf("invalid HL7 date/time value: %s", dateTime)
	}

	return nil
}

// ValidateFHIRReference validates FHIR resource references
func (h *HL7ValidationRules) ValidateFHIRReference(reference string) error {
	// FHIR reference format: ResourceType/id
	parts := strings.Split(reference, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid FHIR reference format: %s", reference)
	}

	resourceType := parts[0]
	resourceId := parts[1]

	// Validate resource type
	validResourceTypes := map[string]bool{
		"Patient": true, "Practitioner": true, "Organization": true,
		"Encounter": true, "Observation": true, "DiagnosticReport": true,
		"Condition": true, "Procedure": true, "MedicationRequest": true,
		"AllergyIntolerance": true, "CarePlan": true, "Goal": true,
	}

	if !validResourceTypes[resourceType] {
		return fmt.Errorf("invalid FHIR resource type: %s", resourceType)
	}

	// Validate resource ID format (alphanumeric, dash, dot)
	idRegex := regexp.MustCompile(`^[A-Za-z0-9\-\.]{1,64}$`)
	if !idRegex.MatchString(resourceId) {
		return fmt.Errorf("invalid FHIR resource ID: %s", resourceId)
	}

	return nil
}

// ValidateHL7v2MessageType validates HL7 v2.x message types
func (h *HL7ValidationRules) ValidateHL7v2MessageType(msgType, triggerEvent string) error {
	validCombinations := map[string][]string{
		"ADT": {"A01", "A02", "A03", "A04", "A05", "A06", "A07", "A08", "A09", "A10", "A11", "A12", "A13", "A14", "A15", "A16", "A17", "A18", "A19", "A20", "A21", "A22", "A23", "A24", "A25", "A26", "A27", "A28", "A29", "A30", "A31", "A32", "A33", "A34", "A35", "A36", "A37", "A38", "A39", "A40", "A41", "A42", "A43", "A44", "A45", "A46", "A47", "A48", "A49", "A50", "A51", "A52", "A53", "A54", "A55", "A60", "A61", "A62"},
		"ORM": {"O01", "O02", "O03"},
		"ORU": {"R01", "R02", "R03", "R04", "R30", "R31", "R32"},
		"OML": {"O21", "O22", "O23", "O24", "O33", "O34", "O35", "O36"},
		"SIU": {"S12", "S13", "S14", "S15", "S16", "S17", "S18", "S19", "S20", "S21", "S22", "S23", "S24", "S25", "S26"},
		"MDM": {"T01", "T02", "T03", "T04", "T05", "T06", "T07", "T08", "T09", "T10", "T11"},
		"BAR": {"P01", "P02", "P05", "P06", "P10", "P12"},
		"DFT": {"P03", "P11"},
		"QRY": {"A19", "P04", "PC4", "PC6", "PC7", "PC8", "R02"},
	}

	validEvents, exists := validCombinations[msgType]
	if !exists {
		return fmt.Errorf("invalid HL7 v2.x message type: %s", msgType)
	}

	for _, event := range validEvents {
		if event == triggerEvent {
			return nil
		}
	}

	return fmt.Errorf("invalid trigger event %s for message type %s", triggerEvent, msgType)
}

// ValidateHL7v2FieldSeparators validates HL7 v2.x encoding characters
func (h *HL7ValidationRules) ValidateHL7v2FieldSeparators(fieldSep, encodingChars string) error {
	if fieldSep != "|" {
		return fmt.Errorf("invalid field separator: expected '|', got '%s'", fieldSep)
	}

	if encodingChars != "^~\\&" {
		return fmt.Errorf("invalid encoding characters: expected '^~\\&', got '%s'", encodingChars)
	}

	return nil
}

// ValidateMedicalRecordNumber validates medical record number formats
func (h *HL7ValidationRules) ValidateMedicalRecordNumber(mrn string) error {
	// MRN should be alphanumeric, 1-20 characters
	mrnRegex := regexp.MustCompile(`^[A-Za-z0-9]{1,20}$`)
	if !mrnRegex.MatchString(mrn) {
		return fmt.Errorf("invalid medical record number format: %s", mrn)
	}

	return nil
}

// ValidatePhoneNumber validates phone number formats
func (h *HL7ValidationRules) ValidatePhoneNumber(phone string) error {
	if !h.phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone number format: %s", phone)
	}

	// Check for invalid patterns
	if phone == "0000000000" || phone == "1111111111" {
		return fmt.Errorf("invalid phone number pattern: %s", phone)
	}

	return nil
}

// ValidateZipCode validates US ZIP code formats
func (h *HL7ValidationRules) ValidateZipCode(zip string) error {
	if !h.zipRegex.MatchString(zip) {
		return fmt.Errorf("invalid ZIP code format: %s", zip)
	}

	return nil
}

// ValidateGender validates gender codes
func (h *HL7ValidationRules) ValidateGender(gender string) error {
	validGenders := map[string]bool{
		"M": true, "F": true, "O": true, "U": true, "A": true, "N": true,
	}

	if !validGenders[gender] {
		return fmt.Errorf("invalid gender code: %s", gender)
	}

	return nil
}

// ValidateMaritalStatus validates marital status codes
func (h *HL7ValidationRules) ValidateMaritalStatus(status string) error {
	validStatuses := map[string]bool{
		"A": true, "D": true, "I": true, "L": true, "M": true,
		"P": true, "S": true, "T": true, "U": true, "W": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid marital status code: %s", status)
	}

	return nil
}

// ValidatePatientClass validates patient class codes
func (h *HL7ValidationRules) ValidatePatientClass(class string) error {
	validClasses := map[string]bool{
		"E": true, // Emergency
		"I": true, // Inpatient
		"O": true, // Outpatient
		"P": true, // Preadmit
		"R": true, // Recurring patient
		"B": true, // Obstetrics
		"C": true, // Commercial Account
		"N": true, // Not Applicable
	}

	if !validClasses[class] {
		return fmt.Errorf("invalid patient class: %s", class)
	}

	return nil
}

// ValidateObservationStatus validates observation result status
func (h *HL7ValidationRules) ValidateObservationStatus(status string) error {
	validStatuses := map[string]bool{
		"C": true, // Corrected
		"D": true, // Deleted
		"F": true, // Final
		"I": true, // Specimen in lab
		"N": true, // Not asked
		"O": true, // Order received
		"P": true, // Preliminary
		"R": true, // Results entered
		"S": true, // Partial
		"U": true, // Results status change
		"W": true, // Post original as wrong
		"X": true, // Results cannot be obtained
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid observation status: %s", status)
	}

	return nil
}

// validateLuhnChecksum validates checksum using Luhn algorithm
func (h *HL7ValidationRules) validateLuhnChecksum(number string) error {
	sum := 0
	alternate := false

	// Process digits from right to left
	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return fmt.Errorf("non-numeric character in number: %s", number)
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
		alternate = !alternate
	}

	if sum%10 != 0 {
		return fmt.Errorf("invalid checksum for number: %s", number)
	}

	return nil
}

// ValidateCrossFieldRules validates business rules across multiple fields
func (h *HL7ValidationRules) ValidateCrossFieldRules(data map[string]interface{}) []error {
	var errors []error

	// Birth date must be before admission date
	if birthDate, ok := data["birth_date"].(string); ok {
		if admissionDate, ok := data["admission_date"].(string); ok {
			if err := h.validateDateOrder(birthDate, admissionDate, "birth_date", "admission_date"); err != nil {
				errors = append(errors, err)
			}
		}
	}

	// Admission date must be before or equal to discharge date
	if admissionDate, ok := data["admission_date"].(string); ok {
		if dischargeDate, ok := data["discharge_date"].(string); ok {
			if err := h.validateDateOrder(admissionDate, dischargeDate, "admission_date", "discharge_date"); err != nil {
				errors = append(errors, err)
			}
		}
	}

	// Observation effective time must be before issued time
	if effectiveTime, ok := data["effective_time"].(string); ok {
		if issuedTime, ok := data["issued_time"].(string); ok {
			if err := h.validateDateOrder(effectiveTime, issuedTime, "effective_time", "issued_time"); err != nil {
				errors = append(errors, err)
			}
		}
	}

	// Final results must have values
	if status, ok := data["result_status"].(string); ok {
		if status == "F" || status == "final" {
			if value, ok := data["observation_value"]; !ok || value == "" {
				errors = append(errors, fmt.Errorf("final results must have observation values"))
			}
		}
	}

	// Cancelled orders must have response flag
	if orderStatus, ok := data["order_status"].(string); ok {
		if orderStatus == "CA" {
			if responseFlag, ok := data["response_flag"]; !ok || responseFlag == "" {
				errors = append(errors, fmt.Errorf("cancelled orders must have response flag"))
			}
		}
	}

	return errors
}

// validateDateOrder validates that date1 <= date2
func (h *HL7ValidationRules) validateDateOrder(date1, date2, field1, field2 string) error {
	if err := h.ValidateHL7DateTime(date1); err != nil {
		return fmt.Errorf("invalid %s: %v", field1, err)
	}

	if err := h.ValidateHL7DateTime(date2); err != nil {
		return fmt.Errorf("invalid %s: %v", field2, err)
	}

	// Parse dates for comparison
	layout1 := h.getDateLayout(date1)
	layout2 := h.getDateLayout(date2)

	time1, err := time.Parse(layout1, date1)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %v", field1, err)
	}

	time2, err := time.Parse(layout2, date2)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %v", field2, err)
	}

	if time1.After(time2) {
		return fmt.Errorf("%s (%s) must be before or equal to %s (%s)", field1, date1, field2, date2)
	}

	return nil
}

// getDateLayout returns appropriate time layout based on date string length
func (h *HL7ValidationRules) getDateLayout(date string) string {
	switch len(date) {
	case 8:
		return "20060102"
	case 12:
		return "200601021504"
	case 14:
		return "20060102150405"
	default:
		return "20060102150405" // default
	}
}
