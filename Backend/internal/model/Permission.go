package model

import "time"

type Permission struct {
	ID             int       `json:"id" db:"id"`
	AssetType      string    `json:"asset_type" db:"asset_type"`
	AssetID        int       `json:"asset_id" db:"asset_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	PermissionType string    `json:"permission_type" db:"permission_type"`
	GrantedBy      string    `json:"granted_by" db:"granted_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

func (p Permission) Public() Permission {
	return Permission{
		ID:             p.ID,
		AssetType:      p.AssetType,
		AssetID:        p.AssetID,
		UserID:         p.UserID,
		PermissionType: p.PermissionType,
		GrantedBy:      p.GrantedBy,
		CreatedAt:      p.CreatedAt,
	}
}
