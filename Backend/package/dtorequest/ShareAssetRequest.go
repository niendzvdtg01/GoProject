package dtorequest

type ShareAssetRequest struct {
	Email          string `json:"email" binding:"required,email"`
	NoteID         int    `json:"note_id"`
	FolderID       int    `json:"folder_id"`
	PermissionType string `json:"permission_type" binding:"required,oneof=read write"`
}
