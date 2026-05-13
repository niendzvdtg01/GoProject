package repository

import (
	"backend/internal/model"
	"database/sql"
	"errors"
	"fmt"
)

var ErrPermissionAlreadyExists = errors.New("permission already exists")
var ErrPermissionNotFound = errors.New("permission not found")

// Asset types
const (
	AssetTypeFolder = "folder"
	AssetTypeNote   = "note"
)

type PermissionRepository struct {
	db *sql.DB
}

func NewPermissionRepository(db *sql.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

func (pr *PermissionRepository) GetPermissionByAssetAndUser(assetType string, assetID int, userID string) (model.Permission, error) {
	const query = `
	SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at
	FROM permissions
	WHERE asset_type = ? AND asset_id = ? AND user_id = ?;`

	var permission model.Permission
	err := pr.db.QueryRow(query, assetType, assetID, userID).Scan(
		&permission.ID,
		&permission.AssetType,
		&permission.AssetID,
		&permission.UserID,
		&permission.PermissionType,
		&permission.GrantedBy,
		&permission.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Permission{}, sql.ErrNoRows
		}
		return model.Permission{}, fmt.Errorf("get permission by asset and user: %w", err)
	}

	return permission, nil
}

func (pr *PermissionRepository) RevokePermission(assetType string, assetID int, userID string) error {
	const query = `
	DELETE FROM permissions
	WHERE asset_type = ? AND asset_id = ? AND user_id = ?;`

	result, err := pr.db.Exec(query, assetType, assetID, userID)
	if err != nil {
		return fmt.Errorf("revoke permission: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke permission rows affected: %w", err)
	}
	if affected == 0 {
		return ErrPermissionNotFound
	}

	return nil
}

func (pr *PermissionRepository) CreatePermission(assetType string, assetID int, userID, permissionType, grantedBy string) error {
	existing, err := pr.GetPermissionByAssetAndUser(assetType, assetID, userID)
	if err == nil && existing.ID != 0 {
		return ErrPermissionAlreadyExists
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	const query = `
	INSERT INTO permissions (asset_type, asset_id, user_id, permission_type, granted_by)
	VALUES (?, ?, ?, ?, ?);`

	_, execErr := pr.db.Exec(query, assetType, assetID, userID, permissionType, grantedBy)
	return execErr
}
