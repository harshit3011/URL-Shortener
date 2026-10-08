package models

import "time"

type Click struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Shortcode string    `json:"shortcode" bson:"shortcode"`
	ClickedAt time.Time `json:"clicked_at" bson:"clicked_at"`
}
