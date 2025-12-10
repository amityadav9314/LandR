package scraper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Scraper struct {
	client *http.Client
}

func NewScraper() *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Scrape fetches the URL and extracts text content.
func (s *Scraper) Scrape(url string) (string, error) {
	log.Printf("[Scraper] Fetching URL: %s", url)

	// First try direct scraping
	content, err := s.directScrape(url)
	if err == nil && len(content) > 100 {
		return content, nil
	}
	log.Printf("[Scraper] Direct scrape failed or insufficient content, trying Jina Reader...")

	// Fallback: Use Jina AI Reader for JS-rendered sites
	content, err = s.jinaReaderScrape(url)
	if err == nil && len(content) > 100 {
		return content, nil
	}

	return "", fmt.Errorf("all scraping methods failed")
}

// directScrape uses goquery to extract content from static HTML
func (s *Scraper) directScrape(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch url: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[Scraper.Direct] Response status: %d", resp.StatusCode)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse html: %w", err)
	}

	// Remove unwanted elements
	doc.Find("script, style, nav, footer, header, aside, .sidebar, .advertisement, .ads").Remove()

	var sb strings.Builder

	// Try content selectors
	selectors := []string{"article", "[role='main']", "main", ".post-content", ".article-content", ".entry-content", ".content"}
	for _, selector := range selectors {
		selection := doc.Find(selector)
		if selection.Length() > 0 {
			log.Printf("[Scraper.Direct] Found content with selector: %s", selector)
			selection.Find("p, h1, h2, h3, li").Each(func(i int, s *goquery.Selection) {
				text := strings.TrimSpace(s.Text())
				if len(text) > 20 {
					sb.WriteString(text)
					sb.WriteString("\n\n")
				}
			})
			break
		}
	}

	// Fallback: all paragraphs
	if sb.Len() == 0 {
		doc.Find("body p").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if len(text) > 30 {
				sb.WriteString(text)
				sb.WriteString("\n\n")
			}
		})
	}

	return strings.TrimSpace(sb.String()), nil
}

// jinaReaderScrape uses Jina AI Reader to render JS and extract content
func (s *Scraper) jinaReaderScrape(url string) (string, error) {
	jinaURL := "https://r.jina.ai/" + url
	log.Printf("[Scraper.Jina] Fetching via Jina Reader: %s", jinaURL)

	req, err := http.NewRequest("GET", jinaURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create jina request: %w", err)
	}

	req.Header.Set("Accept", "text/plain")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("[Scraper.Jina] Request failed: %v", err)
		return "", fmt.Errorf("jina request failed: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[Scraper.Jina] Response status: %d", resp.StatusCode)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("jina status code error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read jina response: %w", err)
	}

	content := string(body)

	// Truncate if too long
	maxLen := 15000
	if len(content) > maxLen {
		log.Printf("[Scraper.Jina] Truncating from %d to %d chars", len(content), maxLen)
		content = content[:maxLen]
	}

	log.Printf("[Scraper.Jina] Successfully extracted %d characters", len(content))
	return content, nil
}
