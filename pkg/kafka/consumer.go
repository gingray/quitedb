package kafka

import (
	"context"
	"fmt"

	"github.com/gingray/quitedb/pkg/app"
	"github.com/gingray/quitedb/pkg/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Consume interface {
	Topic() string
	Consume(ctx context.Context, key string, value []byte) error
}
type Consumer struct {
	client   *kgo.Client
	logger   config.Logger
	handlers map[string]Consume
}

func NewConsumer(app *app.App, consumers ...Consume) *Consumer {
	handlers := buildHandlers(consumers)
	return &Consumer{client: app.Kafka, logger: app.Logger, handlers: handlers}
}

func (k *Consumer) Name() string {
	return "kafka"
}

func (k *Consumer) Ready(ctx context.Context) error {
	return nil
}

func (k *Consumer) Run(ctx context.Context) error {
	for {
		fetches := k.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return nil
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			return fmt.Errorf("fetch error: %v", errs)
		}

		fetches.EachRecord(func(r *kgo.Record) {
			handler, ok := k.handlers[r.Topic]
			key := string(r.Key)
			if !ok {
				k.logger.Warn("no consumer handler for topic", "topic", r.Topic, "key", key)
				return
			}
			err := handler.Consume(ctx, key, r.Value)
			if err != nil {
				k.logger.Error("consumer error", "topic", r.Topic, "key", key, "error", err)
			}
		})

		// commit offsets if using consumer group
		if err := k.client.CommitUncommittedOffsets(ctx); err != nil {
			return fmt.Errorf("failed to commit offsets: %w", err)
		}
	}
}

func (k *Consumer) Shutdown(ctx context.Context) error {
	k.client.Close()
	return nil
}

func buildHandlers(consumers []Consume) map[string]Consume {
	consumerMap := make(map[string]Consume)
	for _, consumer := range consumers {
		consumerMap[consumer.Topic()] = consumer
	}
	return consumerMap
}
