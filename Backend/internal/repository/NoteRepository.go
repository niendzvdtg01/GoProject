package repository

import (
	"backend/internal/model"
	"database/sql"
	"errors"
	"fmt"
)

var ErrNoteNotFound = errors.New("note not found")

type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (nr *NoteRepository) CreateNote(folderID int, ownerID, title, content string) (int, error) {
	const query = `
	INSERT INTO notes (folder_id, owner_id, title, content)
	VALUES (?, ?, ?, ?);`

	result, err := nr.db.Exec(query, folderID, ownerID, title, content)
	if err != nil {
		return 0, err
	}

	noteID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(noteID), nil
}

func (nr *NoteRepository) GetNoteByID(noteID int) (model.Note, error) {
	const query = `
	SELECT id, folder_id, owner_id, title, content, created_at, updated_at
	FROM notes
	WHERE id = ?;`

	var note model.Note
	err := nr.db.QueryRow(query, noteID).Scan(
		&note.ID,
		&note.FolderID,
		&note.OwnerID,
		&note.Title,
		&note.Content,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Note{}, ErrNoteNotFound
		}
		return model.Note{}, fmt.Errorf("get note by id: %w", err)
	}
	return note, nil
}

func (nr *NoteRepository) ListNotesByFolder(folderID int) ([]model.Note, error) {
	const query = `
	SELECT id, folder_id, owner_id, title, content, created_at, updated_at
	FROM notes
	WHERE folder_id = ?
	ORDER BY created_at DESC;`

	rows, err := nr.db.Query(query, folderID)
	if err != nil {
		return nil, fmt.Errorf("list notes by folder: %w", err)
	}
	defer rows.Close()

	notes := make([]model.Note, 0)
	for rows.Next() {
		var note model.Note
		if err := rows.Scan(&note.ID, &note.FolderID, &note.OwnerID, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (nr *NoteRepository) UpdateNote(noteID int, title, content string) (model.Note, error) {
	const query = `
	UPDATE notes
	SET title = ?, content = ?, updated_at = NOW()
	WHERE id = ?;`

	_, err := nr.db.Exec(query, title, content, noteID)
	if err != nil {
		return model.Note{}, fmt.Errorf("update note: %w", err)
	}

	return nr.GetNoteByID(noteID)
}

func (nr *NoteRepository) DeleteNote(noteID int) error {
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
