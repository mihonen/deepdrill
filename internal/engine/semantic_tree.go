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
		"p": true, "span": true, "h1": true, "h2": true, "h3": true,
		"h4": true, "h5": true, "h6": true,
	    "td": true, "th": true, "time": true,
		"figcaption": true, "blockquote": true, "article": true,
		"form": true,
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

func (n *SemanticNode) Value(nodeType string) string {
    switch nodeType {
    case "link":
        return n.Attrs["href"]
    case "src":
        return n.Attrs["src"]
    case "datetime":
        return n.Attrs["datetime"]
    default:
        return n.Content
    }
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
	case "link":
		return "a"
	case "group":
		return "div"
	case "image":
		return "img"
	default:
		return string(node.Type) // p, span, h1, h2, time, figcaption, etc.
	}
}

func htmlAttrs(node *SemanticNode, index string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(` id="%s"`, index))
	switch node.Type {
	case "link":
		sb.WriteString(fmt.Sprintf(` href="%s"`, node.Attrs["href"]))
	case "image":
		sb.WriteString(fmt.Sprintf(` src="%s" alt="%s"`, node.Attrs["src"], node.Attrs["alt"]))
	case "time":
		sb.WriteString(fmt.Sprintf(` datetime="%s"`, node.Attrs["datetime"]))
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

func walkTree(s *goquery.Selection) ([]*SemanticNode, bool) {
	tag := goquery.NodeName(s)

	if dropTags[tag] {
		return nil, false
	}

	href, hasHref := s.Attr("href")
	src, hasSrc := s.Attr("src")
	poster, _ := s.Attr("poster")
	alt, _ := s.Attr("alt")
	datetime, _ := s.Attr("datetime")

	directText := strings.TrimSpace(s.Clone().Children().Remove().End().Text())

	switch {
	case tag == "a" && hasHref:
		node := &SemanticNode{
			Type:    "link",
			Content: directText,
			Attrs:   map[string]string{"href": href},
		}
		s.Children().Each(func(i int, child *goquery.Selection) {
			childNodes, _ := walkTree(child)
			node.Children = append(node.Children, childNodes...)
		})
		return []*SemanticNode{node}, true

	case tag == "img":
		return []*SemanticNode{{
			Type:  "image",
			Attrs: map[string]string{"src": src, "alt": alt},
		}}, true

	case tag == "time":
		return []*SemanticNode{{
			Type:    "time",
			Content: strings.TrimSpace(s.Text()),
			Attrs:   map[string]string{"datetime": datetime},
		}}, true

	case tag == "figcaption":
		text := strings.TrimSpace(s.Text())
		if text == "" {
			return nil, false
		}
		return []*SemanticNode{{
			Type:    "figcaption",
			Content: text,
		}}, true

	case tag == "video":
		node := &SemanticNode{Type: "video", Attrs: map[string]string{}}
		if hasSrc && src != "" {
			node.Attrs = map[string]string{"src": src, "poster": poster}
			return []*SemanticNode{node}, true
		}
		s.Find("source").Each(func(i int, source *goquery.Selection) {
			if i > 0 {
				return
			}
			srcVal, _ := source.Attr("src")
			typeVal, _ := source.Attr("type")
			if srcVal != "" {
				node.Attrs = map[string]string{"src": srcVal, "type": typeVal, "poster": poster}
			}
		})
		return []*SemanticNode{node}, true

	case tag == "button":
		text := strings.TrimSpace(s.Text())
		if text == "" {
			return nil, false
		}
		return []*SemanticNode{{
			Type:    "button",
			Content: text,
		}}, true

	case semanticTags[tag]:
	    node := &SemanticNode{Type: NodeType(tag)}
	    if directText != "" {
	        node.Content = directText
	    }
	    s.Children().Each(func(i int, child *goquery.Selection) {
	        childNodes, _ := walkTree(child)
	        node.Children = append(node.Children, childNodes...)
	    })
	    if node.Content == "" && len(node.Children) == 0 {
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