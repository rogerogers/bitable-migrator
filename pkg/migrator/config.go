package migrator

import (
	"fmt"
	"strings"

	"go.yaml.in/yaml/v4"
)

// BitableConfig represents the root config structure of bitable.yaml
type BitableConfig struct {
	AppToken string        `yaml:"app_token"`
	Tables   []TableConfig `yaml:"tables"`
}

// TableConfig represents a table configuration
type TableConfig struct {
	TableID string        `yaml:"table_id"`
	Name    string        `yaml:"name"`
	Fields  []FieldConfig `yaml:"fields"`
}

// FieldConfig represents a field definition
type FieldConfig struct {
	FieldID     string                 `yaml:"field_id,omitempty"`
	Name        string                 `yaml:"name"`
	Type        FieldType              `yaml:"type"`
	Description string                 `yaml:"description,omitempty"`
	Property    map[string]interface{} `yaml:"property,omitempty"`
}

// FieldType is a custom type to handle both string and integer representation of Bitable Field Types
type FieldType int

var typeToString = map[int]string{
	1:    "Text",
	2:    "Number",
	3:    "SingleSelect",
	4:    "MultiSelect",
	5:    "Date",
	7:    "Checkbox",
	11:   "User",
	13:   "Phone",
	15:   "Url",
	17:   "Attachment",
	18:   "SingleLink",
	20:   "Formula",
	21:   "DuplexLink",
	22:   "Location",
	23:   "GroupChat",
	1001: "CreatedTime",
	1002: "ModifiedTime",
	1003: "CreatedUser",
	1004: "ModifiedUser",
	1005: "AutoNumber",
}

var stringToType = map[string]int{
	// Standard / Developer-Friendly names
	"text":         1,
	"multiline":    1,
	"number":       2,
	"singleselect": 3,
	"single option": 3,
	"singleoption": 3,
	"multiselect":  4,
	"multiple options": 4,
	"multipleoptions": 4,
	"date":         5,
	"checkbox":     7,
	"user":         11,
	"person":       11,
	"phone":        13,
	"phonenumber":  13,
	"url":          15,
	"link":         15,
	"attachment":   17,
	"singlelink":   18,
	"one-way link": 18,
	"onewaylink":   18,
	"formula":      20,
	"duplexlink":   21,
	"two-way link": 21,
	"twowaylink":   21,
	"location":     22,
	"groupchat":    23,
	"group":        23,
	"createdtime":  1001,
	"date created": 1001,
	"createdtimefield": 1001,
	"modifiedtime": 1002,
	"last modified date": 1002,
	"modifiedtimefield": 1002,
	"createduser":  1003,
	"created by":   1003,
	"createduserfield": 1003,
	"modifieduser": 1004,
	"modified by":  1004,
	"modifieduserfield": 1004,
	"autonumber":   1005,
	"autoserial":   1005,
}

// UnmarshalYAML implements Custom YAML parsing for FieldType
func (ft *FieldType) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err == nil {
		normalized := strings.ToLower(strings.TrimSpace(s))
		val, ok := stringToType[normalized]
		if ok {
			*ft = FieldType(val)
			return nil
		}
		return fmt.Errorf("unknown bitable field type name: %q", s)
	}

	var i int
	if err := value.Decode(&i); err == nil {
		*ft = FieldType(i)
		return nil
	}

	return fmt.Errorf("invalid field type value at line %d", value.Line)
}

// MarshalYAML implements Custom YAML generation for FieldType
func (ft FieldType) MarshalYAML() (interface{}, error) {
	if str, ok := typeToString[int(ft)]; ok {
		return str, nil
	}
	return int(ft), nil
}

// String returns a human-readable representation of the field type
func (ft FieldType) String() string {
	if str, ok := typeToString[int(ft)]; ok {
		return str
	}
	return fmt.Sprintf("Unknown(%d)", int(ft))
}
