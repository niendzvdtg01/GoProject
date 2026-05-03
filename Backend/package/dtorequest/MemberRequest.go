package dtorequest

type MemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"` // "member" or "manager"
}
