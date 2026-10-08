package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

type testWriter struct {
	messages []kafkago.Message
}

func (w *testWriter) WriteMessages(_ context.Context, messages ...kafkago.Message) error {
	w.messages = append(w.messages, messages...)
	return nil
}

func TestKafkaProducer(t *testing.T) {
	writer := &testWriter{}
	Writer = writer

	err := TestPublishClick()
	if err != nil {
		t.Fatal(err)
	}

	if len(writer.messages) != 1 {
		t.Fatalf("expected 1 Kafka message, got %d", len(writer.messages))
	}

	var event ClickEvent
	if err := json.Unmarshal(writer.messages[0].Value, &event); err != nil {
		t.Fatalf("decode published click event: %v", err)
	}
	if event.Shortcode != "000001" {
		t.Errorf("expected shortcode 000001, got %q", event.Shortcode)
	}
	if event.ClickedAt.IsZero() || event.ClickedAt.After(time.Now()) {
		t.Errorf("expected a valid click timestamp, got %v", event.ClickedAt)
	}
}
