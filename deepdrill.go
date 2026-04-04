package deepdrill

import (
    "context"
    "fmt"
    "github.com/mihonen/deepdrill/internal/engine"
)

// Alias types so the user stays in the 'deepdrill' namespace
type Field      = engine.Field
type FieldType  = engine.FieldType
type Schema     = engine.Schema
type Options    = engine.Options

const (
    FieldTypeText  = engine.FieldTypeText
    FieldTypeLink  = engine.FieldTypeLink
    FieldTypeImage = engine.FieldTypeImage
    FieldTypeValue = engine.FieldTypeValue
)

var defaultEngine *engine.Engine

        


func Init(provider string, model string, apiKey string) {
    defaultEngine = engine.New(engine.Config{
        Provider: provider,
        APIKey:   apiKey,
        Model:    model, 
    })
}

func Fill(ctx context.Context, schema Schema, opts Options) ([]map[string]any, error) {
    if defaultEngine == nil {
        return nil, fmt.Errorf("deepdrill not initialized: call Init() first")
    }
    return defaultEngine.Execute(ctx, schema, opts)
}