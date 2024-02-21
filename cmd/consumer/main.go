package main

import (
	"context"

	"github.com/jamesseanwright/knative-kafka-cloudevents/internal"
	"github.com/segmentio/kafka-go"
)

const (
	maxReadBytes    = 10e6 // 10 MB
	topic           = "test-events"
	consumerGroupID = "go-cloudevents-consumer"
	broker          = "kafka:9092"
)

func main() {
	ctx := context.Background() // TODO: context.WithCancel/SIGINT
	logger := internal.NewLogger()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{broker},
		Topic:     topic,
		Partition: 0,
		GroupID:   consumerGroupID,
		MaxBytes:  maxReadBytes,
	})

	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("reader close: %w", err)
		}
	}()

	eventReader := internal.NewKafkaCloudEventsReader(reader)
	consumer := internal.NewConsumer(eventReader, logger)

	consumer.Run(ctx)
}
