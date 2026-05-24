package model

import "time"

// AuditLog mirrors one row of the audit_logs table. Payload is stored as raw
// JSON so we don't lose information when new event shapes are introduced.
type AuditLog struct {
	ID          int64     `json:"id"`
	Topic       string    `json:"topic"`
	EventType   string    `json:"event_type"`
	RoutingKey  string    `json:"routing_key"`
	PerformedBy string    `json:"performed_by"`
	Payload     string    `json:"payload"`
	OccurredAt  time.Time `json:"occurred_at"`
	CreatedAt   time.Time `json:"created_at"`
}
