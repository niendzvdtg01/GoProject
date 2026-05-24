package event

import (
	"context"
	"time"

	"backend/internal/rabbitmq"

	"github.com/rs/zerolog"
)

// Publisher emits domain events to the message broker. All methods are
// fire-and-forget: publishing must never break a business operation, so errors
// are logged but not surfaced. Use NewNoopPublisher in tests / when the broker
// is unavailable.
type Publisher interface {
	PublishTeamEvent(ctx context.Context, eventType, performedBy string, payload map[string]any)
	PublishAssetEvent(ctx context.Context, eventType, performedBy string, payload map[string]any)
}

type rabbitPublisher struct {
	broker rabbitmq.RabbitMQService
	logger *zerolog.Logger
}

// NewPublisher wires a Publisher backed by RabbitMQ. The caller is responsible
// for ensuring the broker connection stays alive; this publisher does not own it.
// Both topic exchanges are declared up-front so publishes never fail on a cold start.
func NewPublisher(broker rabbitmq.RabbitMQService, logger *zerolog.Logger) (Publisher, error) {
	for _, topic := range []string{TopicTeamActivity, TopicAssetChanges} {
		if err := broker.DeclareTopicExchange(topic); err != nil {
			return nil, err
		}
	}
	return &rabbitPublisher{broker: broker, logger: logger}, nil
}

func (p *rabbitPublisher) PublishTeamEvent(ctx context.Context, eventType, performedBy string, payload map[string]any) {
	p.emit(ctx, TopicTeamActivity, eventType, performedBy, payload)
}

func (p *rabbitPublisher) PublishAssetEvent(ctx context.Context, eventType, performedBy string, payload map[string]any) {
	p.emit(ctx, TopicAssetChanges, eventType, performedBy, payload)
}

func (p *rabbitPublisher) emit(ctx context.Context, topic, eventType, performedBy string, payload map[string]any) {
	routingKey := RoutingKey(topic, eventType)
	evt := Event{
		Topic:       topic,
		EventType:   eventType,
		RoutingKey:  routingKey,
		PerformedBy: performedBy,
		Timestamp:   time.Now().UTC(),
		Payload:     payload,
	}
	if err := p.broker.PublishTopic(ctx, topic, routingKey, evt); err != nil {
		// Best-effort: the business write already succeeded; the broker outage
		// must not be visible to the API caller.
		p.logger.Error().Err(err).
			Str("topic", topic).
			Str("event_type", eventType).
			Msg("event publish failed")
	}
}
