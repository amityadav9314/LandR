package scraper

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Scraper struct {
	client *http.Client
}

func NewScraper() *Scraper {
	return &Scraper{
		client: &http.Client{},
	}
}

// Scrape fetches the URL and extracts text content from paragraph tags.
func (s *Scraper) Scrape(url string) (string, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse html: %w", err)
	}

	var sb strings.Builder
	// Extract text from paragraphs. This is a simple heuristic.
	// We could improve this by looking at article tags, main, etc.
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if len(text) > 0 {
			sb.WriteString(text)
			sb.WriteString("\n\n")
		}
	})

	content := sb.String()
	if content == "" {
		return "", fmt.Errorf("no content found")
	}

	return content, nil
}
