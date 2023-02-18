package main

import (
	"context"

	v2 "github.com/cloudevents/sdk-go/v2"
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

	protocol := internal.NewKafkaCloudEventsProtocol(conn)
	client, err := v2.NewClient(protocol)

	if err != nil {
		logger.Fatal(client)
	}

	producer := internal.NewProducer(client)

	if err := producer.Run(ctx); err != nil {
		logger.Fatal(err)
	}
}
