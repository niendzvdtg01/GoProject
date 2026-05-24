package consumer

import (
	"context"
	"encoding/json"

	"backend/internal/rabbitmq"
	"backend/internal/repository"
	"backend/package/event"

	"github.com/rs/zerolog"
)

// audit_queue captures every event from both topics. Bindings use the topic-
// exchange wildcard "#" so any new event type is recorded automatically.
const auditQueueName = "audit.logs"

// NewAuditConsumer builds two Consumers (one per topic) that persist every
// event to audit_logs. The repository write owns the only durability decision:
// if the DB write fails the handler returns an error and the message is
// requeued for retry.
func NewAuditConsumer(broker rabbitmq.RabbitMQService, repo *repository.AuditLogRepository, logger *zerolog.Logger) []*Consumer {
	handler := auditHandler(repo, logger)
	return []*Consumer{
		New(broker, logger, "audit/team", event.TopicTeamActivity,
			rabbitmq.TopicBinding{Queue: auditQueueName + ".team", RoutingKeys: []string{"team.#"}},
			handler,
		),
		New(broker, logger, "audit/asset", event.TopicAssetChanges,
			rabbitmq.TopicBinding{Queue: auditQueueName + ".asset", RoutingKeys: []string{"asset.#"}},
			handler,
		),
	}
}

func auditHandler(repo *repository.AuditLogRepository, logger *zerolog.Logger) func(context.Context, event.Event) error {
	return func(ctx context.Context, evt event.Event) error {
		payload, err := json.Marshal(evt.Payload)
		if err != nil {
			logger.Error().Err(err).Str("event_type", evt.EventType).Msg("audit: re-marshal payload")
			return nil
		}

		if err := repo.Insert(ctx, evt.Topic, evt.EventType, evt.RoutingKey, evt.PerformedBy, string(payload), evt.Timestamp); err != nil {
			logger.Error().Err(err).
				Str("event_type", evt.EventType).
				Str("performed_by", evt.PerformedBy).
				Msg("audit: persist event")
			return err
		}
		return nil
	}
}
