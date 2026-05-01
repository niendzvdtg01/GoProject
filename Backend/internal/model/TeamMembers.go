package model

type TeamMember struct {
	TeamID   Team       `json:"team_id" db:"team_id"`
	UserID   PublicUser `json:"user_id" db:"user_id"`
	Role     string     `json:"role" db:"role"`
	JoinedAt string     `json:"joined_at" db:"joined_at"`
}

func (tm TeamMember) Public() TeamMember {
	return TeamMember{
		TeamID:   tm.TeamID.Public(),
		UserID:   tm.UserID,
		Role:     tm.Role,
		JoinedAt: tm.JoinedAt,
	}
}
