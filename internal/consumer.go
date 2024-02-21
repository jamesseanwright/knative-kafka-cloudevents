package internal

import (
	"context"
	"errors"
	"io"
)

type Consumer struct {
	eventReader CloudEventsReader
	logger      *Logger
}

func NewConsumer(eventReader CloudEventsReader, logger *Logger) Consumer {
	return Consumer{eventReader, logger}
}

func (p Consumer) Run(ctx context.Context) {
	// We're scaling consumers with Knative,
	// so it's perfectly fine to block here

	for {
		event, err := p.eventReader.Read(ctx)

		if err != nil && !errors.Is(err, io.EOF) {
			p.logger.Error("read batch: %w", err)
		}

		p.logger.Info(event)
	}
}
