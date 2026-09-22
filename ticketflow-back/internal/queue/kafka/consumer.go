package kafka

import (
	"context"
	"log/slog"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Consumer is a minimal consumer-group wrapper over franz-go.
type Consumer struct {
	cl  *kgo.Client
	log *slog.Logger
}

// NewConsumer joins the group and subscribes to the topics.
func NewConsumer(brokers []string, group string, topics []string, log *slog.Logger) (*Consumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topics...),
		// commit offsets only AFTER successful handling (at-least-once)
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
	)
	if err != nil {
		return nil, err
	}
	return &Consumer{cl: cl, log: log}, nil
}

// Handler processes one record. Returning an error means:
// do not commit this offset — the record will be redelivered.
type Handler func(ctx context.Context, topic string, key string, value []byte) error

// Run blocks until ctx is canceled: poll loop with manual commits.
func (c *Consumer) Run(ctx context.Context, h Handler) error {
	for {
		fetches := c.cl.PollRecords(ctx, 10)
		if fetches.IsClientClosed() {
			c.log.Info("kafka consumer stopped")
			return nil
		}
		fetches.EachError(func(t string, part int32, err error) {
			c.log.Error("kafka fetch error", "topic", t, "partition", part, "err", err)
		})
		fetches.EachRecord(func(rec *kgo.Record) {
			if err := h(ctx, rec.Topic, string(rec.Key), rec.Value); err != nil {
				c.log.Error("handler failed, offset NOT committed",
					"topic", rec.Topic,
					"partition", rec.Partition,
					"offset", rec.Offset,
					"err", err)
				return
			}
			// success → commit this record's offset
			c.cl.CommitRecords(ctx, rec)
		})
	}
}

// Close disconnects.
func (c *Consumer) Close() {
	c.cl.Close()
}
