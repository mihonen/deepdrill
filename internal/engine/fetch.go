package engine

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

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

func (e *Engine) clean(doc *goquery.Document) {
	doc.Find("script, style, noscript, iframe, head, svg, link").Remove()
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		s.RemoveAttr("class")
		s.RemoveAttr("id")
		s.RemoveAttr("style")
		if s.Is("a") {
			href, _ := s.Attr("href")
			s.SetAttr("href", href)
		}
	})
}
