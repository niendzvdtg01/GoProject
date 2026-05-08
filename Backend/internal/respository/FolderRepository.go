package respository

import (
	"backend/internal/model"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
)

var ErrFolderNotFound = errors.New("folder not found")

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

func (fr *FolderRepository) GetFolderByID(folderID int) (model.Folder, error) {
	const query = `
	SELECT id, owner_id, name, created_at, updated_at
	FROM folders
	WHERE id = ?;`

	var folder model.Folder
	var id int
	err := fr.db.QueryRow(query, folderID).Scan(
		&id,
		&folder.OwnerID,
		&folder.Name,
		&folder.CreatedAt,
		&folder.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Folder{}, ErrFolderNotFound
		}
		return model.Folder{}, fmt.Errorf("get folder by id: %w", err)
	}
	folder.ID = strconv.Itoa(id)
	return folder, nil
}

func (fr *FolderRepository) ListFoldersByOwner(ownerID string) ([]model.Folder, error) {
	const query = `
	SELECT id, owner_id, name, created_at, updated_at
	FROM folders
	WHERE owner_id = ?
	ORDER BY created_at DESC;`

	rows, err := fr.db.Query(query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list folders by owner: %w", err)
	}
	defer rows.Close()

	folders := make([]model.Folder, 0)
	for rows.Next() {
		var folder model.Folder
		var id int
		if err := rows.Scan(&id, &folder.OwnerID, &folder.Name, &folder.CreatedAt, &folder.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan folder: %w", err)
		}
		folder.ID = strconv.Itoa(id)
		folders = append(folders, folder)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate folders: %w", err)
	}

	return folders, nil
}

func (fr *FolderRepository) UpdateFolder(folderID int, name string) (model.Folder, error) {
	const query = `
	UPDATE folders
	SET name = ?, updated_at = NOW()
	WHERE id = ?;`

	_, err := fr.db.Exec(query, name, folderID)
	if err != nil {
		return model.Folder{}, fmt.Errorf("update folder: %w", err)
	}

	return fr.GetFolderByID(folderID)
}

func (fr *FolderRepository) DeleteFolder(folderID int) error {
	const query = `
	DELETE FROM folders
	WHERE id = ?;`

	result, err := fr.db.Exec(query, folderID)
	if err != nil {
		return fmt.Errorf("delete folder: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete folder rows affected: %w", err)
	}
	if affected == 0 {
		return ErrFolderNotFound
	}

	return nil
}
