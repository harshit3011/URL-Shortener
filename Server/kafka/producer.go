package kafka

import (
	"context"
	"errors"
	"os"

	"github.com/segmentio/kafka-go"
)

type messageWriter interface {
	WriteMessages(context.Context, ...kafka.Message) error
}

var Writer messageWriter

func brokerAddress() string {
	address := os.Getenv("KAFKA_ADDR")
	if address == "" {
		return "localhost:9092"
	}

	return address
}

func InitProducer() {
	Writer = &kafka.Writer{
		Addr:     kafka.TCP(brokerAddress()),
		Topic:    "url-clicked",
		Balancer: &kafka.LeastBytes{},
	}
}

func PublishClick(ctx context.Context, message []byte) error {
	if Writer == nil {
		return errors.New("Kafka producer is not initialized")
	}

	return Writer.WriteMessages(ctx, kafka.Message{
		Value: message,
	})
}
