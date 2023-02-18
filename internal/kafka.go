package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/types"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/protocol"
)

const (
	cloudEventPrefix = "ce_"
	kibibyte         = 1024
	mebibyte         = 1024 * kibibyte
	readMinBatchSize = 10 * kibibyte
	// readMaxBatchSize = 10 * mebibyte
)

type message struct {
	value       []byte
	headers     map[string]string
	contentType string
	version     spec.Version
}

func newMessage(msg kafka.Message) *message {
	var headers map[string]string
	var contentType string
	var version spec.Version
	spc := spec.WithPrefix(cloudEventPrefix)

	for _, v := range msg.Headers {
		if v.Key == spc.PrefixedSpecVersionName() {
			version = spc.Version(string(v.Value))
		} else if strings.ToLower(v.Key) == "content-type" {
			contentType = string(v.Value)
		} else {
			headers[v.Key] = string(v.Value)
		}
	}

	return &message{
		value:       msg.Value,
		headers:     headers,
		contentType: contentType,
		version:     version,
	}
}

func (m message) ReadEncoding() binding.Encoding {
	return binding.EncodingBinary
}

func (m message) ReadStructured(context.Context, binding.StructuredWriter) error {
	return binding.ErrNotStructured
}

func (m message) ReadBinary(ctx context.Context, writer binding.BinaryWriter) error {
	for k, v := range m.headers {
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

	if m.value != nil {
		if err := writer.SetData(bytes.NewBuffer(m.value)); err != nil {
			return fmt.Errorf("kafka cloudevent parse data: %w", err)
		}
	}

	return nil
}

func (m message) Finish(error) error {
	return nil
}

type kafkaMessageWriter kafka.Message

func (m *kafkaMessageWriter) Start(context.Context) error {
	return nil
}

func (m *kafkaMessageWriter) End(context.Context) error {
	return nil
}

func (m *kafkaMessageWriter) SetData(reader io.Reader) error {
	data, err := io.ReadAll(reader)

	if err != nil {
		return fmt.Errorf("kafka message writer read data: %w", err)
	}

	m.Value = data

	return nil
}

func (m *kafkaMessageWriter) SetAttribute(attribute spec.Attribute, value any) error {
	prefix := cloudEventPrefix

	if attribute.Kind() == spec.DataContentType {
		prefix = ""
	}

	return m.setHeader(prefix, attribute.Name(), value)
}

func (m *kafkaMessageWriter) SetExtension(name string, value any) error {
	return m.setHeader("", name, value)
}

func (m *kafkaMessageWriter) setHeader(prefix, name string, value any) error {
	s, err := types.Format(value)

	if err != nil {
		return fmt.Errorf("kafka message writer format header value: %w", err)
	}

	m.Headers = append(m.Headers, protocol.Header{
		Key:   prefix + name,
		Value: []byte(s),
	})

	return nil
}

type kafkaConn interface {
	ReadMessage(int) (kafka.Message, error)
	WriteMessages(...kafka.Message) (int, error)
}

type KafkaCloudEventsProtocol struct {
	conn kafkaConn
}

func NewKafkaCloudEventsProtocol(conn kafkaConn) KafkaCloudEventsProtocol {
	return KafkaCloudEventsProtocol{conn}
}

// TODO: investigate batching support (Requestor?)
func (p KafkaCloudEventsProtocol) Receive(context.Context) (binding.Message, error) {
	msg, err := p.conn.ReadMessage(readMinBatchSize)

	if err != nil {
		return nil, fmt.Errorf("receive kafka cloudevent: %w", err)
	}

	return newMessage(msg), nil
}

func (p KafkaCloudEventsProtocol) Send(_ context.Context, message binding.Message, transformers ...binding.Transformer) error {
	// TODO: WriteProducerMessage here (https://github.com/cloudevents/sdk-go/blob/82f2b61ecde41fd0577969cc58a7c2a18eeda250/protocol/kafka_sarama/v2/sender.go#L70)
	// TODO: general attribution
	return nil
}
