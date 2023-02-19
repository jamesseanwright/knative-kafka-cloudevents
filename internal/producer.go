package internal

import (
	"context"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const writeInterval = 1 * time.Second

type cloudEventSender interface {
	Send(context.Context, event.Event) protocol.Result
}

type Producer struct {
	sender cloudEventSender
	logger stdLogger
}

func NewProducer(sender cloudEventSender, logger stdLogger) Producer {
	return Producer{sender, logger}
}

func (p Producer) Run(ctx context.Context) error {
	ticker := time.NewTicker(writeInterval)
	done := ctx.Done()
	g, ctx := errgroup.WithContext(ctx)

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

	if err := p.sender.Send(ctx, event); err != nil {
		return fmt.Errorf("cloudevent send: %w", err)
	}

	return nil
}
