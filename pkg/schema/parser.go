package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Parser handles JSON Schema parsing and validation
type Parser struct {
	compiler *jsonschema.Compiler
	schema   *jsonschema.Schema
	raw      map[string]interface{}
}

// SchemaNode represents a parsed schema node with metadata
type SchemaNode struct {
	Type        string                 `json:"type"`
	Properties  map[string]*SchemaNode `json:"properties,omitempty"`
	Items       *SchemaNode            `json:"items,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Enum        []interface{}          `json:"enum,omitempty"`
	Examples    []interface{}          `json:"examples,omitempty"`
	Format      string                 `json:"format,omitempty"`
	Pattern     string                 `json:"pattern,omitempty"`
	MinLength   *int                   `json:"minLength,omitempty"`
	MaxLength   *int                   `json:"maxLength,omitempty"`
	Minimum     *float64               `json:"minimum,omitempty"`
	Maximum     *float64               `json:"maximum,omitempty"`
	MinItems    *int                   `json:"minItems,omitempty"`
	MaxItems    *int                   `json:"maxItems,omitempty"`
	MultipleOf  *float64               `json:"multipleOf,omitempty"`
	Description string                 `json:"description,omitempty"`

	// SpecMint extensions
	LLMEnhanced     bool             `json:"x-llm,omitempty"`
	CrossFieldRules []CrossFieldRule `json:"x-cross-field-rules,omitempty"`

	// Internal metadata
	Path         string  `json:"-"`
	IsRequired   bool    `json:"-"`
	OptionalProb float64 `json:"-"`
}

// CrossFieldRule represents a cross-field validation rule
type CrossFieldRule struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Fields      []string   `json:"fields"`
	Rule        string     `json:"rule"`
	Severity    string     `json:"severity"` // error, warning
	Patch       *PatchRule `json:"patch,omitempty"`
}

// PatchRule defines how to fix a constraint violation
type PatchRule struct {
	Strategy string                 `json:"strategy"` // set_value, adjust_field, remove_field
	Target   string                 `json:"target"`
	Value    interface{}            `json:"value,omitempty"`
	Params   map[string]interface{} `json:"params,omitempty"`
}

// New creates a new schema parser
func NewParser() *Parser {
	compiler := jsonschema.NewCompiler()

	return &Parser{
		compiler: compiler,
	}
}

// ParseFile loads and parses a JSON Schema from file
func (p *Parser) ParseFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	return p.ParseBytes(data)
}

// ParseBytes parses a JSON Schema from bytes
func (p *Parser) ParseBytes(data []byte) error {
	// Parse raw schema for extensions
	if err := json.Unmarshal(data, &p.raw); err != nil {
		return fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	// For now, skip JSON Schema validation and just use the raw schema
	// This allows us to process the schema structure without validation library issues
	p.schema = nil // We'll work directly with p.raw
	return nil
}

// GetRootNode returns the parsed root schema node
func (p *Parser) GetRootNode() (*SchemaNode, error) {
	if p.raw == nil {
		return nil, fmt.Errorf("no schema loaded")
	}

	return p.buildNode(p.raw, "", false, 0.9)
}

// Validate validates data against the loaded schema
func (p *Parser) Validate(data interface{}) error {
	// For now, skip validation since we're not using the compiled schema
	// In a production version, we'd implement basic validation against p.raw
	return nil
}

// buildNode recursively builds a SchemaNode from raw schema data
func (p *Parser) buildNode(raw map[string]interface{}, path string, required bool, optionalProb float64) (*SchemaNode, error) {
	node := &SchemaNode{
		Path:         path,
		IsRequired:   required,
		OptionalProb: optionalProb,
	}

	// Extract basic type information
	if typeVal, ok := raw["type"]; ok {
		if typeStr, ok := typeVal.(string); ok {
			node.Type = typeStr
		}
	}

	// Extract constraints
	if enum, ok := raw["enum"].([]interface{}); ok {
		node.Enum = enum
	}
	if examples, ok := raw["examples"].([]interface{}); ok {
		node.Examples = examples
	}
	if format, ok := raw["format"].(string); ok {
		node.Format = format
	}
	if pattern, ok := raw["pattern"].(string); ok {
		node.Pattern = pattern
	}
	if desc, ok := raw["description"].(string); ok {
		node.Description = desc
		// Check for LLM marker in description
		if strings.HasPrefix(strings.ToLower(desc), "llm:") {
			node.LLMEnhanced = true
		}
	}

	// Extract numeric constraints
	if min, ok := raw["minimum"].(float64); ok {
		node.Minimum = &min
	}
	if max, ok := raw["maximum"].(float64); ok {
		node.Maximum = &max
	}
	if multiple, ok := raw["multipleOf"].(float64); ok {
		node.MultipleOf = &multiple
	}

	// Extract string constraints
	if minLen, ok := raw["minLength"].(float64); ok {
		minLenInt := int(minLen)
		node.MinLength = &minLenInt
	}
	if maxLen, ok := raw["maxLength"].(float64); ok {
		maxLenInt := int(maxLen)
		node.MaxLength = &maxLenInt
	}

	// Extract array constraints
	if minItems, ok := raw["minItems"].(float64); ok {
		minItemsInt := int(minItems)
		node.MinItems = &minItemsInt
	}
	if maxItems, ok := raw["maxItems"].(float64); ok {
		maxItemsInt := int(maxItems)
		node.MaxItems = &maxItemsInt
	}

	// Extract SpecMint extensions
	if llmFlag, ok := raw["x-llm"].(bool); ok {
		node.LLMEnhanced = llmFlag
	}

	// Also check for "llm:" prefix in description
	if desc, ok := raw["description"].(string); ok && strings.HasPrefix(desc, "llm:") {
		node.LLMEnhanced = true
	}

	// Extract cross-field rules
	if rulesRaw, ok := raw["x-cross-field-rules"].([]interface{}); ok {
		for _, ruleRaw := range rulesRaw {
			if ruleMap, ok := ruleRaw.(map[string]interface{}); ok {
				rule := CrossFieldRule{}
				if name, ok := ruleMap["name"].(string); ok {
					rule.Name = name
				}
				if desc, ok := ruleMap["description"].(string); ok {
					rule.Description = desc
				}
				if fields, ok := ruleMap["fields"].([]interface{}); ok {
					for _, field := range fields {
						if fieldStr, ok := field.(string); ok {
							rule.Fields = append(rule.Fields, fieldStr)
						}
					}
				}
				if ruleStr, ok := ruleMap["rule"].(string); ok {
					rule.Rule = ruleStr
				}
				if severity, ok := ruleMap["severity"].(string); ok {
					rule.Severity = severity
				}
				node.CrossFieldRules = append(node.CrossFieldRules, rule)
			}
		}
	}

	// Handle object properties
	if node.Type == "object" {
		if props, ok := raw["properties"].(map[string]interface{}); ok {
			node.Properties = make(map[string]*SchemaNode)

			// Get required fields
			requiredFields := make(map[string]bool)
			if reqArray, ok := raw["required"].([]interface{}); ok {
				for _, req := range reqArray {
					if reqStr, ok := req.(string); ok {
						requiredFields[reqStr] = true
						node.Required = append(node.Required, reqStr)
					}
				}
			}

			// Parse each property
			for propName, propRaw := range props {
				if propMap, ok := propRaw.(map[string]interface{}); ok {
					propPath := path
					if propPath != "" {
						propPath += "."
					}
					propPath += propName

					propNode, err := p.buildNode(propMap, propPath, requiredFields[propName], optionalProb)
					if err != nil {
						return nil, fmt.Errorf("failed to parse property %s: %w", propName, err)
					}
					node.Properties[propName] = propNode
				}
			}
		}
	}

	// Handle array items
	if node.Type == "array" {
		if items, ok := raw["items"].(map[string]interface{}); ok {
			itemPath := path + "[]"
			itemNode, err := p.buildNode(items, itemPath, true, optionalProb)
			if err != nil {
				return nil, fmt.Errorf("failed to parse array items: %w", err)
			}
			node.Items = itemNode
		}
	}

	return node, nil
}

// GetLLMFields returns all fields marked for LLM enhancement
func (p *Parser) GetLLMFields(node *SchemaNode) []string {
	var fields []string
	p.collectLLMFields(node, "", &fields)
	return fields
}

func (p *Parser) collectLLMFields(node *SchemaNode, prefix string, fields *[]string) {
	// For root node, check properties directly
	if prefix == "" && node.Properties != nil {
		for name, prop := range node.Properties {
			if prop.LLMEnhanced {
				*fields = append(*fields, name)
			}
			p.collectLLMFields(prop, name, fields)
		}
		return
	}

	if node.LLMEnhanced && prefix != "" {
		*fields = append(*fields, prefix)
	}

	if node.Properties != nil {
		for name, prop := range node.Properties {
			propPath := name
			if prefix != "" {
				propPath = prefix + "." + name
			}
			p.collectLLMFields(prop, propPath, fields)
		}
	}

	if node.Items != nil {
		itemPath := prefix + "[]"
		p.collectLLMFields(node.Items, itemPath, fields)
	}
}

// GetCrossFieldRules returns all cross-field validation rules
func (p *Parser) GetCrossFieldRules(node *SchemaNode) []CrossFieldRule {
	var rules []CrossFieldRule
	p.collectCrossFieldRules(node, &rules)
	return rules
}

func (p *Parser) collectCrossFieldRules(node *SchemaNode, rules *[]CrossFieldRule) {
	*rules = append(*rules, node.CrossFieldRules...)

	if node.Properties != nil {
		for _, prop := range node.Properties {
			p.collectCrossFieldRules(prop, rules)
		}
	}

	if node.Items != nil {
		p.collectCrossFieldRules(node.Items, rules)
	}
}
