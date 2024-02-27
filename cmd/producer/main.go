package main

import (
	"context"
	"os"

	"github.com/jamesseanwright/knative-kafka-cloudevents/internal"
	"github.com/segmentio/kafka-go"
)

const (
	topic     = "test-events"
	partition = 0
)

func main() {
	ctx := context.Background() // TODO: context.WithCancel/SIGINT
	logger := internal.NewLogger()
	broker := os.Getenv("KAFKA_BROKER")

	if broker == "" {
		logger.Fatal("KAFKA_BROKER environment variable is missing")
	}

	conn, err := kafka.DialLeader(ctx, "tcp", broker, topic, partition)

	if err != nil {
		logger.Fatal(err)
	}

	eventSender := internal.NewKafkaCloudEventsSender(conn)
	producer := internal.NewProducer(eventSender, logger)

	producer.Run(ctx)
}
