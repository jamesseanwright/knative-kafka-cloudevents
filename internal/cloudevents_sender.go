package internal

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/segmentio/kafka-go"
)

type CloudEventsSender interface {
	Send(messages ...cloudevents.Event) error
}

type kafkaWriteConn interface {
	WriteMessages(...kafka.Message) (int, error)
}

type KafkaCloudEventsSender struct {
	conn kafkaWriteConn
}

func NewKafkaCloudEventsSender(conn kafkaWriteConn) KafkaCloudEventsSender {
	return KafkaCloudEventsSender{conn}
}

func (s KafkaCloudEventsSender) Send(messages ...cloudevents.Event) error {
	var kafkaMsgs []kafka.Message

	for _, v := range messages {
		b, err := format.JSON.Marshal(&v)

		if err != nil {
			return fmt.Errorf("marshal cloudevent: %w", err)
		}

		kafkaMsgs = append(kafkaMsgs, kafka.Message{
			Value: b,
		})
	}

	if _, err := s.conn.WriteMessages(kafkaMsgs...); err != nil {
		return fmt.Errorf("write kafka messages: %w", err)
	}

	return nil
}
