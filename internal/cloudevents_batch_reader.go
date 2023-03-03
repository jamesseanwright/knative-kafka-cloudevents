package internal

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/segmentio/kafka-go"
)

const (
	minMessageSizeBytes   = 250
	maxReadBatchSizeBytes = minMessageSizeBytes * 10
)

type CloudEventsBatchReader interface {
	ReadBatch(eventChan chan cloudevents.Event, errChan chan error)
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

func (r KafkaCloudEventsBatchReader) ReadBatch(eventChan chan cloudevents.Event, errChan chan error) {
	buf := make([]byte, minMessageSizeBytes)
	batch := r.conn.ReadBatch(minMessageSizeBytes, maxReadBatchSizeBytes)

	for {
		if n, err := batch.Read(buf); err != nil {
			errChan <- err
		} else {
			var event cloudevents.Event

			if err := format.JSON.Unmarshal(buf[:n], &event); err != nil {
				errChan <- fmt.Errorf("unmarshal received cloudevent: %w", err)
			} else {
				eventChan <- event
			}
		}
	}
}
