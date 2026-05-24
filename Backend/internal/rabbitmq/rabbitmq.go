package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

const (
	exchangeKind     = "topic"
	contentTypeJSON  = "application/json"
	deliveryModePers = amqp.Persistent
)

type rabbitMQService struct {
	conn   *amqp.Connection
	chanel *amqp.Channel
	logger *zerolog.Logger
}

func NewRabbitMQService(amqpURL string, logger *zerolog.Logger) (RabbitMQService, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		logger.Error().Err(err).Msg("fail to connect rabbitmq")
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		logger.Error().Err(err).Msg("fail to open channel")
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	return &rabbitMQService{
		conn:   conn,
		chanel: ch,
		logger: logger,
	}, nil
}

func (r *rabbitMQService) DeclareTopicExchange(exchange string) error {
	if err := r.chanel.ExchangeDeclare(exchange, exchangeKind, true, false, false, false, nil); err != nil {
		r.logger.Error().Err(err).Str("exchange", exchange).Msg("declare exchange")
		return fmt.Errorf("declare exchange %s: %w", exchange, err)
	}
	return nil
}

func (r *rabbitMQService) PublishTopic(ctx context.Context, exchange, routingKey string, message any) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	err = r.chanel.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  contentTypeJSON,
		DeliveryMode: deliveryModePers,
		Body:         body,
	})
	if err != nil {
		r.logger.Error().Err(err).
			Str("exchange", exchange).Str("routing_key", routingKey).
			Msg("publish message")
		return fmt.Errorf("publish to %s/%s: %w", exchange, routingKey, err)
	}
	return nil
}

func (r *rabbitMQService) ConsumeTopic(ctx context.Context, exchange string, binding TopicBinding, handler func([]byte) error) error {
	if len(binding.RoutingKeys) == 0 {
		return fmt.Errorf("binding %s: at least one routing key required", binding.Queue)
	}

	queue, err := r.chanel.QueueDeclare(binding.Queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue %s: %w", binding.Queue, err)
	}

	for _, key := range binding.RoutingKeys {
		if err := r.chanel.QueueBind(queue.Name, key, exchange, false, nil); err != nil {
			return fmt.Errorf("bind queue %s to %s/%s: %w", queue.Name, exchange, key, err)
		}
	}

	// Manual ack so we can nack-and-requeue on transient handler failures.
	msgs, err := r.chanel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume queue %s: %w", queue.Name, err)
	}

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					r.logger.Info().Str("queue", queue.Name).Msg("consumer channel closed")
					return
				}
				if err := handler(msg.Body); err != nil {
					r.logger.Error().Err(err).Str("queue", queue.Name).Msg("handler failed; requeueing")
					_ = msg.Nack(false, true)
					continue
				}
				_ = msg.Ack(false)
			case <-ctx.Done():
				r.logger.Info().Str("queue", queue.Name).Msg("consumer context done")
				return
			}
		}
	}()
	return nil
}

func (r *rabbitMQService) Close() error {
	if r.chanel != nil {
		if err := r.chanel.Close(); err != nil {
			r.logger.Error().Err(err).Msg("close channel")
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			r.logger.Error().Err(err).Msg("close connection")
			return err
		}
	}
	return nil
}
