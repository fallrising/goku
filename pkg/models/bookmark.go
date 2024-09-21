package models

import (
	"time"
)

type Bookmark struct {
	ID          int64     `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (b *Bookmark) AddTag(tag string) {
	for _, t := range b.Tags {
		if t == tag {
			return
		}
	}
	b.Tags = append(b.Tags, tag)
}

func (b *Bookmark) RemoveTag(tag string) {
	for i, t := range b.Tags {
		if t == tag {
			b.Tags = append(b.Tags[:i], b.Tags[i+1:]...)
			return
		}
	}
}
