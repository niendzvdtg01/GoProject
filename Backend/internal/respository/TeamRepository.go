package respository

import (
	"backend/internal/model"
	"database/sql"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRespository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(teamName string) (int, error) {
	const query = `
	INSERT INTO teams (team_name)
	VALUES (?);`

	result, err := r.db.Exec(query, teamName)
	if err != nil {
		return 0, err
	}

	teamID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(teamID), nil
}

func (r *TeamRepository) GetTeamByName(teamName string) (model.Team, error) {
	const query = `
	SELECT team_id, team_name, created_at, updated_at
	FROM teams
	WHERE team_name = ?;`

	var team model.Team
	err := r.db.QueryRow(query, teamName).Scan(
		&team.TeamID,
		&team.TeamName,
		&team.CreatedAt,
		&team.UpDatedAt,
	)
	if err != nil {
		return model.Team{}, err
	}
	return team, nil
}

func (r *TeamRepository) UpdateTeamByName(teamName string, newName string) (model.Team, error) {
	const query = `
	UPDATE teams
	SET team_name = ?, updated_at = NOW()
	WHERE team_name = ?;`

	_, err := r.db.Exec(query, newName, teamName)
	if err != nil {
		return model.Team{}, err
	}

	return r.GetTeamByName(newName)
}

func (r *TeamRepository) DeleteTeamByName(teamName string) error {
	const query = `
	DELETE FROM teams
	WHERE team_name = ?;`

	_, err := r.db.Exec(query, teamName)
	return err
}
