package models

import "time"

type URL struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	OriginalURL  string    `json:"original_url" bson:"original_url"`
	ShortenedURL string    `json:"shortened_url" bson:"shortened_url"`
	BelongsTo    string    `json:"belongs_to" bson:"belongs_to"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}
