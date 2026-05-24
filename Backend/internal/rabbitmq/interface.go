package rabbitmq

import "context"

// TopicBinding pairs a queue name with the routing-key patterns it should receive
// from a topic exchange (e.g. "team.#", "asset.folder.*").
type TopicBinding struct {
	Queue       string
	RoutingKeys []string
}

// RabbitMQService is the message-broker abstraction used by publishers and
// consumers. Topic-exchange methods support the event stream (team.activity,
// asset.changes); the legacy queue methods are kept for one-off worker patterns.
type RabbitMQService interface {
	// DeclareTopicExchange ensures a durable topic exchange exists. Idempotent.
	DeclareTopicExchange(exchange string) error

	// PublishTopic publishes a message to a topic exchange with the given routing key.
	PublishTopic(ctx context.Context, exchange, routingKey string, message any) error

	// ConsumeTopic binds a queue to one or more routing keys on the topic exchange
	// and dispatches messages to handler in a goroutine. Returns when consumer
	// registration succeeds; the goroutine exits when ctx is canceled.
	ConsumeTopic(ctx context.Context, exchange string, binding TopicBinding, handler func([]byte) error) error

	Close() error
}
