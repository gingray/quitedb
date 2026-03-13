package kafka

import (
	"context"
	"encoding/json"

	"github.com/gingray/quitedb/pkg/app"
	"github.com/gingray/quitedb/pkg/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Client interface {
	Produce(ctx context.Context, topic, key string, value any) error
}

type Producer struct {
	client *kgo.Client
	logger config.Logger
}

func NewProducer(app *app.App) *Producer {
	return &Producer{
		client: app.Kafka,
		logger: app.Logger,
	}
}

func (p *Producer) Produce(ctx context.Context, topic, key string, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	record := kgo.Record{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
	}
	return p.client.ProduceSync(ctx, &record).FirstErr()
}
