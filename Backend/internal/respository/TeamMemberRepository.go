package respository

import (
	"backend/internal/model"
	"database/sql"
)

const (
	roleOwner   = "owner"
	roleManager = "manager"
	roleMember  = "member"
)

type TeamMemberRepository struct {
	db *sql.DB
}

func NewTeamMemberRepository(db *sql.DB) *TeamMemberRepository {
	return &TeamMemberRepository{db: db}
}

// AddTeamMember adds a user to a team with a specific role
func (r *TeamMemberRepository) AddTeamMember(teamID int, userID string, role string) error {
	const query = `
	INSERT INTO team_members (team_id, user_id, role)
	VALUES (?, ?, ?);`

	_, err := r.db.Exec(query, teamID, userID, role)
	return err
}

// GetTeamMemberRole returns the assigned role for a user in a team.
func (r *TeamMemberRepository) GetTeamMemberRole(teamID int, userID string) (string, error) {
	const query = `
	SELECT role
	FROM team_members
	WHERE team_id = ? AND user_id = ?;`

	var role string
	err := r.db.QueryRow(query, teamID, userID).Scan(&role)
	if err != nil {
		return "", err
	}

	return role, nil
}

// GetTeamMembers retrieves all members of a specific team
func (r *TeamMemberRepository) GetTeamMembers(teamID int) ([]model.TeamMember, error) {
	const query = `
	SELECT team_id, user_id, role, joined_at
	FROM team_members
	WHERE team_id = ?;`

	rows, err := r.db.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.TeamMember
	for rows.Next() {
		var member model.TeamMember
		err := rows.Scan(&member.TeamID, &member.UserID, &member.Role, &member.JoinedAt)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}
