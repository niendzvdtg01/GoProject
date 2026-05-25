-- Full initial schema for miniproject_database.
-- Run once on a fresh database; migrations in migration_*.sql are already
-- folded in (role column, owner_id on teams, audit_logs table).

CREATE TABLE IF NOT EXISTS users (
    user_id       VARCHAR(36)  NOT NULL,
    username      VARCHAR(100) NOT NULL,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'member',
    created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id),
    UNIQUE KEY uniq_users_email    (email),
    UNIQUE KEY uniq_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS teams (
    team_id    INT          NOT NULL AUTO_INCREMENT,
    team_name  VARCHAR(255) NOT NULL,
    owner_id   VARCHAR(36)  NOT NULL DEFAULT '',
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (team_id),
    UNIQUE KEY uniq_teams_name (team_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS team_members (
    team_id   INT         NOT NULL,
    user_id   VARCHAR(36) NOT NULL,
    role      VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (team_id, user_id),
    KEY idx_team_members_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS folders (
    folder_id  INT          NOT NULL AUTO_INCREMENT,
    owner_id   VARCHAR(36)  NOT NULL,
    name       VARCHAR(255) NOT NULL,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (folder_id),
    KEY idx_folders_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS notes (
    note_id    INT          NOT NULL AUTO_INCREMENT,
    folder_id  INT          NOT NULL,
    owner_id   VARCHAR(36)  NOT NULL,
    title      VARCHAR(255) NOT NULL,
    content    TEXT         NOT NULL,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (note_id),
    KEY idx_notes_folder (folder_id),
    KEY idx_notes_owner  (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS permissions (
    id              INT         NOT NULL AUTO_INCREMENT,
    asset_type      VARCHAR(50) NOT NULL,
    asset_id        INT         NOT NULL,
    user_id         VARCHAR(36) NOT NULL,
    permission_type VARCHAR(50) NOT NULL,
    granted_by      VARCHAR(36) NOT NULL,
    created_at      DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uniq_permissions (asset_type, asset_id, user_id),
    KEY idx_permissions_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS import_tasks (
    task_id        BIGINT       NOT NULL AUTO_INCREMENT,
    created_by     VARCHAR(36)  NOT NULL,
    file_name      VARCHAR(255) NOT NULL,
    status         VARCHAR(50)  NOT NULL DEFAULT 'pending',
    total_rows     INT          NOT NULL DEFAULT 0,
    processed_rows INT          NOT NULL DEFAULT 0,
    succeeded      INT          NOT NULL DEFAULT 0,
    failed         INT          NOT NULL DEFAULT 0,
    error_log      JSON,
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at     DATETIME,
    completed_at   DATETIME,
    PRIMARY KEY (task_id),
    KEY idx_import_tasks_user   (created_by),
    KEY idx_import_tasks_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS audit_logs (
    id           BIGINT       NOT NULL AUTO_INCREMENT,
    topic        VARCHAR(64)  NOT NULL,
    event_type   VARCHAR(64)  NOT NULL,
    routing_key  VARCHAR(128) NOT NULL,
    performed_by VARCHAR(64)  NOT NULL,
    payload      JSON         NOT NULL,
    occurred_at  DATETIME(3)  NOT NULL,
    created_at   DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    KEY idx_audit_topic        (topic),
    KEY idx_audit_event_type   (event_type),
    KEY idx_audit_performed_by (performed_by),
    KEY idx_audit_occurred_at  (occurred_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
