package internal

import (
	"context"
	"fmt"

	v2 "github.com/cloudevents/sdk-go/v2"
)

// TODO: move into Kafka transport
// const (
// 	kibibyte         = 1024
// 	mebibyte         = 1024 * kibibyte
// 	readMinBatchSize = 10 * kibibyte
// 	readMaxBatchSize = 10 * mebibyte
// )

// Can't type more specifically as the CloudEvents
// SDK explicitly types these func params as interface{}
type receiverFunc any

type cloudEventReceiver interface {
	StartReceiver(context.Context, receiverFunc) error
}

type logger interface {
	Info(args ...any)
}

type Consumer struct {
	receiver cloudEventReceiver
	logger   logger
}

func NewConsumer(receiver cloudEventReceiver, logger logger) Consumer {
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
