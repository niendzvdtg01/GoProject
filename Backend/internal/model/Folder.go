package model

type Folder struct {
	ID        int    `json:"id" db:"id"`
	OwnerID   string `json:"owner_id" db:"owner_id"`
	Name      string `json:"name" db:"name"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

func (f Folder) Public() Folder {
	return Folder{
		ID:        f.ID,
		Name:      f.Name,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}
