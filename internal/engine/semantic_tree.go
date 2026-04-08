package engine

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	dropTags = map[string]bool{
		"script": true, "style": true, "noscript": true,
		"svg": true, "template": true,
		"head": true, "meta": true, "link": true,
		"input": true,
	}
	semanticTags = map[string]bool{
	    "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	    "p": true, "pre": true, "blockquote": true,
	    "span": true, "b": true, "i": true, "em": true, "strong": true,
	    "time": true, "data": true,
	    "td": true, "th": true, "caption": true,
	    "figcaption": true,
	    "label": true, "summary": true,
	    "dt": true, "dd": true,
	    "code": true, "kbd": true, "samp": true,
	    "a": true, "img": true, "video": true, "button": true, "source": true,
	}
)

type NodeType string

type SemanticNode struct {
	Type     NodeType
	Content  string
	Attrs    map[string]string
	Children []*SemanticNode
	OriginalID string
}

type SemanticTree struct {
	Root *SemanticNode
}


func (t *SemanticTree) String() string {
	var sb strings.Builder
	for idx, child := range t.Root.Children {
		renderNode(&sb, child, 0, strconv.Itoa(idx))
	}
	return strings.TrimSpace(sb.String())
}

func (t *SemanticTree) HTMLString() string {
	var sb strings.Builder
	for idx, child := range t.Root.Children {
		renderHTMLNode(&sb, child, 0, strconv.Itoa(idx))
	}
	return strings.TrimSpace(sb.String())
}

func renderHTMLNode(sb *strings.Builder, node *SemanticNode, depth int, index string) {
	indent := strings.Repeat("  ", depth)

	tag := htmlTag(node)
	attrs := htmlAttrs(node, index)

	if len(node.Children) == 0 {
		if node.Content != "" {
			sb.WriteString(fmt.Sprintf("%s<%s%s>%s</%s>\n", indent, tag, attrs, node.Content, tag))
		} else {
			sb.WriteString(fmt.Sprintf("%s<%s%s />\n", indent, tag, attrs))
		}
		return
	}

	sb.WriteString(fmt.Sprintf("%s<%s%s>%s\n", indent, tag, attrs, node.Content))
	for idx, child := range node.Children {
		renderHTMLNode(sb, child, depth+1, index+strconv.Itoa(idx))
	}
	sb.WriteString(fmt.Sprintf("%s</%s>\n", indent, tag))
}

func htmlTag(node *SemanticNode) string {
	switch node.Type {
	case "group":
		return "div"
	default:
		return string(node.Type)
	}
}

func htmlAttrs(node *SemanticNode, index string) string {
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf(` id="%s"`, index))
    keys := make([]string, 0, len(node.Attrs))
    for k := range node.Attrs {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    for _, k := range keys {
        sb.WriteString(fmt.Sprintf(` %s="%s"`, k, node.Attrs[k]))
    }
    return sb.String()
}

func (t *SemanticTree) Skeleton() string {
	var sb strings.Builder
	for _, child := range t.Root.Children {
		renderSkeleton(&sb, child, 0)
	}
	return strings.TrimSpace(sb.String())
}

func (n *SemanticNode) Count() int {
	count := 1
	for _, child := range n.Children {
		count += child.Count()
	}
	return count
}

func (t *SemanticTree) Split(maxNodes int) []*SemanticTree {
	var chunks []*SemanticTree
	
	var traverse func(node *SemanticNode)
	traverse = func(node *SemanticNode) {
		if node == nil {
			return
		}
		if node.Count() <= maxNodes {
			chunks = append(chunks, &SemanticTree{Root: node})
			return
		}

		if node.Content != "" || len(node.Attrs) > 0 {
			hollow := &SemanticNode{
				Type:       node.Type,
				Content:    node.Content,
				Attrs:      node.Attrs,
				OriginalID: node.OriginalID,
			}
			chunks = append(chunks, &SemanticTree{Root: hollow})
		}

		for _, child := range node.Children {
			traverse(child)
		}
	}

	traverse(t.Root)
	return chunks
}

