package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/mihonen/deepdrill"
)

func main() {
	godotenv.Load()

	provider := os.Getenv("DEEPDRILL_PROVIDER")
	model := os.Getenv("DEEPDRILL_MODEL")
	apiKey := os.Getenv("DEEPDRILL_API_KEY")

	if provider == "" {
		provider = "deepseek"
	}
	if model == "" {
		model = "deepseek-chat"
	}
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "DEEPDRILL_API_KEY is required (set in .env or environment)")
		os.Exit(1)
	}

	deepdrill.Init(provider, model, apiKey)

	schema := deepdrill.Schema{
		Fields: []deepdrill.Field{
			{Name: "title",     Type: deepdrill.FieldTypeText,  Hint: "the article headline"},
			{Name: "body",      Type: deepdrill.FieldTypeText,  Hint: "full article body, preserve paragraphs"},
			{Name: "link",      Type: deepdrill.FieldTypeLink,  Hint: "link to the full article"},
			{Name: "thumbnail", Type: deepdrill.FieldTypeImage, Hint: "main article image, not a logo"},
			{Name: "author",    Type: deepdrill.FieldTypeText,  Hint: "author name"},
			{Name: "date",      Type: deepdrill.FieldTypeValue, Hint: "publication date in ISO format"},
			{Name: "topic",     Type: deepdrill.FieldTypeText,  Hint: "article topic",
				Options: []string{
					"politics", "business", "technology", "science", "health",
					"environment", "sports", "entertainment", "arts", "world",
					"economy", "education", "law", "opinion",
				},
			},
		},
	}

	options := deepdrill.Options{
		Multiple: true,
		Depth:    1,
		URL:      "https://www.nytimes.com",
	}

	res, err := deepdrill.Fill(context.Background(), schema, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	filename := fmt.Sprintf("nytimes_%s.json", time.Now().Format("2006-01-02_15-04-05"))

	b, _ := json.MarshalIndent(res, "", "  ")
	if err := os.WriteFile(filename, b, 0644); err != nil {
		fmt.Println("Error saving file:", err)
		return
	}

	fmt.Printf("fetched %d articles - saved to file %s\n", len(res), filename)
}
