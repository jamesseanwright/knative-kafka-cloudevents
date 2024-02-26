package internal

import (
	"context"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const writeInterval = 100 * time.Millisecond

type Producer struct {
	eventSender CloudEventsSender
	logger      *Logger
}

func NewProducer(eventSender CloudEventsSender, logger *Logger) Producer {
	return Producer{eventSender, logger}
}

func (p Producer) Run(ctx context.Context) error {
	ticker := time.NewTicker(writeInterval)
	done := ctx.Done()
	g, ctx := errgroup.WithContext(ctx)

	// TODO: we shouldn't need to execute this 
	// in a child goroutine; we should just
	// be able to block the main goroutine!
	g.Go(func() error {
		for {
			select {
			case <-done:
				return nil

			case <-ticker.C:
				if err := p.sendMessage(ctx); err != nil {
					return err
				}
			}
		}
	})

	return g.Wait()
}

func (p Producer) sendMessage(ctx context.Context) error {
	id, err := uuid.NewRandom()

	if err != nil {
		return fmt.Errorf("cloudevent id generation: %w", err)
	}

	p.logger.Info("creating new event with ID", id)

	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetID(id.String())
	event.SetSource("knative-kafka-cloudevents/producer")
	event.SetType("com.example.testevent.produced")

	if err := event.SetData("text/plain", "test event!"); err != nil {
		return fmt.Errorf("cloudevent set data: %w", err)
	}

	if err := p.eventSender.Send(event); err != nil {
		return fmt.Errorf("cloudevent send: %w", err)
	}

	return nil
}
