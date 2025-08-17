package validator

import (
	"testing"
)

func TestHL7ValidationRules_ValidateICD10Code(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"Valid ICD-10 code", "I10", false},
		{"Valid ICD-10 with decimal", "E11.9", false},
		{"Valid ICD-10 with full decimal", "I25.10", false},
		{"Invalid format - starts with number", "1I10", true},
		{"Invalid format - no decimal after 3 chars", "I109", true},
		{"Reserved U code", "U07.1", true},
		{"Valid Z code", "Z00.00", false},
		{"Valid external cause", "V12.34", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateICD10Code(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateICD10Code() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHL7ValidationRules_ValidateCPTCode(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"Valid CPT code", "99213", false},
		{"Valid surgery code", "27447", false},
		{"Valid radiology code", "71020", false},
		{"Valid lab code", "80053", false},
		{"Invalid - too short", "9921", true},
		{"Invalid - too long", "992134", true},
		{"Invalid - non-numeric", "9921A", true},
		{"Invalid - out of range", "00050", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCPTCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCPTCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHL7ValidationRules_ValidateNPI(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name    string
		npi     string
		wantErr bool
	}{
		{"Valid NPI", "1234567893", false}, // Test NPI - skip checksum validation for testing
		{"Invalid - too short", "123456789", true},
		{"Invalid - too long", "12345678901", true},
		{"Invalid - non-numeric", "123456789A", true},
		{"Invalid - bad checksum", "1234567890", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateNPI(tt.npi)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNPI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHL7ValidationRules_ValidateSSN(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name    string
		ssn     string
		wantErr bool
	}{
		{"Valid SSN", "555123456", false},
		{"Invalid - all zeros", "000000000", true},
		{"Invalid - sequential pattern", "123456789", true}, // This is actually invalid pattern
		{"Invalid - area 000", "000123456", true},
		{"Invalid - area 666", "666123456", true},
		{"Invalid - area 900", "900123456", true},
		{"Invalid - too short", "12345678", true},
		{"Invalid - non-numeric", "12345678A", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSSN(tt.ssn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSSN() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHL7ValidationRules_ValidateHL7DateTime(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name     string
		dateTime string
		wantErr  bool
	}{
		{"Valid date YYYYMMDD", "20231215", false},
		{"Valid datetime YYYYMMDDHHMM", "202312151430", false},
		{"Valid datetime YYYYMMDDHHMMSS", "20231215143045", false},
		{"Invalid - wrong format", "2023-12-15", true},
		{"Invalid - bad date", "20231315", true},
		{"Invalid - bad time", "20231215256045", true},
		{"Invalid - too short", "2023121", true},
		{"Invalid - too long", "202312151430456", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateHL7DateTime(tt.dateTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHL7DateTime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHL7ValidationRules_ValidateFHIRReference(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name      string
		reference string
		wantErr   bool
	}{
		{"Valid Patient reference", "Patient/123", false},
		{"Valid Practitioner reference", "Practitioner/ABC-123", false},
		{"Valid Organization reference", "Organization/org.123", false},
		{"Invalid - no slash", "Patient123", true},
		{"Invalid - multiple slashes", "Patient/123/456", true},
		{"Invalid - invalid resource type", "InvalidType/123", true},
		{"Invalid - empty ID", "Patient/", true},
		{"Invalid - ID too long", "Patient/" + string(make([]byte, 65)), true},
		{"Invalid - invalid ID chars", "Patient/123@456", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFHIRReference(tt.reference)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFHIRReference() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHL7ValidationRules_ValidateHL7v2MessageType(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name         string
		msgType      string
		triggerEvent string
		wantErr      bool
	}{
		{"Valid ADT A01", "ADT", "A01", false},
		{"Valid ORM O01", "ORM", "O01", false},
		{"Valid ORU R01", "ORU", "R01", false},
		{"Invalid message type", "INVALID", "A01", true},
		{"Invalid trigger event", "ADT", "X99", true},
		{"Wrong combination", "ORM", "A01", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateHL7v2MessageType(tt.msgType, tt.triggerEvent)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHL7v2MessageType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHL7ValidationRules_ValidateHL7v2FieldSeparators(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name          string
		fieldSep      string
		encodingChars string
		wantErr       bool
	}{
		{"Valid separators", "|", "^~\\&", false},
		{"Invalid field separator", "!", "^~\\&", true},
		{"Invalid encoding chars", "|", "^~!&", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateHL7v2FieldSeparators(tt.fieldSep, tt.encodingChars)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHL7v2FieldSeparators() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHL7ValidationRules_ValidateCrossFieldRules(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name    string
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name: "Valid date order",
			data: map[string]interface{}{
				"birth_date":     "19900101",
				"admission_date": "20231215",
			},
			wantErr: false,
		},
		{
			name: "Invalid date order - birth after admission",
			data: map[string]interface{}{
				"birth_date":     "20231215",
				"admission_date": "19900101",
			},
			wantErr: true,
		},
		{
			name: "Final result without value",
			data: map[string]interface{}{
				"result_status":     "F",
				"observation_value": "",
			},
			wantErr: true,
		},
		{
			name: "Final result with value",
			data: map[string]interface{}{
				"result_status":     "F",
				"observation_value": "Normal",
			},
			wantErr: false,
		},
		{
			name: "Cancelled order without response flag",
			data: map[string]interface{}{
				"order_status":  "CA",
				"response_flag": "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateCrossFieldRules(tt.data)
			hasErr := len(errors) > 0
			if hasErr != tt.wantErr {
				t.Errorf("ValidateCrossFieldRules() errors = %v, wantErr %v", errors, tt.wantErr)
			}
		})
	}
}

func TestHL7ValidationRules_ValidateGender(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name    string
		gender  string
		wantErr bool
	}{
		{"Valid Male", "M", false},
		{"Valid Female", "F", false},
		{"Valid Other", "O", false},
		{"Valid Unknown", "U", false},
		{"Invalid gender", "X", true},
		{"Invalid lowercase", "m", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateGender(tt.gender)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGender() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHL7ValidationRules_ValidatePatientClass(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name    string
		class   string
		wantErr bool
	}{
		{"Valid Emergency", "E", false},
		{"Valid Inpatient", "I", false},
		{"Valid Outpatient", "O", false},
		{"Invalid class", "X", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePatientClass(tt.class)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePatientClass() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHL7ValidationRules_ValidateObservationStatus(t *testing.T) {
	validator := NewHL7ValidationRules()

	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"Valid Final", "F", false},
		{"Valid Preliminary", "P", false},
		{"Valid Corrected", "C", false},
		{"Invalid status", "Z", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateObservationStatus(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateObservationStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
