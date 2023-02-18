package internal

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/segmentio/kafka-go"
)

type kafkaConn interface {
	ReadBatch(int, int) *kafka.Batch
	WriteMessages(...kafka.Message) (int, error)
}

type KafkaCloudEventsProtocol struct {
	conn kafkaConn
}

func NewKafkaCloudEventsProtocol(conn kafkaConn) KafkaCloudEventsProtocol {
	return KafkaCloudEventsProtocol{conn}
}

func (p KafkaCloudEventsProtocol) Receive(context.Context) (binding.Message, error) {
	return nil, nil
}

func (p KafkaCloudEventsProtocol) Send(context.Context, binding.Message, ...binding.Transformer) error {
	return nil
}
