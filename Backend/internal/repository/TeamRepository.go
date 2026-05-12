package repository

import (
	"backend/internal/model"
	"database/sql"
	"errors"
)

var ErrTeamNotFound = errors.New("team not found")

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (tr *TeamRepository) CreateTeam(teamName string) (int, error) {
	const query = `
	INSERT INTO teams (team_name)
	VALUES (?);`

	result, err := tr.db.Exec(query, teamName)
	if err != nil {
		return 0, err
	}

	teamID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(teamID), nil
}

func (tr *TeamRepository) GetTeamByName(teamName string) (model.Team, error) {
	const query = `
	SELECT team_id, team_name, created_at, updated_at
	FROM teams
	WHERE team_name = ?;`

	var team model.Team
	err := tr.db.QueryRow(query, teamName).Scan(
		&team.TeamID,
		&team.TeamName,
		&team.CreatedAt,
		&team.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Team{}, ErrTeamNotFound
		}
		return model.Team{}, err
	}
	return team, nil
}

func (tr *TeamRepository) GetTeamByID(teamID int) (model.Team, error) {
	const query = `
	SELECT team_id, team_name, created_at, updated_at
	FROM teams
	WHERE team_id = ?;`

	var team model.Team
	err := tr.db.QueryRow(query, teamID).Scan(
		&team.TeamID,
		&team.TeamName,
		&team.CreatedAt,
		&team.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Team{}, ErrTeamNotFound
		}
		return model.Team{}, err
	}
	return team, nil
}

func (tr *TeamRepository) UpdateTeamByName(teamName string, newName string) (model.Team, error) {
	const query = `
	UPDATE teams
	SET team_name = ?, updated_at = NOW()
	WHERE team_name = ?;`

	_, err := tr.db.Exec(query, newName, teamName)
	if err != nil {
		return model.Team{}, err
	}

	return tr.GetTeamByName(newName)
}

func (tr *TeamRepository) DeleteTeamByName(teamName string) error {
	const query = `
	DELETE FROM teams
	WHERE team_name = ?;`

	_, err := tr.db.Exec(query, teamName)
	return err
}
