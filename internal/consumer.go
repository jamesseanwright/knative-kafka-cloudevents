package internal

import (
	"context"
	"fmt"

	v2 "github.com/cloudevents/sdk-go/v2"
)

type cloudEventReceiver interface {
	StartReceiver(ctx context.Context, receiverFunc any) error
}

type Consumer struct {
	receiver cloudEventReceiver
	logger   stdLogger
}

func NewConsumer(receiver cloudEventReceiver, logger stdLogger) Consumer {
	return Consumer{receiver, logger}
}

func (p Consumer) Run(ctx context.Context) error {
	// We're scaling consumers with Knative,
	// so it's perfectly fine to block here
	if err := p.receiver.StartReceiver(ctx, func(event v2.Event) {
		p.logger.Info(event)
	}); err != nil {
		return fmt.Errorf("cloudevents receiver: %w", err)
	}

	return nil
}
