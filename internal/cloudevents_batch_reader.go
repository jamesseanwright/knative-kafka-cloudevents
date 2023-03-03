package internal

import (
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/segmentio/kafka-go"
)

const (
	minMessageSizeBytes   = 210
	maxReadBatchSizeBytes = minMessageSizeBytes * 5
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
	batch := r.conn.ReadBatch(minMessageSizeBytes, maxReadBatchSizeBytes)
	buf := make([]byte, minMessageSizeBytes)

	for {
		n, err := batch.Read(buf)

		if err != nil {
			errChan <- fmt.Errorf("kafka batch read: %w", err)
			break
		}

		var event cloudevents.Event

		if err := format.JSON.Unmarshal(buf[:n], &event); err != nil {
			errChan <- fmt.Errorf("unmarshal received cloudevent: %w", err)
		} else {
			eventChan <- event
		}
	}

	if err := batch.Close(); err != nil {
		errChan <- fmt.Errorf("kafka batch close: %w", err)
	}

	// TODO: should be really be recursing once the initial
	// batch is empty? Might not matter when the consumer is
	// invoked through Knative, but it'd be worth verifying best
	// practices around batch sizes, keeping topics hydrated etc.
	r.ReadBatch(eventChan, errChan)
}
