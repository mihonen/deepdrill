package engine




type Heuristic struct {
    Path string   `json:"path"` 
}

type FieldHeuristic struct {
    Join       bool        `json:"join"`       // true = concat all into one string
    Multiple   bool        `json:"multiple"`   // true = map each to []string element
    Heuristics []Heuristic `json:"heuristics"`
    Content    []string    `json:"content"`
}
