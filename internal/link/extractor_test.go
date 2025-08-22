package link

import (
	"net/url"
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com")
	body := `
		<html>
			<body>
				<a href="/about">About</a>
				<a href="https://example.com/contact">Contact</a>
				<a href="https://google.com">Google</a>
			</body>
		</html>
	`

	links, err := Extract(strings.NewReader(body), baseURL)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedLinks := []string{
		"https://example.com/about",
		"https://example.com/contact",
		"https://google.com",
	}

	if len(links) != len(expectedLinks) {
		t.Fatalf("Expected %d links, but got %d", len(expectedLinks), len(links))
	}

	for i, link := range links {
		if link.String() != expectedLinks[i] {
			t.Errorf("Expected link to be '%s', but got '%s'", expectedLinks[i], link.String())
		}
	}
}
