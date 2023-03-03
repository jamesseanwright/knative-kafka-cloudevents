package internal

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type Consumer struct {
	eventReader CloudEventsBatchReader
	logger      *Logger
}

func NewConsumer(eventReader CloudEventsBatchReader, logger *Logger) Consumer {
	return Consumer{eventReader, logger}
}

func (p Consumer) Run(ctx context.Context) {
	events := make(chan cloudevents.Event)
	errs := make(chan error)

	p.eventReader.ReadBatch(events, errs)

	// We're scaling consumers with Knative,
	// so it's perfectly fine to block here
	for {
		select {
		case evt := <-events:
			p.logger.Info(evt)
		case err := <-errs:
			p.logger.Error(err)
		}
	}
}
