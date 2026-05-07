package respository

import "database/sql"

type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// CreateNote creates a new note in the database
func (nr *NoteRepository) CreateNote(ownerID, folderID, title, content string) (int, error) {
	const query = `
	INSERT INTO notes (owner_id, folder_id, title, content)
	VALUES (?, ?, ?, ?);`

	result, err := nr.db.Exec(query, ownerID, folderID, title, content)
	if err != nil {
		return 0, err
	}

	noteID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(noteID), nil
}
