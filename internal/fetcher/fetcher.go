// internal/fetcher/fetcher.go

package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
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

func FetchPageContent(pageURL string) (*PageContent, bool, error) {
	// Validate URL structure
	parsedURL, err := url.ParseRequestURI(pageURL)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("Invalid URL format: %v", err)}, false, nil
	}

	// Check if the URL has a valid host
	if parsedURL.Host == "" {
		return &PageContent{FetchError: "URL must have a valid host"}, false, nil
	}

	if ValidateIfInternalIP(pageURL) {
		return &PageContent{FetchError: "Internal IP addresses are not supported"}, false, nil
	}

	alive, err := IsWebsiteAccessible(pageURL)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("Failed to check website accessibility: %v", err)}, true, nil
	}
	if !alive {
		return &PageContent{FetchError: "Website is not accessible"}, false, nil
	}

	client := &http.Client{
		Timeout: 250 * time.Millisecond,
	}

	resp, err := client.Get(pageURL)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("Failed to fetch URL: %v", err)}, false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &PageContent{FetchError: fmt.Sprintf("HTTP code: %d, cannot get metadata", resp.StatusCode)}, false, nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("Failed to parse HTML: %v", err)}, false, nil
	}

	content := &PageContent{
		Title:       extractTitle(doc),
		Description: extractDescription(doc, parsedURL.Host),
		Tags:        extractTags(doc),
	}

	return content, false, nil
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

func ValidateIfInternalIP(urlString string) bool {
	u, err := url.Parse(urlString)
	if err != nil {
		return false
	}

	host := u.Hostname()
	ip := net.ParseIP(host)
	if ip == nil {
		// If it's not a valid IP, try to resolve it
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return false
		}
		ip = ips[0]
	}

	return ip.IsLoopback() || ip.IsPrivate()
}

func CheckSiteAvailability(urlStr string, timeout time.Duration) (bool, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false, fmt.Errorf("invalid URL: %w", err)
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false, nil // Site is unavailable, but it's not an error in our logic
	}
	defer conn.Close()

	return true, nil
}

func IsWebsiteAccessible(url string) (bool, error) {
	accessible, err := CheckSiteAvailability(url, 1*time.Second)
	if err != nil {
		return false, err
	}
	return accessible, nil
}

type WaybackResponse struct {
	ArchivedSnapshots struct {
		Closest struct {
			Available bool   `json:"available"`
			URL       string `json:"url"`
			Timestamp string `json:"timestamp"`
		} `json:"closest"`
	} `json:"archived_snapshots"`
}

func FetchMetadataFromWaybackMachine(urlStr string) (*PageContent, error) {
	// Construct the Wayback Machine API URL
	apiURL := fmt.Sprintf("https://archive.org/wayback/available?url=%s", url.QueryEscape(urlStr))

	// Create an HTTP client with a timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Make the request to the Wayback Machine API
	resp, err := client.Get(apiURL)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("failed to fetch Wayback Machine API: %v", err)}, nil
	}
	defer resp.Body.Close()

	// Read and parse the JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("failed to read Wayback Machine response: %v", err)}, nil
	}

	var waybackResp WaybackResponse
	err = json.Unmarshal(body, &waybackResp)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("failed to parse Wayback Machine response: %v", err)}, nil
	}

	// Check if an archived version is available
	if !waybackResp.ArchivedSnapshots.Closest.Available {
		return &PageContent{FetchError: "No archived version available"}, nil
	}

	// Fetch the archived page
	archivedResp, err := client.Get(waybackResp.ArchivedSnapshots.Closest.URL)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("failed to fetch archived page: %v", err)}, nil
	}
	defer archivedResp.Body.Close()

	// Parse the HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return &PageContent{FetchError: fmt.Sprintf("Failed to parse HTML: %v", err)}, nil
	}

	// Extract title and description
	content := &PageContent{
		Title:       extractTitle(doc),
		Description: extractDescription(doc, urlStr),
		Tags:        extractTags(doc),
	}

	return content, nil
}
