package repository

import (
	"context"
	"database/sql"
	"time"
)

type AuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Insert persists one audit row. Payload is the raw JSON body emitted by the
// publisher so we never lose fields when event schemas evolve.
func (r *AuditLogRepository) Insert(ctx context.Context, topic, eventType, routingKey, performedBy, payload string, occurredAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_logs (topic, event_type, routing_key, performed_by, payload, occurred_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		topic, eventType, routingKey, performedBy, payload, occurredAt,
	)
	return err
}
