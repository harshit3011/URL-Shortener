package kafka

import (
	"testing"
)

func TestInitConsumerUsesKafkaAddress(t *testing.T) {
	t.Setenv("KAFKA_ADDR", "kafka:9092")
	InitConsumer()
	t.Cleanup(func() {
		if err := Reader.Close(); err != nil {
			t.Errorf("close Kafka reader: %v", err)
		}
	})

	brokers := Reader.Config().Brokers
	if len(brokers) != 1 || brokers[0] != "kafka:9092" {
		t.Fatalf("expected Kafka broker kafka:9092, got %v", brokers)
	}
}

func TestInitConsumerDefaultsToLocalKafkaAddress(t *testing.T) {
	t.Setenv("KAFKA_ADDR", "")
	InitConsumer()
	t.Cleanup(func() {
		if err := Reader.Close(); err != nil {
			t.Errorf("close Kafka reader: %v", err)
		}
	})

	brokers := Reader.Config().Brokers
	if len(brokers) != 1 || brokers[0] != "localhost:9092" {
		t.Fatalf("expected default Kafka broker localhost:9092, got %v", brokers)
	}
}
