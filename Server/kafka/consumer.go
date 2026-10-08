package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

var Reader *kafka.Reader

func InitConsumer() {
	Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{brokerAddress()},
		Topic:   "url-clicked",
		GroupID: "url-shortener-consumer",
	})
}

func ConsumeClicks(ctx context.Context) {
	for {
		message, err := Reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Println("Kafka consumer error:", err)
			return
		}

		log.Printf("Received Kafka message: %s", string(message.Value))
	}
}
