package repository

import (
	"backend/internal/model"
	"context"
	"database/sql"
	"time"
)

type ImportTaskRepository struct {
	db *sql.DB
}

func NewImportTaskRepository(db *sql.DB) *ImportTaskRepository {
	return &ImportTaskRepository{db: db}
}

func (r *ImportTaskRepository) CreateTask(ctx context.Context, createdBy, fileName string) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO import_tasks (created_by, file_name, status, total_rows, processed_rows, succeeded, failed)
		 VALUES (?, ?, 'pending', 0, 0, 0, 0)`,
		createdBy, fileName,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *ImportTaskRepository) MarkProcessing(ctx context.Context, taskID int64, totalRows int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE import_tasks SET status = 'processing', total_rows = ?, started_at = ? WHERE task_id = ?`,
		totalRows, time.Now(), taskID,
	)
	return err
}

func (r *ImportTaskRepository) MarkCompleted(ctx context.Context, taskID int64, succeeded, failed int, errorLogJSON string) error {
	processedRows := succeeded + failed
	_, err := r.db.ExecContext(ctx,
		`UPDATE import_tasks
		 SET status = 'completed', succeeded = ?, failed = ?, processed_rows = ?,
		     error_log = ?, completed_at = ?
		 WHERE task_id = ?`,
		succeeded, failed, processedRows, errorLogJSON, time.Now(), taskID,
	)
	return err
}

func (r *ImportTaskRepository) MarkFailed(ctx context.Context, taskID int64, reason string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE import_tasks SET status = 'failed', error_log = JSON_ARRAY(?), completed_at = ? WHERE task_id = ?`,
		reason, time.Now(), taskID,
	)
	return err
}

func (r *ImportTaskRepository) UpdateProgress(ctx context.Context, taskID int64, processedRows int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE import_tasks SET processed_rows = ? WHERE task_id = ?`,
		processedRows, taskID,
	)
	return err
}

func (r *ImportTaskRepository) GetTask(ctx context.Context, taskID int64) (*model.ImportTask, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT task_id, created_by, file_name, status, total_rows, processed_rows, succeeded, failed, error_log, created_at, started_at, completed_at
		 FROM import_tasks WHERE task_id = ?`, taskID,
	)

	var task model.ImportTask
	var errorLog sql.NullString
	err := row.Scan(&task.TaskID, &task.CreatedBy, &task.FileName, &task.Status,
		&task.TotalRows, &task.ProcessedRows, &task.Succeeded, &task.Failed,
		&errorLog, &task.CreatedAt, &task.StartedAt, &task.CompletedAt)
	if err != nil {
		return nil, err
	}
	task.ErrorLog = errorLog.String
	return &task, nil
}
