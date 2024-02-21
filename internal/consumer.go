package internal

import (
	"context"
	"errors"
	"io"
)

type Consumer struct {
	eventReader CloudEventsBatchReader
	logger      *Logger
}

func NewConsumer(eventReader CloudEventsBatchReader, logger *Logger) Consumer {
	return Consumer{eventReader, logger}
}

func (p Consumer) Run(ctx context.Context) {
	// We're scaling consumers with Knative,
	// so it's perfectly fine to block here

	for {
		batch, err := p.eventReader.ReadBatch()

		if err != nil && !errors.Is(err, io.EOF) {
			p.logger.Error("read batch: %w", err)
		}

		for _, v := range batch {
			p.logger.Info(v)
		}
	}
}
