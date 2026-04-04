# deepdrill

LLM-powered web scraper for extracting structured data from any webpage — with minimal token usage.

```
go get github.com/mihonen/deepdrill
```

---

## How it works

Most LLM scrapers dump raw HTML into a prompt and let the model deal with the noise. deepdrill prunes first:

1. **Build a semantic tree** — strip scripts, styles, ads, and collapse meaningless div nesting down to actual content nodes
2. **Render clean HTML** — the pruned tree is serialized back to minimal HTML with node IDs attached
3. **LLM extracts directly** — the model receives the clean HTML and returns structured JSON with the extracted values

The HTML sent to the LLM is a fraction of the original page size, so token usage stays low even on content-heavy pages.

---

## Quick start

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/joho/godotenv"
    "github.com/mihonen/deepdrill"
)

func main() {
    godotenv.Load()

    deepdrill.Init(
        os.Getenv("DEEPDRILL_PROVIDER"), // "deepseek" or "openai"
        os.Getenv("DEEPDRILL_MODEL"),    // e.g. "deepseek-chat"
        os.Getenv("DEEPDRILL_API_KEY"),
    )

    schema := deepdrill.Schema{
        Fields: []deepdrill.Field{
            {Name: "title",  Type: deepdrill.FieldTypeText,  Hint: "article headline"},
            {Name: "link",   Type: deepdrill.FieldTypeLink,  Hint: "link to the full article"},
            {Name: "author", Type: deepdrill.FieldTypeText,  Hint: "author name"},
            {Name: "date",   Type: deepdrill.FieldTypeValue, Hint: "publication date in ISO format"},
            {Name: "topic",  Type: deepdrill.FieldTypeText,  Hint: "article topic",
                Options: []string{"politics", "business", "sport"}},
        },
    }

    results, err := deepdrill.Fill(context.Background(), schema, deepdrill.Options{
        URL:      "https://www.nytimes.com",
        Multiple: true,
    })
    if err != nil {
        fmt.Println(err)
        return
    }

    b, _ := json.MarshalIndent(results, "", "  ")
    fmt.Println(string(b))
}
```

**Output:**

```json
[
  {
    "author": "Rebecca Robbins",
    "date": "2025-04-03",
    "link": "https://www.nytimes.com/2025/04/03/health/...",
    "title": "Drug Pricing Order Targets Middlemen",
    "topic": "politics"
  },
  ...
]
```

---

## Setup

Copy `.env.example` to `.env` and fill in your credentials:

```
DEEPDRILL_PROVIDER=deepseek
DEEPDRILL_MODEL=deepseek-chat
DEEPDRILL_API_KEY=your-api-key-here
```

Supported providers out of the box: **`deepseek`**, **`openai`**. Any OpenAI-compatible API works via a custom `BaseURL`.

---

## Schema

A `Schema` is a list of `Field`s describing what you want to extract.

```go
type Field struct {
    Name    string      // key in the output map
    Type    FieldType   // see field types below
    Hint    string      // natural language hint for the LLM
    Options []string    // constrain the value to a fixed set (optional)
}
```

### Field types

| Type | Description |
|---|---|
| `FieldTypeText` | Text content of a node |
| `FieldTypeLink` | `href` attribute of an `<a>` tag |
| `FieldTypeImage` | `src` attribute of an `<img>` tag |
| `FieldTypeValue` | Scalar value — dates, numbers, etc. |
| `FieldTypeFlag` | Boolean presence of something on the page |

### Options

Constrain a field to a fixed set of values — the LLM will only pick from the list:

```go
{Name: "category", Type: deepdrill.FieldTypeText, Options: []string{"news", "opinion", "sport"}}
```

---

## Options

```go
type Options struct {
    URL      string // page to scrape
    Multiple bool   // expect multiple items (e.g. article listings)
    Depth    uint64 // TODO: follow links recursively up to this depth
}
```

---

## Semantic tree

deepdrill compresses the DOM into a flat, readable representation before sending it to the LLM. This reduces token usage and makes it easier for the model to reason about structure.

**Input HTML:**

```html
<div>
  <div>
    <h1>Top Stories</h1>
    <div>
      <div>
        <p>Trump this and that</p>
        <p>Another war</p>
      </div>
      <div>
        <p>China</p>
        <p>Something crazy going on in China</p>
      </div>
    </div>
  </div>
</div>
```

**Semantic tree:**

```
[group]
  [h1] Top Stories
  [group]
    [p] Trump this and that
    [p] Another war
  [group]
    [p] China
    [p] Something crazy going on in China
```

Deeply nested single-child divs collapse to their content. Groups with mixed content are preserved. Links and images carry their `href`/`src` attributes.

---

## Roadmap

- [ ] **Heuristic caching** — after extraction, match result values back to their node IDs to build reusable path heuristics. Cache them keyed by schema + page skeleton (SQLite). Same site structure on a future run? Skip the LLM entirely.
- [ ] **Recursive depth** — follow extracted links and scrape sub-pages automatically

---

## License

MIT
