package queue

import "github.com/rabbitmq/amqp091-go"

const (
	ImportExchange   = "import.exchange"
	ImportQueue      = "import.users.queue"
	ImportRoutingKey = "import.users"
	// Dead letter exchange and queue for failed imports
	ImportDLX          = "import.dlx"
	ImportDLQueue      = "import.users.dlq"
	ImportDLRoutingKey = "import.users.dlq"
)

type RabbitMQ struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

func NewRabbitMQ(connStr string) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(connStr)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	rmq := &RabbitMQ{conn: conn, channel: ch}
	if err := rmq.setup(); err != nil {
		rmq.Close()
		return nil, err
	}

	return rmq, nil
}

func (r *RabbitMQ) setup() error {
	// Declare main exchange
	if err := r.channel.ExchangeDeclare(
		ImportExchange, "direct", true, false, false, false, nil,
	); err != nil {
		return err
	}

	// Declare dead letter exchange
	if err := r.channel.ExchangeDeclare(
		ImportDLX, "direct", true, false, false, false, nil,
	); err != nil {
		return err
	}

	// Declare main queue with DLX settings
	if _, err := r.channel.QueueDeclare(
		ImportQueue,
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-dead-letter-exchange":    ImportDLX,
			"x-dead-letter-routing-key": ImportDLRoutingKey,
			"x-message-ttl":             int32(60000), // 1 minute TTL for messages
		},
	); err != nil {
		return err
	}

	if err := r.channel.QueueBind(
		ImportQueue, ImportRoutingKey, ImportExchange, false, nil,
	); err != nil {
		return err
	}

	// Declare dead letter queue
	if _, err := r.channel.QueueDeclare(
		ImportDLQueue,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if err := r.channel.QueueBind(
		ImportDLQueue, ImportDLRoutingKey, ImportDLX, false, nil,
	); err != nil {
		return err
	}

	return nil
}

func (r *RabbitMQ) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
