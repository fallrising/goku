package bookmarks

import (
	"context"
	"fmt"
	"github.com/fallrising/goku-cli/internal/fetcher"
	"github.com/fallrising/goku-cli/pkg/interfaces"
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

type BookmarkService struct {
	repo interfaces.BookmarkRepository
}

func NewBookmarkService(repo interfaces.BookmarkRepository) *BookmarkService {
	return &BookmarkService{repo: repo}
}

func (s *BookmarkService) CreateBookmark(ctx context.Context, bookmark *models.Bookmark) error {
	log.Printf("CreateBookmark called with URL: %s", bookmark.URL)

	if bookmark.URL == "" {
		log.Println("Error: URL is required")
		return fmt.Errorf("URL is required")
	}

	// Check if URL already exists in the database
	existingBookmark, err := s.repo.GetByURL(ctx, bookmark.URL)
	if err != nil {
		log.Printf("Error checking for existing bookmark: %v", err)
		return fmt.Errorf("failed to check for existing bookmark: %w", err)
	}
	if existingBookmark != nil {
		log.Printf("Bookmark already exists with URL: %s", existingBookmark.URL)
		return fmt.Errorf("bookmark with this URL already exists: %s", existingBookmark.URL)
	}

	// Check if URL starts with "http://" or "https://"
	if !(strings.HasPrefix(bookmark.URL, "http://") || strings.HasPrefix(bookmark.URL, "https://")) {
		bookmark.URL = "https://" + bookmark.URL
		log.Printf("URL updated to: %s", bookmark.URL)
	}

	// Fetch page content if title, description, or tags are not provided
	if bookmark.Title == "" || bookmark.Description == "" || len(bookmark.Tags) == 0 {
		log.Println("Fetching page content for metadata")
		content, err := fetcher.FetchPageContent(bookmark.URL)
		if err != nil {
			log.Printf("Warning: failed to fetch page content: %v", err)
		}

		if content != nil && content.FetchError != "" {
			log.Printf("Warning: %s", content.FetchError)
			bookmark.Description = fmt.Sprintf("Metadata fetch failed: %s", content.FetchError)
		} else if content != nil {
			if bookmark.Title == "" {
				bookmark.Title = content.Title
				log.Printf("Title set from fetched content: %s", bookmark.Title)
			}
			if bookmark.Description == "" {
				bookmark.Description = content.Description
				log.Printf("Description set from fetched content: %s", bookmark.Description)
			}
			if len(bookmark.Tags) == 0 {
				bookmark.Tags = content.Tags
				log.Printf("Tags set from fetched content: %v", bookmark.Tags)
			}
		}
	}

	log.Printf("Attempting to create bookmark in repository: %+v", bookmark)
	err = s.repo.Create(ctx, bookmark)
	if err != nil {
		log.Printf("Error creating bookmark in repository: %v", err)
		return fmt.Errorf("failed to create bookmark in repository: %w", err)
	}

	log.Printf("Bookmark successfully created with ID: %d", bookmark.ID)
	return nil
}

func (s *BookmarkService) GetBookmark(ctx context.Context, id int64) (*models.Bookmark, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *BookmarkService) UpdateBookmark(ctx context.Context, updatedBookmark *models.Bookmark) error {
	if updatedBookmark.ID == 0 {
		return fmt.Errorf("bookmark ID is required")
	}

	// Fetch existing bookmark
	existingBookmark, err := s.repo.GetByID(ctx, updatedBookmark.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing bookmark: %w", err)
	}
	if existingBookmark == nil {
		return fmt.Errorf("bookmark not found with ID: %d", updatedBookmark.ID)
	}

	// Check if the URL has changed
	if updatedBookmark.URL != "" && updatedBookmark.URL != existingBookmark.URL {
		// Check for duplicates
		duplicate, err := s.repo.GetByURL(ctx, updatedBookmark.URL)
		if err != nil {
			return fmt.Errorf("failed to check for duplicate URL: %w", err)
		}
		if duplicate != nil {
			return fmt.Errorf("another bookmark with URL '%s' already exists", updatedBookmark.URL)
		}

		// Fetch new metadata for the new URL
		content, err := fetcher.FetchPageContent(updatedBookmark.URL)
		if err != nil {
			return fmt.Errorf("failed to fetch metadata for the updated URL: %w", err)
		}

		if content.FetchError != "" {
			fmt.Printf("Warning: %s\n", content.FetchError)
			updatedBookmark.Description = fmt.Sprintf("Metadata fetch failed: %s", content.FetchError)
		} else {
			// Update the metadata with fetched content
			updatedBookmark.Title = content.Title
			updatedBookmark.Description = content.Description
			updatedBookmark.Tags = content.Tags
		}
	}

	// Track changes
	updated := false

	if updatedBookmark.URL != "" && updatedBookmark.URL != existingBookmark.URL {
		existingBookmark.URL = updatedBookmark.URL
		updated = true
	}
	if updatedBookmark.Title != "" && updatedBookmark.Title != existingBookmark.Title {
		existingBookmark.Title = updatedBookmark.Title
		updated = true
	}
	if updatedBookmark.Description != "" && updatedBookmark.Description != existingBookmark.Description {
		existingBookmark.Description = updatedBookmark.Description
		updated = true
	}
	if len(updatedBookmark.Tags) > 0 && !equalTags(updatedBookmark.Tags, existingBookmark.Tags) {
		existingBookmark.Tags = updatedBookmark.Tags
		updated = true
	}

	// Update only if necessary
	if updated {
		return s.repo.Update(ctx, existingBookmark)
	}

	fmt.Println("No changes detected, bookmark update skipped.")
	return nil
}

// Helper function to check if tags are equal
func equalTags(tags1, tags2 []string) bool {
	if len(tags1) != len(tags2) {
		return false
	}
	for i, tag := range tags1 {
		if tag != tags2[i] {
			return false
		}
	}
	return true
}

func (s *BookmarkService) DeleteBookmark(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *BookmarkService) ListBookmarks(ctx context.Context, limit, offset int) ([]*models.Bookmark, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *BookmarkService) ExportToHTML(ctx context.Context) (string, error) {
	bookmarks, err := s.ListBookmarks(ctx, 0, 0) // Get all bookmarks
	if err != nil {
		return "", fmt.Errorf("failed to fetch bookmarks: %w", err)
	}

	bar := progressbar.Default(int64(len(bookmarks)))

	var sb strings.Builder

	// Write HTML header
	sb.WriteString("<!DOCTYPE NETSCAPE-Bookmark-file-1>\n")
	sb.WriteString("<META HTTP-EQUIV=\"Content-Type\" CONTENT=\"text/html; charset=UTF-8\">\n")
	sb.WriteString("<TITLE>Bookmarks</TITLE>\n")
	sb.WriteString("<H1>Bookmarks</H1>\n")
	sb.WriteString("<DL><p>\n")

	// Write bookmarks
	for _, bookmark := range bookmarks {
		sb.WriteString(fmt.Sprintf("    <DT><A HREF=\"%s\" ADD_DATE=\"%d\">%s</A>\n",
			bookmark.URL,
			bookmark.CreatedAt.Unix(),
			bookmark.Title))

		if bookmark.Description != "" {
			sb.WriteString(fmt.Sprintf("    <DD>%s\n", bookmark.Description))
		}
		bar.Add(1)
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

	bookmarkCount := countBookmarksFromHTML(doc)
	log.Printf("Found %d bookmarks to import", bookmarkCount)

	bar := progressbar.NewOptions(bookmarkCount,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][1/3][reset] Importing bookmarks..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))

	bookmarkChan := make(chan *models.Bookmark, 100)
	resultChan := make(chan struct {
		success bool
		err     error
	}, 100)
	doneChan := make(chan struct{})

	// Use a WaitGroup to ensure all workers have finished
	var wg sync.WaitGroup

	// Start worker goroutines
	const numWorkers = 5 // Reduced number of workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Printf("Worker %d started", workerID)
			for bookmark := range bookmarkChan {
				err := s.CreateBookmark(ctx, bookmark)
				if err != nil {
					log.Printf("Worker %d failed to import bookmark %s: %v", workerID, bookmark.URL, err)
					resultChan <- struct {
						success bool
						err     error
					}{false, fmt.Errorf("failed to import bookmark %s: %w", bookmark.URL, err)}
				} else {
					log.Printf("Worker %d successfully imported bookmark: %s", workerID, bookmark.URL)
					resultChan <- struct {
						success bool
						err     error
					}{true, nil}
				}
			}
			log.Printf("Worker %d finished", workerID)
		}(i)
	}

	// Start a goroutine to close channels when traversal is done
	go func() {
		log.Println("Starting HTML traversal")
		var traverse func(*html.Node)
		traverse = func(n *html.Node) {
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
					bookmark := &models.Bookmark{
						URL:   url,
						Title: title,
					}

					if addDate != 0 {
						bookmark.CreatedAt = time.Unix(addDate, 0)
					}

					log.Printf("Sending bookmark to channel: %s", url)
					bookmarkChan <- bookmark
				}
			}

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				traverse(c)
			}
		}

		traverse(doc)
		log.Println("HTML traversal complete, closing bookmark channel")
		close(bookmarkChan)

		// Wait for all workers to finish
		wg.Wait()
		log.Println("All workers finished, closing result channel")
		close(resultChan)
		close(doneChan)
	}()

	// Handle results and errors
	recordsCreated := 0
	var errors []error
