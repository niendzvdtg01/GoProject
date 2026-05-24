package consumer

import (
	"context"

	"backend/internal/rabbitmq"
	"backend/package/event"

	"github.com/rs/zerolog"
)

// Consumer registers one queue binding against a topic exchange and runs its
// handler until ctx is canceled. Concrete consumers (audit, notification) are
// built by passing a handler to Register.
type Consumer struct {
	Name     string
	Topic    string
	Binding  rabbitmq.TopicBinding
	Handler  func(ctx context.Context, evt event.Event) error
	broker   rabbitmq.RabbitMQService
	logger   *zerolog.Logger
}

func New(broker rabbitmq.RabbitMQService, logger *zerolog.Logger, name, topic string, binding rabbitmq.TopicBinding, handler func(context.Context, event.Event) error) *Consumer {
	return &Consumer{
		Name:    name,
		Topic:   topic,
		Binding: binding,
		Handler: handler,
		broker:  broker,
		logger:  logger,
	}
}

// Start declares the topic exchange, binds the queue, and launches the consume
// goroutine. The goroutine exits when ctx is canceled or the channel closes.
func (c *Consumer) Start(ctx context.Context) error {
	if err := c.broker.DeclareTopicExchange(c.Topic); err != nil {
		return err
	}

	c.logger.Info().
		Str("consumer", c.Name).
		Str("topic", c.Topic).
		Str("queue", c.Binding.Queue).
		Strs("routing_keys", c.Binding.RoutingKeys).
		Msg("starting consumer")

	return c.broker.ConsumeTopic(ctx, c.Topic, c.Binding, func(body []byte) error {
		evt, err := decodeEvent(body)
		if err != nil {
			// Malformed payload — log and ack (return nil) so it doesn't loop forever.
			c.logger.Error().Err(err).Str("consumer", c.Name).Msg("drop malformed event")
			return nil
		}
		return c.Handler(ctx, evt)
	})
}
