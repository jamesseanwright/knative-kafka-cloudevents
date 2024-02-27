package main

import (
	"context"

	"github.com/jamesseanwright/knative-kafka-cloudevents/internal"
	"github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background() // TODO: context.WithCancel/SIGINT
	logger := internal.NewLogger()
	conn, err := kafka.DialLeader(ctx, "tcp", "kafka:9092", "test-events", 0)

	if err != nil {
		logger.Fatal(err)
	}

	eventSender := internal.NewKafkaCloudEventsSender(conn)
	producer := internal.NewProducer(eventSender, logger)

	producer.Run(ctx)
}
