// internal/fetcher/fetcher.go

package fetcher

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type PageContent struct {
	Title       string
	Description string
	Tags        []string
}

func FetchPageContent(url string) (*PageContent, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	content := &PageContent{
		Title:       extractTitle(doc),
		Description: extractDescription(doc),
		Tags:        extractTags(doc),
	}

	// If description is empty, try to extract first paragraph
	if content.Description == "" {
		content.Description = extractFirstParagraph(doc)
	}

	return content, nil
}

func extractTitle(n *html.Node) string {
	var title string
	var extractTitleFunc func(*html.Node)
	extractTitleFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractTitleFunc(c)
		}
	}
	extractTitleFunc(n)
	return strings.TrimSpace(title)
}

func extractDescription(n *html.Node) string {
	var description string
	var extractDescFunc func(*html.Node)
	extractDescFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			for _, a := range n.Attr {
				if a.Key == "name" && a.Val == "description" {
					for _, a := range n.Attr {
						if a.Key == "content" {
							description = a.Val
							return
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractDescFunc(c)
		}
	}
	extractDescFunc(n)
	return strings.TrimSpace(description)
}

func extractTags(n *html.Node) []string {
	var tags []string
	var extractTagsFunc func(*html.Node)
	extractTagsFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			for _, a := range n.Attr {
				if a.Key == "name" && (a.Val == "keywords" || a.Val == "tags") {
					for _, a := range n.Attr {
						if a.Key == "content" {
							tags = append(tags, strings.Split(a.Val, ",")...)
							return
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractTagsFunc(c)
		}
	}
	extractTagsFunc(n)

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

func extractFirstParagraph(n *html.Node) string {
	var paragraph string
	var extractParagraphFunc func(*html.Node)
	extractParagraphFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "p" {
			paragraph = renderNode(n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractParagraphFunc(c)
		}
	}
	extractParagraphFunc(n)
	return strings.TrimSpace(paragraph)
}

func renderNode(n *html.Node) string {
	var buf strings.Builder
	renderNodeFunc(&buf, n)
	return buf.String()
}

func renderNodeFunc(w *strings.Builder, n *html.Node) {
	if n.Type == html.TextNode {
		w.WriteString(n.Data)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderNodeFunc(w, c)
	}
}
