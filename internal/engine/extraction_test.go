package engine




import (
	"testing"
	"os"
	"fmt"
	"reflect"
    "github.com/PuerkitoBio/goquery"
    "encoding/json"
    "context"
    "strings"
)



func TestStringMatching(t *testing.T) {

	a := "For Democrats, the Era of the Girl Dad and Male Ally Is Over"
	b := []string{"For Democrats, the Era of the Girl Dad and Male Ally Is Over"}

	score := stringSimilarity(normalizeField(a), normalizeField(b))


	if score != 1.0 {
	    t.Errorf("Wrong score for same strings: A\n%s\n B: \n%s\nScore: %.2f\n", a, b, score)
	}

}

func TestNYTimesExtraction(t *testing.T) {

	apiKey := os.Getenv("DEEPDRILL_API_KEY")
	if apiKey == "" {
	    t.Skip("Skipping test: DEEPDRILL_API_KEY environment variable not set")
	}
	
	model := os.Getenv("DEEPDRILL_MODEL")
	if model == "" {
	    model = "deepseek-chat" // default
	}
	
	provider := os.Getenv("DEEPDRILL_PROVIDER")
	if provider == "" {
	    model = "deepseek" // default
	}
	

	f, err := os.Open("pages/nytimes.html")
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		t.Fatalf("failed to parse html: %v", err)
	}

	gtBytes, err := os.ReadFile("pages/nytimes.json")
	if err != nil {
		t.Fatalf("failed to read ground truth: %v", err)
	}
	var groundTruth []map[string]any
	if err := json.Unmarshal(gtBytes, &groundTruth); err != nil {
		t.Fatalf("failed to parse ground truth: %v", err)
	}

	schema := Schema{
	    Fields: []Field{
	        {Name: "title",     Type: FieldTypeText,  Hint: "the article headline"},
	        {Name: "body",      Type: FieldTypeText,  Hint: "full article body, preserve paragraphs, do not include title or read time in the body"},
	        {Name: "link",      Type: FieldTypeLink,  Hint: "link to the full article"},
	        {Name: "thumbnail", Type: FieldTypeImage, Hint: "main article image, not a logo"},
	        {Name: "author",    Type: FieldTypeText,  Hint: "author name"},
	        {Name: "date",      Type: FieldTypeValue, Hint: "publication date in ISO format"},
	    },
	}

	e := New(Config{
	    Provider: provider,
	    APIKey:   apiKey,
	    Model:    model, 
	})


	results, err := e.ExecuteFromDoc(context.Background(), schema, doc)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}



    // b, _ := json.MarshalIndent(results, "", "  ")
    // fmt.Println(string(b))

	score := evaluate(results, groundTruth)
	t.Logf("results: %d, ground truth: %d", len(results), len(groundTruth))
	t.Logf("score: %.2f%%", score*100)


}



func evaluate(results, groundTruth []map[string]any) float64 {

	totalScore := 0.0
	for _, gt := range groundTruth {
		best := 0.0
		for _, r := range results {
			s := scoreItem(r, gt)
			if s > best {
				best = s
			}
		}

		totalScore += best
	}
	return totalScore / float64(len(groundTruth))
}

func normalizeField(v any) string {
    if v == nil {
        return ""
    }
    
    if s, ok := v.(string); ok {
        return strings.TrimSpace(s)
    }
    
    // Handle any slice
    rv := reflect.ValueOf(v)
    if rv.Kind() == reflect.Slice {
        parts := []string{}
        for i := 0; i < rv.Len(); i++ {
            elem := rv.Index(i).Interface()
            if s, ok := elem.(string); ok && strings.TrimSpace(s) != "" {
                parts = append(parts, strings.TrimSpace(s))
            }
        }
        return strings.Join(parts, "\n\n")
    }
    
    return strings.TrimSpace(fmt.Sprint(v))
}

func scoreItem(result, gt map[string]any) float64 {
    fields := []string{"title", "body", "link", "thumbnail", "author", "date"}
    total := 0.0
    matched := 0.0
    for _, f := range fields {
        gtVal := normalizeField(gt[f])
        rVal := normalizeField(result[f])
        total++
        matched += stringSimilarity(gtVal, rVal)

    }
    return matched / total
}


func stringSimilarity(a, b string) float64 {
    a = strings.TrimSpace(strings.ToLower(a))
    b = strings.TrimSpace(strings.ToLower(b))
    if a == "" && b == "" {
        return 1.0
    }
    if a == "" || b == "" {
        return 0.0
    }
    
    aWords := strings.Fields(a)
    bSet := map[string]bool{}
    for _, w := range strings.Fields(b) {
        bSet[w] = true
    }
    
    overlap := 0
    for _, w := range aWords {
        if bSet[w] {
            overlap++
        }
    }
    
    return float64(overlap) / float64(max(len(aWords), len(strings.Fields(b))))
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

