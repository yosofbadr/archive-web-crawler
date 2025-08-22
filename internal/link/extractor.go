package link

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func Extract(body io.Reader, baseURL *url.URL) ([]*url.URL, error) {
	tokenizer := html.NewTokenizer(body)
	var links []*url.URL

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			err := tokenizer.Err()
			if err == io.EOF {
				return links, nil
			}
			return nil, err
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						parsedLink, err := url.Parse(strings.TrimSpace(attr.Val))
						if err != nil {
							continue
						}
						absoluteLink := baseURL.ResolveReference(parsedLink)
						links = append(links, absoluteLink)
					}
				}
			}
		}
	}
}
