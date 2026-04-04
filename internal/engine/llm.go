package engine



import (
	"context"
    "github.com/sashabaranov/go-openai"
    "encoding/json"
    "strings"
    "fmt"
)

func (e *Engine) getHeuristicsFromLLM(ctx context.Context, tree string, schema string) ([]map[string]FieldHeuristic, error) {
    rules := `
        - "join": true — content is spread across multiple nodes but belongs to one string field
        - "multiple": true — field is a []string, each item is one element
        - Return ONLY a JSON ARRAY of objects, even if only one item is found
        - Identify EVERY distinct item matching the schema in the provided tree
        - No explanation, no markdown
        - If a field is not present set to null
        - If no objects found, return empty array 
        - Only return content exactly as it appears on the page, do not format or edit content`

    responseFormat := `[
        {
          "title":  { "join": false, "multiple": false, "content": ["NASA Unveils First Earth Photos From Artemis II"] },
          "body":   { "join": false, "multiple": false, "content": ["The pictures were released on the third day of the first mission to send people around the moon since 1972."] }
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

    raw := e.cleanLLMResponse(resp.Choices[0].Message.Content)
    
    var results []map[string]FieldHeuristic
    if err := json.Unmarshal([]byte(raw), &results); err != nil {
        // Fallback: If LLM ignored the "Array" rule and sent a single object, wrap it
        var single map[string]FieldHeuristic
        if err2 := json.Unmarshal([]byte(raw), &single); err2 == nil {
            return []map[string]FieldHeuristic{single}, nil
        }
        return nil, fmt.Errorf("failed to parse: %w", err)
    }
    return results, nil
}


func (e *Engine) cleanLLMResponse(raw string) string {
    cleaned := strings.TrimSpace(raw)

    if strings.HasPrefix(cleaned, "```") {
        firstNewline := strings.Index(cleaned, "\n")
        if firstNewline != -1 {
            cleaned = cleaned[firstNewline:]
        }
        // Remove the closing tag
        cleaned = strings.TrimSuffix(cleaned, "```")
    }

    return strings.TrimSpace(cleaned)
}
