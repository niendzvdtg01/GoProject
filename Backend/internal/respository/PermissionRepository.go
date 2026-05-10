package respository

import (
	"backend/internal/model"
	"database/sql"
	"errors"
	"fmt"
)

var ErrPermissionAlreadyExists = errors.New("permission already exists")

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

func (pr *PermissionRepository) GetPermissionByAssetAndUser(assetType, assetID, userID string) (model.Permissions, error) {
	const query = `
	SELECT permissions_id, asset_type, asset_id, user_id, permission_type, granted_by, created_at
	FROM permissions
	WHERE asset_type = ? AND asset_id = ? AND user_id = ?;`

	var permission model.Permissions
	err := pr.db.QueryRow(query, assetType, assetID, userID).Scan(
		&permission.PermissionsID,
		&permission.AssetType,
		&permission.AssetID,
		&permission.UserID,
		&permission.PermissionType,
		&permission.GrantedBy,
		&permission.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Permissions{}, sql.ErrNoRows
		}
		return model.Permissions{}, fmt.Errorf("get permission by asset and user: %w", err)
	}

	return permission, nil
}

func (pr *PermissionRepository) CreatePermission(assetType, assetID, userID, permissionType, grantedBy string) error {
	existing, err := pr.GetPermissionByAssetAndUser(assetType, assetID, userID)
	if err == nil && existing.PermissionsID != "" {
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
