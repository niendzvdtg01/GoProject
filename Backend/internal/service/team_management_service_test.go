package service

import (
	"backend/internal/repository"
	"backend/package/dtorequest"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func newTeamSetup(t *testing.T) (*TeamManagementService, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	svc := NewTeamManagementService(
		repository.NewTeamRepository(db),
		repository.NewTeamMemberRepository(db),
		repository.NewUserRepository(db),
	)
	return svc, mock, func() { db.Close() }
}

func teamRow(id int, name string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"team_id", "team_name", "created_at", "updated_at"}).
		AddRow(id, name, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
}

func userRow(uid, username, email string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"user_id", "username", "email", "password_hash", "role", "created_at"}).
		AddRow(uid, username, email, "hash", "member", time.Now())
}

// ── CreateTeam ──────────────────────────────────────────────────────────────

func TestCreateTeam_Success(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	// Team doesn't exist yet
	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("alpha").
		WillReturnError(sql.ErrNoRows)

	// Insert team
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO teams (team_name)")).
		WithArgs("alpha").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Add creator as OWNER
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO team_members (team_id, user_id, role)")).
		WithArgs(1, "creator-uid", "OWNER").
		WillReturnResult(sqlmock.NewResult(1, 1))

	teamID, err := svc.CreateTeam("alpha", "creator-uid", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if teamID != 1 {
		t.Errorf("expected teamID=1, got %d", teamID)
	}
}

func TestCreateTeam_WithMembers(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("beta").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO teams (team_name)")).
		WillReturnResult(sqlmock.NewResult(2, 1))

	// Add creator
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO team_members (team_id, user_id, role)")).
		WithArgs(2, "creator-uid", "OWNER").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Lookup member
	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE username = ?")).
		WithArgs("alice").
		WillReturnRows(userRow("uid-alice", "alice", "alice@example.com"))

	// Add alice as member
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO team_members (team_id, user_id, role)")).
		WithArgs(2, "uid-alice", "MEMBER").
		WillReturnResult(sqlmock.NewResult(2, 1))

	_, err := svc.CreateTeam("beta", "creator-uid", []dtorequest.MemberRequest{
		{MemberName: "alice", Role: "MEMBER"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateTeam_AlreadyExists(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("alpha").
		WillReturnRows(teamRow(1, "alpha"))

	_, err := svc.CreateTeam("alpha", "uid-1", nil)
	if err == nil {
		t.Fatal("expected error for duplicate team")
	}
}

// ── AddMemberByName ──────────────────────────────────────────────────────────

func TestAddMemberByName_Success(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("alpha").
		WillReturnRows(teamRow(1, "alpha"))

	// Actor is OWNER
	mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "owner-uid").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("OWNER"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE username = ?")).
		WithArgs("alice").
		WillReturnRows(userRow("uid-alice", "alice", "alice@example.com"))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO team_members (team_id, user_id, role)")).
		WithArgs(1, "uid-alice", "MEMBER").
		WillReturnResult(sqlmock.NewResult(1, 1))

	team, err := svc.AddMemberByName("alpha", "owner-uid", "alice", "MEMBER")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.TeamName != "alpha" {
		t.Errorf("expected team alpha, got %s", team.TeamName)
	}
}

func TestAddMemberByName_TeamNotFound(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("ghost").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.AddMemberByName("ghost", "uid-1", "alice", "MEMBER")
	if err == nil {
		t.Fatal("expected error for team not found")
	}
}

func TestAddMemberByName_ActorIsMember_Forbidden(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("alpha").
		WillReturnRows(teamRow(1, "alpha"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "member-uid").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("MEMBER"))

	_, err := svc.AddMemberByName("alpha", "member-uid", "alice", "MEMBER")
	if err == nil {
		t.Fatal("expected error: member cannot add others")
	}
}

// ── RemoveMemberByName ───────────────────────────────────────────────────────

func TestRemoveMemberByName_Success(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("alpha").
		WillReturnRows(teamRow(1, "alpha"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "owner-uid").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("OWNER"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE username = ?")).
		WithArgs("alice").
		WillReturnRows(userRow("uid-alice", "alice", "alice@example.com"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "uid-alice").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("MEMBER"))

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "uid-alice").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.RemoveMemberByName("alpha", "owner-uid", "alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveMemberByName_CannotRemoveOwner(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("alpha").
		WillReturnRows(teamRow(1, "alpha"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "owner-uid").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("OWNER"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE username = ?")).
		WithArgs("owner2").
		WillReturnRows(userRow("uid-owner2", "owner2", "owner2@example.com"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "uid-owner2").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("OWNER"))

	err := svc.RemoveMemberByName("alpha", "owner-uid", "owner2")
	if err == nil {
		t.Fatal("expected error: owner cannot be removed")
	}
}

func TestRemoveMemberByName_ManagerCannotRemoveManager(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("alpha").
		WillReturnRows(teamRow(1, "alpha"))

	// Actor is MANAGER
	mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "mgr-uid").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("MANAGER"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE username = ?")).
		WithArgs("othermgr").
		WillReturnRows(userRow("uid-othermgr", "othermgr", "mgr2@example.com"))

	// Target is also MANAGER
	mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "uid-othermgr").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("MANAGER"))

	err := svc.RemoveMemberByName("alpha", "mgr-uid", "othermgr")
	if err == nil {
		t.Fatal("expected error: manager cannot remove another manager")
	}
}

func TestRemoveMemberByName_ActorIsMember_Forbidden(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("alpha").
		WillReturnRows(teamRow(1, "alpha"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "member-uid").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("MEMBER"))

	err := svc.RemoveMemberByName("alpha", "member-uid", "alice")
	if err == nil {
		t.Fatal("expected error: member cannot remove others")
	}
}

// ── DeleteTeam ───────────────────────────────────────────────────────────────

func TestDeleteTeam_Success(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("alpha").
		WillReturnRows(teamRow(1, "alpha"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "owner-uid").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("OWNER"))

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM teams WHERE team_name = ?")).
		WithArgs("alpha").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.DeleteTeam("alpha", "owner-uid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteTeam_NotOwner(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("alpha").
		WillReturnRows(teamRow(1, "alpha"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT role FROM team_members WHERE team_id = ? AND user_id = ?")).
		WithArgs(1, "member-uid").
		WillReturnRows(sqlmock.NewRows([]string{"role"}).AddRow("MEMBER"))

	err := svc.DeleteTeam("alpha", "member-uid")
	if err == nil {
		t.Fatal("expected error: only owner can delete team")
	}
}

func TestDeleteTeam_TeamNotFound(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT team_id, team_name, created_at, updated_at FROM teams WHERE team_name = ?")).
		WithArgs("ghost").
		WillReturnError(sql.ErrNoRows)

	err := svc.DeleteTeam("ghost", "uid-1")
	if err == nil {
		t.Fatal("expected error: team not found")
	}
}

// ── ListTeamsForUser ──────────────────────────────────────────────────────────

func TestListTeamsForUser_Success(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	now := time.Now().Format(time.RFC3339)
	rows := sqlmock.NewRows([]string{"team_id", "team_name", "created_at", "updated_at"}).
		AddRow(1, "alpha", now, now).
		AddRow(2, "beta", now, now)

	mock.ExpectQuery("SELECT t.team_id, t.team_name, t.created_at, t.updated_at").
		WithArgs("uid-1").
		WillReturnRows(rows)

	teams, err := svc.ListTeamsForUser("uid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(teams) != 2 {
		t.Errorf("expected 2 teams, got %d", len(teams))
	}
}

func TestListTeamsForUser_DBError(t *testing.T) {
	svc, mock, cleanup := newTeamSetup(t)
	defer cleanup()

	mock.ExpectQuery("SELECT t.team_id, t.team_name, t.created_at, t.updated_at").
		WithArgs("uid-1").
		WillReturnError(errors.New("connection lost"))

	_, err := svc.ListTeamsForUser("uid-1")
	if err == nil {
		t.Fatal("expected error")
	}
}
