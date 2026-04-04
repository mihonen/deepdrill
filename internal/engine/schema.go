package engine


import (
	"fmt"
	"strings"
)


type Field struct {
    Name    string
    Hint    string
    Options []string  
    Type    FieldType
}


func (f *Field) String() string {
    s := fmt.Sprintf("- FIELD: %s (%s)\n  DESCRIPTION: %s", f.Name, f.Type, f.Hint)
    
    if len(f.Options) > 0 {
        s += fmt.Sprintf("\n  VALID_OPTIONS: [%s] (Choose ONLY from this list)", strings.Join(f.Options, ", "))
    }
    
    switch f.Type {
    case FieldTypeLink:
        s += "\n  REQUIREMENT: Must be a fully qualified URL."
    case FieldTypeImage:
        s += "\n  REQUIREMENT: Use the 'src' or 'data-src' attribute."
    }
    
    return s
}

type FieldType string
const (
    FieldTypeText  FieldType = "text"
    FieldTypeLink  FieldType = "link"
    FieldTypeImage FieldType = "image"
    FieldTypeValue FieldType = "value"  
    FieldTypeFlag  FieldType = "flag"  
)

type Schema struct {
    Fields   []Field
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
