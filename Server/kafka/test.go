package kafka

import (
	"context"
	"encoding/json"
	"time"
)

type ClickEvent struct {
	Shortcode string    `json:"shortcode"`
	ClickedAt time.Time `json:"clicked_at"`
}

func TestPublishClick() error {
	event := ClickEvent{
		Shortcode: "000001",
		ClickedAt: time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return PublishClick(context.Background(), data)
}