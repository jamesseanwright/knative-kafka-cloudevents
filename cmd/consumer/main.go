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

	eventReader := internal.NewKafkaCloudEventsBatchReader(conn)
	consumer := internal.NewConsumer(eventReader, logger)

	consumer.Run(ctx)
}
