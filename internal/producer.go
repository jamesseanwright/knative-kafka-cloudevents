package internal

import (
	"context"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
)

const interval = 1 * time.Second

type kafkaWriter interface {
	WriteMessages(...kafka.Message) (int, error)
}

type Producer struct {
	writer kafkaWriter
}

func NewProducer(writer kafkaWriter) Producer {
	return Producer{writer}
}

func (p Producer) Run(ctx context.Context) error {
	ticker := time.NewTicker(interval)
	done := ctx.Done()
	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		for {
			select {
			case <-done:
				return nil

			case <-ticker.C:
				if err := p.writeMessage(); err != nil {
					return err
				}
			}
		}
	})

	return g.Wait()
}

func (p Producer) writeMessage() error {
	id, err := uuid.NewRandom()

	if err != nil {
		return fmt.Errorf("cloudevent id generation: %w", err)
	}

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetID(id.String())
	event.SetSource("knative-kafka-cloudevents/producer")
	event.SetType("com.example.testevent.produced")

	if err := event.SetData("text/plain", "test event!"); err != nil {
		return fmt.Errorf("cloudevent set data: %w", err)
	}

	n, err := p.writer.WriteMessages(kafka.Message{
		Value: event.Data(),
	})

	if err != nil {
		return fmt.Errorf("kafka write: %w", err)
	}

	if n != 1 {
		return fmt.Errorf("kafka write: incorrect number of messages published (want: 1, got: %d)", n)
	}

	return nil
}
