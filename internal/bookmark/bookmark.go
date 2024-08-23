package bookmark

import "strings"

// Bookmark represents a single bookmark entry.
type Bookmark struct {
	ID          int      `json:"id"`
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

// GetTagsString returns a comma-separated string of tags.
func (b *Bookmark) GetTagsString() string {
	return strings.Join(b.Tags, ",")
}
