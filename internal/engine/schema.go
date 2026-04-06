package engine

import (
	"fmt"
	"strings"
)

type Field struct {
	Name     string
	Hint     string
	Options  []string
	Type     FieldType
	Schema   *Schema // for FieldTypeCustom: defines the nested object's fields
	Multiple bool    // for FieldTypeCustom: return []object instead of object
	         	    // for FieldTypeList: always true, no need to set
}

func (f *Field) String() string {
	typeLabel := string(f.Type)
	if f.Type == FieldTypeCustom && f.Multiple {
		typeLabel = "custom[]"
	}

	s := fmt.Sprintf("- FIELD: %s (%s)\n  DESCRIPTION: %s", f.Name, typeLabel, f.Hint)

	if len(f.Options) > 0 {
		s += fmt.Sprintf("\n  VALID_OPTIONS: [%s] (Choose ONLY from this list)", strings.Join(f.Options, ", "))
	}

	switch f.Type {
	case FieldTypeLink:
		s += "\n  REQUIREMENT: Must be a fully qualified URL."
	case FieldTypeImage:
		s += "\n  REQUIREMENT: Use the 'src' or 'data-src' attribute."
	case FieldTypeList:
		s += "\n  REQUIREMENT: Return as a JSON array of strings."
		if len(f.Options) > 0 {
			s += " Select all that apply from VALID_OPTIONS."
		}
	case FieldTypeCustom:
		if f.Schema != nil {
			s += "\n  NESTED_SCHEMA:"
			for _, nested := range f.Schema.Fields {
				indented := strings.ReplaceAll(nested.String(), "\n", "\n    ")
				s += "\n    " + indented
			}
		}
		if f.Multiple {
			s += "\n  REQUIREMENT: Return as a JSON array of objects, one per item found."
		} else {
			s += "\n  REQUIREMENT: Return as a single JSON object."
		}
	}

	return s
}

type FieldType string

const (
	FieldTypeText   FieldType = "text"
	FieldTypeLink   FieldType = "link"
	FieldTypeImage  FieldType = "image"
	FieldTypeValue  FieldType = "value"
	FieldTypeFlag   FieldType = "flag"
	FieldTypeList   FieldType = "list"
	FieldTypeCustom FieldType = "custom"
)

type Schema struct {
	Fields []Field
}

func (s *Schema) String() string {
	var builder strings.Builder
	builder.WriteString("DATA EXTRACTION SCHEMA:\n")
	for i, field := range s.Fields {
		builder.WriteString(fmt.Sprintf("%d. %s", i+1, field.String()))
		builder.WriteString("\n")
	}
	return builder.String()
}

type Options struct {
	Multiple bool
	Depth    uint64
	URL      string
}
