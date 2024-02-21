package internal

import (
	"errors"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/segmentio/kafka-go"
)

const (
	minMessageSizeBytes   = 210
	maxReadBatchSizeBytes = minMessageSizeBytes * 5
)

type CloudEventsBatchReader interface {
	ReadBatch() ([]cloudevents.Event, error)
}

type kafkaReadBatchConn interface {
	ReadBatch(minBytes, maxBytes int) *kafka.Batch
}

type KafkaCloudEventsBatchReader struct {
	conn kafkaReadBatchConn
}

func NewKafkaCloudEventsBatchReader(conn kafkaReadBatchConn) KafkaCloudEventsBatchReader {
	return KafkaCloudEventsBatchReader{conn}
}

func (r KafkaCloudEventsBatchReader) ReadBatch() (events []cloudevents.Event, err error) {
	batch := r.conn.ReadBatch(minMessageSizeBytes, maxReadBatchSizeBytes)
	buf := make([]byte, minMessageSizeBytes)

	defer func() {
		err = errors.Join(err, batch.Close())
	}()

	for {
		var event cloudevents.Event
		var n int

		n, err = batch.Read(buf)

		if err != nil {
			return
		}

		if err = format.JSON.Unmarshal(buf[:n], &event); err != nil {
			return
		}

		events = append(events, event)
	}
}
