package bookmarks

import (
	"context"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fallrising/goku-cli/pkg/models"
	"github.com/schollz/progressbar/v3"
)

func (s *BookmarkService) ExportToHTML(ctx context.Context) (string, error) {
	const pageSize = 100 // Number of bookmarks to fetch per page

	// Get total count of bookmarks
	totalCount, err := s.CountBookmarks(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to count bookmarks: %w", err)
	}

	bar := progressbar.Default(int64(totalCount))

	var sb strings.Builder

	// Write HTML header
	sb.WriteString("<!DOCTYPE NETSCAPE-Bookmark-file-1>\n")
	sb.WriteString("<META HTTP-EQUIV=\"Content-Type\" CONTENT=\"text/html; charset=UTF-8\">\n")
	sb.WriteString("<TITLE>Bookmarks</TITLE>\n")
	sb.WriteString("<H1>Bookmarks</H1>\n")
	sb.WriteString("<DL><p>\n")

	// Fetch and write bookmarks in batches
	for offset := 0; offset < totalCount; offset += pageSize {
		bookmarks, err := s.ListBookmarks(ctx, pageSize, offset)
		if err != nil {
			return "", fmt.Errorf("failed to fetch bookmarks at offset %d: %w", offset, err)
		}

		for _, bookmark := range bookmarks {
			sb.WriteString(fmt.Sprintf("    <DT><A HREF=\"%s\" ADD_DATE=\"%d\">%s</A>\n",
				html.EscapeString(bookmark.URL),
				bookmark.CreatedAt.Unix(),
				html.EscapeString(bookmark.Title)))

			if bookmark.Description != "" {
				sb.WriteString(fmt.Sprintf("    <DD>%s\n", html.EscapeString(bookmark.Description)))
			}
			bar.Add(1)
		}
	}

	// Close HTML
	sb.WriteString("</DL><p>")

	return sb.String(), nil
}

func (s *BookmarkService) ImportFromHTML(ctx context.Context, r io.Reader) (int, error) {
	log.Println("Starting ImportFromHTML process")
	content, err := io.ReadAll(r)
	if err != nil {
		log.Printf("Error reading HTML content: %v", err)
		return 0, fmt.Errorf("failed to read HTML content: %w", err)
	}
	log.Printf("Read %d bytes of HTML content", len(content))

	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		log.Printf("Error parsing HTML: %v", err)
		return 0, fmt.Errorf("failed to parse HTML: %w", err)
	}
	log.Println("Successfully parsed HTML content")

	uniqueURLs := make(map[string]struct{})
	var uniqueBookmarks []*models.Bookmark

	// First pass: extract unique bookmarks
	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			var url, title string
			var addDate int64

			for _, attr := range n.Attr {
				switch attr.Key {
				case "href":
					url = attr.Val
				case "add_date":
					addDate, _ = parseAddDate(attr.Val)
				}
			}

			if n.FirstChild != nil {
				title = n.FirstChild.Data
			}

			if url != "" {
				if _, exists := uniqueURLs[url]; !exists {
					uniqueURLs[url] = struct{}{}
					bookmark := &models.Bookmark{
						URL:   url,
						Title: title,
					}
					if addDate != 0 {
						bookmark.CreatedAt = time.Unix(addDate, 0)
					}
					uniqueBookmarks = append(uniqueBookmarks, bookmark)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(doc)
	log.Printf("Found %d unique bookmarks to import", len(uniqueBookmarks))

	bar := progressbar.NewOptions(len(uniqueBookmarks),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][1/1][reset] Importing bookmarks..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	bookmarkChan := make(chan *models.Bookmark, 100)
	resultChan := make(chan error, 100)
	var wg sync.WaitGroup

	// Start worker goroutines
	const numWorkers = 5
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for bookmark := range bookmarkChan {
				err := s.CreateBookmark(ctx, bookmark)
				if err != nil {
					resultChan <- fmt.Errorf("worker %d failed to import bookmark %s: %w", workerID, bookmark.URL, err)
				} else {
					resultChan <- nil
					bar.Add(1)
				}
			}
		}(i)
	}

	// Send bookmarks to workers
	go func() {
		for _, bookmark := range uniqueBookmarks {
			bookmarkChan <- bookmark
		}
		close(bookmarkChan)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results
	var errors []error
	for err := range resultChan {
		if err != nil {
			errors = append(errors, err)
		}
	}

	fmt.Println() // Add a newline after the progress bar

	recordsCreated := len(uniqueBookmarks) - len(errors)
	log.Printf("Import summary: %d records created, %d errors", recordsCreated, len(errors))

	if len(errors) > 0 {
		for i, err := range errors {
			log.Printf("Error %d: %v", i+1, err)
		}
		return recordsCreated, fmt.Errorf("encountered %d errors during import", len(errors))
	}

	// Verify the import by counting records in the database
	totalRecords, err := s.CountBookmarks(ctx)
	if err != nil {
		log.Printf("Error counting bookmarks after import: %v", err)
		return recordsCreated, fmt.Errorf("failed to verify import: %w", err)
	}
	log.Printf("Total records in database after import: %d", totalRecords)

	return recordsCreated, nil
}

func (s *BookmarkService) CountBookmarks(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

func parseAddDate(date string) (int64, error) {
	// First, try parsing as Unix timestamp
	i, err := parseInt64(date)
	if err == nil {
		return i, nil
	}

	// If that fails, try parsing as RFC3339 format
	t, err := time.Parse(time.RFC3339, date)
	if err == nil {
		return t.Unix(), nil
	}

	// If all parsing attempts fail, return 0 (which will use current time)
	return 0, fmt.Errorf("unable to parse date: %s", date)
}

func parseInt64(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}