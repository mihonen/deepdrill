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

	hintDescription := `
		Full description of the listing. Format it in valid Markdown syntax:
		    - Use blank lines between paragraphs.
		    - Use - for bullet points.
		    - Preserve line breaks exactly.
		    - Do not summarize or merge lines.
	`

	categories := []string{
		"UI / UX Design", "Web Design", "Interaction Design", "Product Design (Digital)",
		"Motion Design", "Generative Art", "Code-based / Algorithmic Art", "Game Art / Game Design",
		"Virtual / Augmented Reality", "3D Design", "3D Animation", "Installation Art",
		"Architecture", "Interior Design", "Landscape Architecture", "Urban Design",
		"Illustration", "Painting", "Drawing", "Printmaking", "Photography",
		"Collage / Mixed Media", "Sculpture", "Ceramics", "Textile / Fiber Art",
		"Jewelry / Metalwork", "Product Design (Physical)", "AI Art", "Sound Art",
		"Video Art", "Performance", "Bio Art",
	}

	benefits := []string{
		"cash", "trophy", "exposure", "networking", "certificate", "production",
	}

	hostSchema := deepdrill.Schema{
		Fields: []deepdrill.Field{
			{Name: "name", Type: deepdrill.FieldTypeText,  Hint: "The name of the host organization or individual"},
			{Name: "link", Type: deepdrill.FieldTypeLink,  Hint: "An absolute URL to the host's website or profile"},
			{Name: "logo", Type: deepdrill.FieldTypeImage, Hint: "An absolute URL to the host's logo"},
			{Name: "ig",   Type: deepdrill.FieldTypeText,  Hint: "The host's Instagram handle (if available)"},
		},
	}

	prizeSchema := deepdrill.Schema{
		Fields: []deepdrill.Field{
			{Name: "name",  Type: deepdrill.FieldTypeText, Hint: "Name of the prize e.g. '1st Place' or 'Grand Prize'"},
			{Name: "value", Type: deepdrill.FieldTypeText, Hint: "Prize amount in the disclosed currency e.g. 1000€"},
		},
	}

	applicationInfoSchema := deepdrill.Schema{
		Fields: []deepdrill.Field{
			{Name: "howto",            Type: deepdrill.FieldTypeText, Hint: "Quick instructions on how to apply"},
			{Name: "format",           Type: deepdrill.FieldTypeText, Hint: "The format of the required deliverables"},
			{Name: "eligibility",      Type: deepdrill.FieldTypeText, Hint: "Eligibility criteria for applicants"},
			{Name: "deadline",         Type: deepdrill.FieldTypeText, Hint: "The application deadline in ISO format (e.g. 2025-12-31T23:59:59)"},
			{Name: "application_link", Type: deepdrill.FieldTypeLink, Hint: "An absolute URL to the application page or form"},
		},
	}

	listingSchema := deepdrill.Schema{
		Fields: []deepdrill.Field{
			{Name: "name",             Type: deepdrill.FieldTypeText,   Hint: "The name of the event"},
			{Name: "description",      Type: deepdrill.FieldTypeText,   Hint: "A brief description of the event"},
			{Name: "body",             Type: deepdrill.FieldTypeText,   Hint: hintDescription},
			{Name: "category",         Type: deepdrill.FieldTypeList,   Hint: "A list of categories or tags associated with the listing", Options: categories},
			{Name: "thumbnail",        Type: deepdrill.FieldTypeImage,  Hint: "An absolute URL to the best looking image on the page, should not be a logo"},
			{Name: "benefits",         Type: deepdrill.FieldTypeList,   Hint: "Benefits provided by the listing", Options: benefits},
			{Name: "additional_prizes",Type: deepdrill.FieldTypeText,   Hint: "Max two additional prizes or rewards (if any), each max 2 words"},
			{Name: "host",             Type: deepdrill.FieldTypeCustom, Hint: "Information about the host of the event", Schema: &hostSchema},
			{Name: "prizes",           Type: deepdrill.FieldTypeCustom, Hint: "A list of prize objects", Schema: &prizeSchema, Multiple: true},
			{Name: "application_info", Type: deepdrill.FieldTypeCustom, Hint: "Information about applying and eligibility", Schema: &applicationInfoSchema},
		},
	}

	options := deepdrill.Options{
		Multiple: true,
		Depth:    1,
		URL:      "https://www.contestwatchers.com/category/open/",
	}

	res, err := deepdrill.Fill(context.Background(), listingSchema, options)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	filename := fmt.Sprintf("artevents_%s.json", time.Now().Format("2006-01-02_15-04-05"))

	b, _ := json.MarshalIndent(res, "", "  ")
	if err := os.WriteFile(filename, b, 0644); err != nil {
		fmt.Println("Error saving file:", err)
		return
	}

	fmt.Printf("fetched %d events - saved to file %s\n", len(res), filename)
}
