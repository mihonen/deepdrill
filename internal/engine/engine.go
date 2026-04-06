package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/sashabaranov/go-openai"
)

type Engine struct {
	client    *http.Client
	llmClient *openai.Client
	model     string
	cache     *Cache
}

type Config struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}

func New(cfg Config) *Engine {
	openaiCfg := openai.DefaultConfig(cfg.APIKey)

	switch cfg.Provider {
	case "deepseek":
		if cfg.BaseURL == "" {
			openaiCfg.BaseURL = "https://api.deepseek.com/v1"
		}
		if cfg.Model == "" {
			cfg.Model = "deepseek-chat"
		}
	case "openai":
		if cfg.BaseURL == "" {
			openaiCfg.BaseURL = "https://api.openai.com/v1"
		}
		if cfg.Model == "" {
			cfg.Model = "gpt-4o-mini"
		}
	default:
		panic("unknown provider: " + cfg.Provider)
	}

	if cfg.BaseURL != "" {
		openaiCfg.BaseURL = cfg.BaseURL
	}

	return &Engine{
		client:    &http.Client{},
		llmClient: openai.NewClientWithConfig(openaiCfg),
		model:     cfg.Model,
		cache:     NewCache(),
	}
}

func (e *Engine) Execute(ctx context.Context, schema Schema, options Options) ([]map[string]any, error) {
	doc, err := e.fetch(options.URL)
	if err != nil {
		return nil, err
	}
	return e.ExecuteFromDoc(ctx, schema, doc)
}

func (e *Engine) ExecuteFromDoc(ctx context.Context, schema Schema, doc *goquery.Document) ([]map[string]any, error) {
	cleanDoc := goquery.NewDocumentFromNode(doc.Clone().Get(0))
	e.clean(cleanDoc)

	subTrees := CreateSemanticTree(cleanDoc).Split(100)

	var wg sync.WaitGroup
	resultsChan := make(chan []map[string]any, len(subTrees))
	errChan     := make(chan error, len(subTrees))

	for _, sub := range subTrees {
		wg.Add(1)
		go func(st *SemanticTree) {
			defer wg.Done()

			if strings.TrimSpace(st.HTMLString()) == "" {
				return
			}

			extracted, err := e.extract(ctx, st.HTMLString(), schema.String())
			if err != nil {
				errChan <- err
				return
			}

			var chunkResults []map[string]any
			for _, row := range extracted {
				obj := make(map[string]any)
				for field, f := range row {
					if len(f.Content) == 0 {
						continue
					}
					var val any
					if err := json.Unmarshal(f.Content, &val); err == nil {
						obj[field] = val
					}
				}
				chunkResults = append(chunkResults, obj)
			}

			resultsChan <- chunkResults
		}(sub)
	}

	wg.Wait()
	close(resultsChan)
	close(errChan)

	var results []map[string]any
	for batch := range resultsChan {
		if hasAnyValue(batch) {
			results = append(results, batch...)
		}
	}
	return results, nil
}

func hasAnyValue(v any) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) != ""
	case []any:
		for _, item := range val {
			if hasAnyValue(item) {
				return true
			}
		}
		return false
	case map[string]any:
		for _, inner := range val {
			if hasAnyValue(inner) {
				return true
			}
		}
		return false
	default:
		return true
	}
}
