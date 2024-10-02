// internal/fetcher/fetcher.go

package fetcher

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type PageContent struct {
	Title       string
	Description string
	Tags        []string
	FetchError  string
}

func FetchPageContent(pageURL string) (*PageContent, error) {
	// Validate URL structure
	parsedURL, err := url.ParseRequestURI(pageURL)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("Invalid URL format: %v", err)}, nil
	}

	// Check if the URL has a valid host
	if parsedURL.Host == "" {
		return &PageContent{FetchError: "URL must have a valid host"}, nil
	}

	client := &http.Client{
		Timeout: 200 * time.Millisecond,
	}

	resp, err := client.Get(pageURL)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("Failed to fetch URL: %v", err)}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &PageContent{FetchError: fmt.Sprintf("HTTP code: %d, cannot get metadata", resp.StatusCode)}, nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("Failed to parse HTML: %v", err)}, nil
	}

	content := &PageContent{
		Title:       extractTitle(doc),
		Description: extractDescription(doc, parsedURL.Host),
		Tags:        extractTags(doc),
	}

	return content, nil
}

func extractTitle(doc *goquery.Document) string {
	title := doc.Find("title").First().Text()
	return strings.TrimSpace(title)
}

func extractDescription(doc *goquery.Document, host string) string {
	// Try standard meta description
	description, _ := doc.Find("meta[name='description']").Attr("content")
	if description != "" {
		return strings.TrimSpace(description)
	}

	// Try Open Graph description
	description, _ = doc.Find("meta[property='og:description']").Attr("content")
	if description != "" {
		return strings.TrimSpace(description)
	}

	// Special handling for known sites
	switch {
	case strings.Contains(host, "news.ycombinator.com"):
		description = extractHackerNewsDescription(doc)
	default:
		// For other sites, try to get the first paragraph or heading
		description = doc.Find("p, h1, h2").First().Text()
	}

	return strings.TrimSpace(description)
}

func extractHackerNewsDescription(doc *goquery.Document) string {
	title := doc.Find("td.title").First().Text()
	return strings.TrimSpace(title)
}

func extractTags(doc *goquery.Document) []string {
	var tags []string

	// Try to get tags from meta keywords
	keywords, _ := doc.Find("meta[name='keywords']").Attr("content")
	if keywords != "" {
		tags = append(tags, strings.Split(keywords, ",")...)
	}

	// Try to get tags from meta tags
	metaTags, _ := doc.Find("meta[name='tags']").Attr("content")
	if metaTags != "" {
		tags = append(tags, strings.Split(metaTags, ",")...)
	}

	// Clean and deduplicate tags
	uniqueTags := make(map[string]bool)
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag != "" {
			uniqueTags[tag] = true
		}
	}

	var cleanedTags []string
	for tag := range uniqueTags {
		cleanedTags = append(cleanedTags, tag)
	}

	return cleanedTags
}
