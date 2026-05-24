package consumer

import (
	"context"

	"backend/internal/rabbitmq"
	"backend/package/event"

	"github.com/rs/zerolog"
)

// Routing keys this consumer cares about — user-facing events worth notifying
// someone over. Keep this narrow; broad bindings belong to the audit consumer.
var notificationRoutingKeys = []string{
	"team.member.added",
	"team.member.removed",
	"asset.folder.shared",
	"asset.note.shared",
}

// NewNotificationConsumer logs interesting events as a placeholder for a real
// notification fan-out (email, push, in-app). The two topic exchanges have
// separate queues so each can be scaled / acked independently.
func NewNotificationConsumer(broker rabbitmq.RabbitMQService, logger *zerolog.Logger) []*Consumer {
	handler := notificationHandler(logger)
	return []*Consumer{
		New(broker, logger, "notify/team", event.TopicTeamActivity,
			rabbitmq.TopicBinding{Queue: "notifications.team", RoutingKeys: filterKeys(notificationRoutingKeys, "team.")},
			handler,
		),
		New(broker, logger, "notify/asset", event.TopicAssetChanges,
			rabbitmq.TopicBinding{Queue: "notifications.asset", RoutingKeys: filterKeys(notificationRoutingKeys, "asset.")},
			handler,
		),
	}
}

func notificationHandler(logger *zerolog.Logger) func(context.Context, event.Event) error {
	return func(_ context.Context, evt event.Event) error {
		logger.Info().
			Str("topic", evt.Topic).
			Str("event_type", evt.EventType).
			Str("performed_by", evt.PerformedBy).
			Interface("payload", evt.Payload).
			Msg("notification: dispatch")
		// TODO: swap log for real notification fan-out (email/push/in-app).
		return nil
	}
}

func filterKeys(keys []string, prefix string) []string {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			out = append(out, k)
		}
	}
	return out
}
