package dtorequest

type RevokeAccessRequest struct {
	Email    string `json:"email" binding:"required,email"`
	NoteID   int    `json:"note_id"`
	FolderID int    `json:"folder_id"`
}
