package dtorequest

type MemberRequest struct {
	MemberName string `json:"user_id"`
	Role       string `json:"role"` // "member" or "manager"
}
