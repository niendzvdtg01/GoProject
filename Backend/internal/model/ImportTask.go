package model

import (
	"database/sql"
	"time"
)

type ImportTask struct {
	TaskID        int64          `json:"task_id" db:"task_id"`
	CreatedBy     string         `json:"created_by" db:"created_by"`
	FileName      string         `json:"file_name" db:"file_name"`
	Status        string         `json:"status" db:"status"`
	TotalRows     int            `json:"total_rows" db:"total_rows"`
	ProcessedRows int            `json:"processed_rows" db:"processed_rows"`
	Succeeded     int            `json:"succeeded" db:"succeeded"`
	Failed        int            `json:"failed" db:"failed"`
	ErrorLog      sql.NullString `json:"error_log" db:"error_log"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
	StartedAt     sql.NullTime   `json:"started_at" db:"started_at"`
	CompletedAt   sql.NullTime   `json:"completed_at" db:"completed_at"`
}