func renderNode(sb *strings.Builder, node *SemanticNode, depth int, index string) {
	indent := strings.Repeat("  ", depth)

	keys := make([]string, 0, len(node.Attrs))
	for k := range node.Attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	attrStr := ""
	for _, k := range keys {
		attrStr += fmt.Sprintf(" %s=%s", k, node.Attrs[k])
	}

	if node.Content != "" {
		sb.WriteString(fmt.Sprintf("%s[%s%s][%s] %s\n", indent, node.Type, attrStr, index, node.Content))
	} else {
		sb.WriteString(fmt.Sprintf("%s[%s%s][%s]\n", indent, node.Type, attrStr, index))
	}

	for idx, child := range node.Children {
	    renderNode(sb, child, depth+1, index + strconv.Itoa(idx))
	}
}

func renderSkeleton(sb *strings.Builder, node *SemanticNode, depth int) {
	indent := strings.Repeat("  ", depth)
	if node.Type == "group" {
		sb.WriteString(fmt.Sprintf("%s[group]\n", indent))
		for _, child := range node.Children {
			renderSkeleton(sb, child, depth+1)
		}
	} else {
		sb.WriteString(fmt.Sprintf("%s[%s]\n", indent, node.Type))
	}
}

func CreateSemanticTree(doc *goquery.Document) *SemanticTree {
	root := &SemanticNode{Type: "group"}

	doc.Find("body").Each(func(i int, s *goquery.Selection) {
		nodes, _ := walkTree(s)
		root.Children = append(root.Children, nodes...)
	})

	return &SemanticTree{Root: root}
}

var keepAttrs = map[string]bool{
    "href":       true,
    "src":        true,
    "alt":        true,
    "datetime":   true,
    "poster":     true,
    "type":       true,
    "action":     true,
    "method":     true,
    "title":      true,
    "aria-label": true,
    "placeholder":true,
    "value":      true,
    "label":      true,
}

func getAttrs(s *goquery.Selection) map[string]string {
    attrs := map[string]string{}
    for _, attr := range s.Get(0).Attr {
        if keepAttrs[attr.Key] {
            attrs[attr.Key] = attr.Val
        }
    }
    return attrs
}

func walkTree(s *goquery.Selection) ([]*SemanticNode, bool) {
	tag := goquery.NodeName(s)

	if dropTags[tag] {
		return nil, false
	}


	directText := strings.TrimSpace(s.Clone().Children().Remove().End().Text())

	switch {

	case semanticTags[tag]:

		attrs := getAttrs(s)
	    node := &SemanticNode{Type: NodeType(tag), Attrs:   attrs}
	    if directText != "" {
	        node.Content = directText
	    }
	    s.Children().Each(func(i int, child *goquery.Selection) {
	        childNodes, _ := walkTree(child)
	        node.Children = append(node.Children, childNodes...)
	    })
	    if node.Content == "" && len(node.Children) == 0 && len(attrs) == 0{
	        return nil, false
	    }
	    return []*SemanticNode{node}, true

	default:
	    var innerNodes []*SemanticNode
	    innerSemantic := directText != ""

	    s.Children().Each(func(i int, child *goquery.Selection) {
	        childNodes, isSemantic := walkTree(child)
	        innerSemantic = innerSemantic || isSemantic
	        innerNodes = append(innerNodes, childNodes...)
	    })

	    if len(innerNodes) == 0 {
	        if directText != "" {
	            return []*SemanticNode{{Type: "group", Content: directText}}, true
	        }
	        return nil, false
	    }


	    if len(innerNodes) == 1 && directText == "" {
	        return innerNodes, innerSemantic
	    }

	    if directText == "" && len(innerNodes) == 1 && innerNodes[0].Type == "group" {
	        return innerNodes, innerSemantic
	    }

	    if innerSemantic {
	        group := &SemanticNode{Type: "group"}
	        if directText != "" {
	            group.Content = directText  // text lives on the group itself
	        }
	        group.Children = append(group.Children, innerNodes...)
	        return []*SemanticNode{group}, false
	    }

	    return innerNodes, false	
	}
}