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

// TODO: general attribution

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
	var contentType string
	var version spec.Version
	headers := map[string]string{}
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

func (m message) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	attr := m.version.AttributeFromKind(k)

	if attr != nil {
		return attr, string(m.headers[attr.PrefixedName()])
	}

	return nil, nil
}

func (m message) GetExtension(name string) interface{} {
	return string(m.headers[cloudEventPrefix+name])
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

func (p KafkaCloudEventsProtocol) Send(ctx context.Context, message binding.Message, transformers ...binding.Transformer) error {
	kafkaMsg := kafka.Message{}

	enc, err := binding.Write(ctx, message, nil, (*kafkaMessageWriter)(&kafkaMsg))

	if err != nil {
		return fmt.Errorf("marshal cloudevent: %w", err)
	}

	if enc != binding.EncodingBinary {
		return fmt.Errorf("unexpected cloudevent encoding (want binary, got %s)", enc.String())
	}

	if _, err := p.conn.WriteMessages(kafkaMsg); err != nil {
		return fmt.Errorf("write kafka message: %w", err)
	}

	return nil
}

// verifying interface conformity
var _ binding.BinaryWriter = (*kafkaMessageWriter)(nil)
