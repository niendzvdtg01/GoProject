-- audit_logs persists every event consumed from the team.activity and
-- asset.changes topics. Used by the audit consumer to build an audit trail.
CREATE TABLE IF NOT EXISTS audit_logs (
    id            BIGINT       NOT NULL AUTO_INCREMENT,
    topic         VARCHAR(64)  NOT NULL,
    event_type    VARCHAR(64)  NOT NULL,
    routing_key   VARCHAR(128) NOT NULL,
    performed_by  VARCHAR(64)  NOT NULL,
    payload       JSON         NOT NULL,
    occurred_at   DATETIME(3)  NOT NULL,
    created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_audit_topic (topic),
    KEY idx_audit_event_type (event_type),
    KEY idx_audit_performed_by (performed_by),
    KEY idx_audit_occurred_at (occurred_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
