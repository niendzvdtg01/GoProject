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

func (tr *TeamRepository) CreateTeam(teamName, ownerID string) (int, error) {
	const query = `
	INSERT INTO teams (team_name, owner_id)
	VALUES (?, ?);`

	result, err := tr.db.Exec(query, teamName, ownerID)
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
	SELECT team_id, team_name, owner_id, created_at, updated_at
	FROM teams
	WHERE team_name = ?;`

	var team model.Team
	err := tr.db.QueryRow(query, teamName).Scan(
		&team.TeamID,
		&team.TeamName,
		&team.OwnerID,
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
	SELECT team_id, team_name, owner_id, created_at, updated_at
	FROM teams
	WHERE team_id = ?;`

	var team model.Team
	err := tr.db.QueryRow(query, teamID).Scan(
		&team.TeamID,
		&team.TeamName,
		&team.OwnerID,
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

func (tr *TeamRepository) FindTeamByOwnerID(ownerID string) (model.Team, error) {
	const query = `
	SELECT 
		t.team_id,
		t.team_name,
		t.owner_id,
		t.created_at,
		t.updated_at
	FROM teams t
	WHERE t.owner_id = ?
	LIMIT 1`

	var team model.Team
	err := tr.db.QueryRow(query, ownerID).Scan(
		&team.TeamID,
		&team.TeamName,
		&team.OwnerID,
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
