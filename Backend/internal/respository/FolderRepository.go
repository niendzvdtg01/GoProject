package respository

import "database/sql"

type FolderRepository struct {
	db *sql.DB
}

func NewFolderRepository(db *sql.DB) *FolderRepository {
	return &FolderRepository{db: db}
}

// CreateFolder creates a new folder in the database
func (fr *FolderRepository) CreateFolder(ownerID, name string) (int, error) {
	const query = `
	INSERT INTO folders (owner_id, name)
	VALUES (?, ?);`

	result, err := fr.db.Exec(query, ownerID, name)
	if err != nil {
		return 0, err
	}

	folderID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(folderID), nil
}

