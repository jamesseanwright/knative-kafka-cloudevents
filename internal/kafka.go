package internal

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/segmentio/kafka-go"
)

const (
	cloudEventPrefix = "ce_"
	kibibyte         = 1024
	mebibyte         = 1024 * kibibyte
	readMinBatchSize = 10 * kibibyte
	readMaxBatchSize = 10 * mebibyte
)

type message struct {
	Value       []byte
	Headers     map[string]string
	ContentType string
	version     spec.Version
}

func (m message) ReadEncoding() binding.Encoding {
	return binding.EncodingBinary
}

func (m message) ReadStructured(context.Context, binding.StructuredWriter) error {
	return binding.ErrNotStructured
}

func (m message) ReadBinary(ctx context.Context, writer binding.BinaryWriter) error {
	for k, v := range m.Headers {
		if strings.HasPrefix(k, cloudEventPrefix) {
			var err error

			if attr := m.version.Attribute(k); attr != nil {
				err = writer.SetAttribute(attr, v)
			} else {
				err = writer.SetExtension(strings.TrimPrefix(k, cloudEventPrefix), v)
			}

			if err != nil {
				return fmt.Errorf("kafka cloudevent parse header: %w", err)
			}
		}
	}

	if m.Value != nil {
		if err := writer.SetData(bytes.NewBuffer(m.Value)); err != nil {
			return fmt.Errorf("kafka cloudevent parse data: %w", err)
		}
	}

	return nil
}

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
	batch := p.conn.ReadBatch(readMinBatchSize, readMaxBatchSize)

}

func (p KafkaCloudEventsProtocol) Send(context.Context, binding.Message, ...binding.Transformer) error {
	return nil
}
