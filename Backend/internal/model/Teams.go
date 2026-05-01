package model

type Team struct {
	TeamID    int    `json:"team_id" db:"team_id"`
	TeamName  string `json:"team_name" db:"name"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpDatedAt string `json:"updated_at" db:"updated_at"`
}

func (t Team) Public() Team {
	return Team{
		TeamID:    t.TeamID,
		TeamName:  t.TeamName,
		CreatedAt: t.CreatedAt,
		UpDatedAt: t.UpDatedAt,
	}
}
