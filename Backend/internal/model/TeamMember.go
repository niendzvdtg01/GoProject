package model

import "time"

type TeamMember struct {
	TeamID   int       `json:"team_id" db:"team_id"`
	UserID   string    `json:"user_id" db:"user_id"`
	Role     string    `json:"role" db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

func (tm TeamMember) Public() TeamMember {
	return TeamMember{
		TeamID:   tm.TeamID,
		UserID:   tm.UserID,
		Role:     tm.Role,
		JoinedAt: tm.JoinedAt,
	}
}
