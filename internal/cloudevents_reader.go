package internal

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/segmentio/kafka-go"
)

type CloudEventsReader interface {
	Read(ctx context.Context) (cloudevents.Event, error)
}

type kafkaMessageReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
}

type KafkaCloudEventsReader struct {
	reader kafkaMessageReader
}

func NewKafkaCloudEventsReader(reader kafkaMessageReader) KafkaCloudEventsReader {
	return KafkaCloudEventsReader{reader}
}

func (r KafkaCloudEventsReader) Read(ctx context.Context) (cloudevents.Event, error) {
	var event cloudevents.Event

	m, err := r.reader.ReadMessage(ctx)

	if err != nil {
		return event, err
	}

	if err = format.JSON.Unmarshal(m.Value, &event); err != nil {
		return event, err
	}

	return event, nil
}
