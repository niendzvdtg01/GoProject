package model

type Note struct {
	ID        int    `json:"id" db:"id"`
	FolderID  int    `json:"folder_id" db:"folder_id"`
	OwnerID   string `json:"owner_id" db:"owner_id"`
	Title     string `json:"title" db:"title"`
	Content   string `json:"content" db:"content"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

func (n Note) Public() Note {
	return Note{
		ID:        n.ID,
		FolderID:  n.FolderID,
		Title:     n.Title,
		Content:   n.Content,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}
