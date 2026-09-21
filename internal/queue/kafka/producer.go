// Package kafka contains the Kafka producer and consumer wiring.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/upuzipu/ticketflow/internal/domain"
)

// Producer publishes domain events to Kafka topics.
type Producer struct {
	cl  *kgo.Client
	log *slog.Logger
}

// NewProducer connects to the broker and ensures the topics exist.
func NewProducer(ctx context.Context, brokers []string, log *slog.Logger) (*Producer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.AllowAutoTopicCreation(),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	if err := cl.Ping(ctx); err != nil {
		cl.Close()
		return nil, fmt.Errorf("ping kafka: %w", err)
	}
	return &Producer{cl: cl, log: log}, nil
}

// Publish writes the event to its topic with the aggregate id as partition key.
// Synchronous: returns only after the broker acks. Used by the relay,
// not by the request path.
func (p *Producer) Publish(ctx context.Context, e domain.DomainEvent) error {
	payload, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", e.EventName(), err)
	}

	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rec := &kgo.Record{
		Topic: topicFor(e.EventName()),
		Key:   []byte(e.AggregateID()),
		Value: payload,
	}

	if err := p.cl.ProduceSync(pubCtx, rec).FirstErr(); err != nil {
		return fmt.Errorf("produce %s: %w", e.EventName(), err)
	}

	p.log.Info("event published",
		"event", e.EventName(),
		"key", e.AggregateID(),
		"topic", rec.Topic,
	)
	return nil
}

// Close flushes and disconnects.
func (p *Producer) Close() {
	p.cl.Close()
}

// topicFor maps event names to topics. Keep in one place — consumers
// subscribe by these constants.
func topicFor(eventName string) string {
	switch eventName {
	case domain.EventOrderPaid, domain.EventOrderCreated, domain.EventOrderFailed:
		return "order-events"
	case domain.EventTicketsIssued, domain.EventTicketsReleased:
		return "ticket-events"
	default:
		return "misc-events"
	}
}
