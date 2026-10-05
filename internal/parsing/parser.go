package parsing

import (
	"io"
	"net/url"

	"golang.org/x/net/html"
)

// ExtractLinks parses HTML and returns all href links found.
func ExtractLinks(base *url.URL, r io.Reader) ([]string, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var links []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					ref := attr.Val
					if ref == "" || ref == "#" || ref == "/#" {
						break
					}
					u, err := url.Parse(ref)
					if err == nil {
						abs := base.ResolveReference(u)
						links = append(links, abs.String())
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return links, nil
}
