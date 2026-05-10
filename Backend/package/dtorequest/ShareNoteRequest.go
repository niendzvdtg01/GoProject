package dtorequest

type ShareNoteRequest struct {
	NoteID         string `json:"note_id" binding:"required"`
	UserID         string `json:"user_id" binding:"required"`
	PermissionType string `json:"permission_type" binding:"required,oneof=read write"`
}
