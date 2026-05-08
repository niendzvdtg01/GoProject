package model

type Notes struct {
	ID       string `json:"id" db:"id"`
	OwnerID  string `json:"owner_id" db:"owner_id"`
	FolderID string `json:"folder_id" db:"folder_id"`
	Title    string `json:"title" db:"title"`
	Content  string `json:"content" db:"content"`

	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

func (n Notes) Public() Notes {
	return Notes{
		ID:        n.ID,
		FolderID:  n.FolderID,
		Title:     n.Title,
		Content:   n.Content,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}
