package model

type Team struct {
	TeamID    int    `json:"team_id" db:"team_id"`
	TeamName  string `json:"team_name" db:"team_name"`
	OwnerID   string `json:"owner_id" db:"owner_id"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

func (t Team) Public() Team {
	return Team{
		TeamID:    t.TeamID,
		TeamName:  t.TeamName,
		OwnerID:   t.OwnerID,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}
