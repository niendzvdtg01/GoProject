package respository

import (
	"backend/internal/model"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

var ErrNoteNotFound = errors.New("note not found")

type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// CreateNote creates a new note in the database
func (nr *NoteRepository) CreateNote(folderID, title, content string) (int, error) {
	const query = `
	INSERT INTO notes (folder_id, title, content)
	VALUES (?, ?, ?);`

	result, err := nr.db.Exec(query, folderID, title, content)
	if err != nil {
		return 0, err
	}

	noteID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(noteID), nil
}

func (nr *NoteRepository) GetNoteByID(noteID string) (model.Notes, error) {
	const query = `
	SELECT id, folder_id, title, content, created_at, updated_at
	FROM notes
	WHERE id = ?;`

	var note model.Notes
	var id int
	var folder int
	err := nr.db.QueryRow(query, noteID).Scan(
		&id,
		&folder,
		&note.Title,
		&note.Content,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Notes{}, ErrNoteNotFound
		}
		return model.Notes{}, fmt.Errorf("get note by id: %w", err)
	}
	note.ID = strconv.Itoa(id)
	note.FolderID = strconv.Itoa(folder)
	return note, nil
}

func (nr *NoteRepository) ListNotesByFolder(folderID string) ([]model.Notes, error) {
	const query = `
	SELECT id, folder_id, title, content, created_at, updated_at
	FROM notes
	WHERE folder_id = ?
	ORDER BY created_at DESC;`

	rows, err := nr.db.Query(query, folderID)
	if err != nil {
		return nil, fmt.Errorf("list notes by folder: %w", err)
	}
	defer rows.Close()

	notes := make([]model.Notes, 0)
	for rows.Next() {
		var note model.Notes
		var id int
		var folder int
		if err := rows.Scan(&id, &folder, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		note.ID = strconv.Itoa(id)
		note.FolderID = strconv.Itoa(folder)
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notes: %w", err)
	}

	return notes, nil
}

func (nr *NoteRepository) UpdateNote(noteID string, title, content string) (model.Notes, error) {
	const query = `
	UPDATE notes
	SET title = ?, content = ?, updated_at = NOW()
	WHERE id = ?;`

	_, err := nr.db.Exec(query, title, content, noteID)
	if err != nil {
		return model.Notes{}, fmt.Errorf("update note: %w", err)
	}

	return nr.GetNoteByID(noteID)
}

func (nr *NoteRepository) DeleteNote(noteID string) error {
	const query = `
	DELETE FROM notes
	WHERE id = ?;`

	result, err := nr.db.Exec(query, noteID)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete note rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNoteNotFound
	}

	return nil
}
