package dtorequest

type MemberRequest struct {
	MemberName string `json:"username"`
	Role       string `json:"role"` // "member" or "manager"
}
