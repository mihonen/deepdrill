package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// llmField is the per-field envelope the LLM returns.
// Content is raw JSON — a string, array, or object depending on the field type.
type llmField struct {
	Content json.RawMessage `json:"content"`
}

func (e *Engine) extract(ctx context.Context, tree string, schema string) ([]map[string]llmField, error) {
	rules := `
        - Return ONLY a JSON ARRAY of objects, even if only one item is found
        - Identify EVERY distinct item matching the schema in the provided tree
        - No explanation, no markdown, no code fences
        - If a field is not present set its content to null
        - If no real content is found matching the schema, return []
        - DO NOT return irrelevant content
        - Only return content exactly as it appears on the page, do not format or edit content
        - content type per field type:
            text / value / flag  →  a single string:  "some text"
            link / image         →  a single string URL: "https://..."
            list                 →  array of strings: ["a", "b", "c"]
            custom               →  a single object:  {"field": "value", ...}
            custom[]             →  array of objects: [{"field": "value"}, ...]`

	responseFormat := `[
        {
          "title":   { "content": "NASA Unveils First Earth Photos From Artemis II" },
          "date":    { "content": "2025-04-03" },
          "tags":    { "content": ["space", "science"] },
          "host":    { "content": {"name": "NASA", "link": "https://nasa.gov"} },
          "missing": { "content": null }
        },
        { ... second item ... }
    ]`

	prompt := fmt.Sprintf(`
        You are a web scraping assistant. Extract ALL items matching the given schema.

        Tree: %s
        Schema: %s
        Rules: %s
        Response format: %s`, tree, schema, rules, responseFormat)

	resp, err := e.llmClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: e.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return nil, err
	}

	raw := cleanResponse(resp.Choices[0].Message.Content)

	var results []map[string]llmField
	if err := json.Unmarshal([]byte(raw), &results); err != nil {
		// Fallback: LLM returned a single object instead of an array
		var single map[string]llmField
		if err2 := json.Unmarshal([]byte(raw), &single); err2 == nil {
			return []map[string]llmField{single}, nil
		}
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}
	return results, nil
}

func cleanResponse(raw string) string {
	s := strings.TrimSpace(raw)
	if strings.HasPrefix(s, "```") {
		if i := strings.Index(s, "\n"); i != -1 {
			s = s[i:]
		}
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}
