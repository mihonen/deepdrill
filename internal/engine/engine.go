package engine

import (
    "fmt"
    "sync"
    "context"
    "strings"
    "net/http"
    "github.com/PuerkitoBio/goquery"
    "github.com/sashabaranov/go-openai"
)

type Engine struct {
    client *http.Client
    llmClient *openai.Client
    model     string 
    cache *Cache
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
            cfg.Model = "gpt-5.4-nano" 
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
    url := options.URL
    doc, err := e.fetch(url)
    if err != nil {
        return nil, err
    }

    return e.ExecuteFromDoc(ctx, schema, doc)

}

func (e *Engine) ExecuteFromDoc(ctx context.Context, schema Schema, doc *goquery.Document) ([]map[string]any, error){

    cleanDoc := goquery.NewDocumentFromNode(doc.Clone().Get(0))
    e.clean(cleanDoc)
    
    masterTree := CreateSemanticTree(cleanDoc)

    subTrees := masterTree.Split(50)

    var wg sync.WaitGroup
    resultsChan := make(chan []map[string]any, len(subTrees))
    errChan := make(chan error, len(subTrees))

    for _, sub := range subTrees {

        wg.Add(1)
        go func(st *SemanticTree) {
            defer wg.Done()

            treeStr := st.HTMLString()
            if strings.TrimSpace(treeStr) == "" {
                return
            }

            heuristics, err := e.getHeuristicsFromLLM(ctx, st.String(), schema.String())
            if err != nil {
                errChan <- err
                fmt.Println(err)
                return
            }



            var chunkResults []map[string]any
            for _, hMap := range heuristics {
                obj := make(map[string]any)
                for field, h := range hMap{
                    obj[field] = h.Content
                }
                chunkResults = append(chunkResults, obj)
            }

            resultsChan <- chunkResults
        }(sub)
    }

    wg.Wait()
    close(resultsChan)
    close(errChan)

    var finalResults []map[string]any
    for batch := range resultsChan {
        finalResults = append(finalResults, batch...)
    }

    return finalResults, nil
}


func (e *Engine) fetch(url string) (*goquery.Document, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")

    res, err := e.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    if res.StatusCode != 200 {
        return nil, fmt.Errorf("status code error: %d", res.StatusCode)
    }

    return goquery.NewDocumentFromReader(res.Body)
}



func (e *Engine) clean(doc *goquery.Document)  {
    doc.Find("script, style, noscript, iframe, head, svg, link").Remove()
    doc.Find("*").Each(func(i int, s *goquery.Selection) {
        if !s.Is("a") {
            s.RemoveAttr("class")
            s.RemoveAttr("id")
            s.RemoveAttr("style")
        } else {
            href, _ := s.Attr("href")
            s.SetAttr("href", href)
            s.RemoveAttr("class")
            s.RemoveAttr("id")
        }
    })
}



func (e *Engine) applyHeuristics(tree *SemanticTree, schema Schema, heuristics map[string]FieldHeuristic) (map[string]any, error) {
    results := make(map[string]any)

    for _, field := range schema.Fields {
        h, ok := heuristics[field.Name]
        if !ok {
            continue // LLM didn't find this field, skip or set null
        }

        getValue := func(path string) string {
            node := walkPath(tree.Root, path)
            if node == nil {
                return ""
            }
            
            switch field.Type {
            case FieldTypeLink:
                return node.Attrs["href"]
            case FieldTypeImage:
                return node.Attrs["src"]
            case FieldTypeFlag:
                return "true"
            default:
                return node.Content 
            }
        }

        switch {
        case h.Join:
            var parts []string
            for _, heuristic := range h.Heuristics {
                if v := getValue(heuristic.Path); v != "" {
                    parts = append(parts, v)
                }
            }
            results[field.Name] = strings.Join(parts, "\n\n")

        case h.Multiple:
            var slice []string
            for _, heuristic := range h.Heuristics {
                if v := getValue(heuristic.Path); v != "" {
                    slice = append(slice, v)
                }
            }
            results[field.Name] = slice

        default:
            if len(h.Heuristics) > 0 {
                results[field.Name] = getValue(h.Heuristics[0].Path)
            }
        }
    }

    return results, nil
}


func walkPath(node *SemanticNode, path string) *SemanticNode {
    current := node
    for _, ch := range path {
        idx := int(ch - '0')
        if idx >= len(current.Children) {
            return nil
        }
        current = current.Children[idx]
    }
    return current
}





