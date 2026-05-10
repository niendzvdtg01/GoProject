package model

import "time"

type Permissions struct {
	PermissionsID  string    `json:"permissions_id" db:"permissions_id"`
	AssetType      string    `json:"asset_type" db:"asset_type"`
	AssetID        string    `json:"asset_id" db:"asset_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	PermissionType string    `json:"permission_type" db:"permission_type"`
	GrantedBy      string    `json:"granted_by" db:"granted_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

func (p Permissions) Public() Permissions {
	return Permissions{
		PermissionsID:  p.PermissionsID,
		AssetType:      p.AssetType,
		AssetID:        p.AssetID,
		UserID:         p.UserID,
		PermissionType: p.PermissionType,
		GrantedBy:      p.GrantedBy,
		CreatedAt:      p.CreatedAt,
	}
}