loop:
	for {
		select {
		case result := <-resultChan:
			if result.success {
				recordsCreated++
				bar.Add(1)
			} else {
				errors = append(errors, result.err)
			}
		case <-doneChan:
			log.Println("Import process complete")
			break loop
		}
	}

	fmt.Println() // Add a newline after the progress bar

	log.Printf("Import summary: %d records created, %d errors", recordsCreated, len(errors))

	if len(errors) > 0 {
		log.Printf("Encountered errors during import: %v", errors)
		return recordsCreated, fmt.Errorf("encountered %d errors during import", len(errors))
	}

	// Verify the import by counting records in the database
	totalRecords, err := s.CountBookmarks(ctx)
	if err != nil {
		log.Printf("Error counting bookmarks after import: %v", err)
		return recordsCreated, fmt.Errorf("failed to verify import: %w", err)
	}
	log.Printf("Total records in database after import: %d", totalRecords)

	if totalRecords < recordsCreated {
		log.Printf("Warning: Mismatch between imported records (%d) and database count (%d)", recordsCreated, totalRecords)
	}

	return recordsCreated, nil
}

// New method to count total bookmarks in the database
func (s *BookmarkService) CountBookmarks(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// countBookmarksFromHTML traverses the HTML structure and counts the number of <a> tags,
// which represent bookmarks in the HTML bookmark file format.
func countBookmarksFromHTML(n *html.Node) int {
	count := 0
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			count++
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(n)
	return count
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

func (s *BookmarkService) SearchBookmarks(ctx context.Context, query string, limit, offset int) ([]*models.Bookmark, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	bookmarks, err := s.repo.Search(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search bookmarks: %w", err)
	}

	if len(bookmarks) == 0 {
		fmt.Println("No bookmarks found matching the query.")
	}

	return bookmarks, nil
}

func (s *BookmarkService) RemoveTagFromBookmark(ctx context.Context, bookmarkID int64, tagToRemove string) error {
	// Fetch the existing bookmark
	bookmark, err := s.repo.GetByID(ctx, bookmarkID)
	if err != nil {
		return fmt.Errorf("failed to fetch bookmark: %w", err)
	}

	// Remove the specified tag
	bookmark.RemoveTag(tagToRemove)

	// Save the updated bookmark
	err = s.repo.Update(ctx, bookmark)
	if err != nil {
		return fmt.Errorf("failed to update bookmark after removing tag: %w", err)
	}

	return nil
}

func (s *BookmarkService) ListAllTags(ctx context.Context) ([]string, error) {
	tags, err := s.repo.ListAllTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	return tags, nil
}

func (s *BookmarkService) GetStatistics(ctx context.Context) (*models.Statistics, error) {
	stats := &models.Statistics{}
	var err error

	stats.HostnameCounts, err = s.repo.CountByHostname(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname counts: %w", err)
	}

	stats.TagCounts, err = s.repo.CountByTag(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tag counts: %w", err)
	}

	stats.LatestBookmarks, err = s.repo.GetLatest(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest bookmarks: %w", err)
	}

	stats.AccessibilityCounts, err = s.repo.CountAccessibility(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get accessibility counts: %w", err)
	}

	stats.TopHostnames, err = s.repo.TopHostnames(ctx, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to get top hostnames: %w", err)
	}

	stats.UniqueHostnames, err = s.repo.ListUniqueHostnames(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique hostnames: %w", err)
	}

	stats.CreatedLastWeek, err = s.repo.CountCreatedLastNDays(ctx, 7)
	if err != nil {
		return nil, fmt.Errorf("failed to get created counts for last week: %w", err)
	}

	return stats, nil
}
