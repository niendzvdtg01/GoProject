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

func newSharingSetup(t *testing.T) (*SharingService, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	svc := NewSharing(
		repository.NewNoteRepository(db),
		repository.NewFolderRepository(db),
		repository.NewUserRepository(db),
		repository.NewPermissionRepository(db),
	)
	return svc, mock, func() { db.Close() }
}

const permSelectQ = "SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?"

func permRow(id int, assetType string, assetID int, userID, permType string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "asset_type", "asset_id", "user_id", "permission_type", "granted_by", "created_at"}).
		AddRow(id, assetType, assetID, userID, permType, "granter-uid", time.Now())
}

// ── ShareAsset ────────────────────────────────────────────────────────────────

func TestShareAsset_Note_Success(t *testing.T) {
	svc, mock, cleanup := newSharingSetup(t)
	defer cleanup()

	// Find user by email
	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnRows(userRow("uid-alice", "alice", "alice@example.com"))

	// Verify note exists
	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "owner-uid", "Title", "Content"))

	// CreatePermission: check existing
	mock.ExpectQuery(regexp.QuoteMeta(permSelectQ)).
		WithArgs("note", 10, "uid-alice").
		WillReturnError(sql.ErrNoRows)

	// Insert permission
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO permissions (asset_type, asset_id, user_id, permission_type, granted_by)")).
		WithArgs("note", 10, "uid-alice", "read", "owner-uid").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := svc.ShareAsset(dtorequest.ShareAssetRequest{
		Email:          "alice@example.com",
		NoteID:         10,
		PermissionType: "read",
	}, "owner-uid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShareAsset_Folder_WithInheritance(t *testing.T) {
	svc, mock, cleanup := newSharingSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnRows(userRow("uid-alice", "alice", "alice@example.com"))

	// Verify folder exists
	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "owner-uid", "Docs"))

	// CreatePermission for folder: check existing
	mock.ExpectQuery(regexp.QuoteMeta(permSelectQ)).
		WithArgs("folder", 1, "uid-alice").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO permissions (asset_type, asset_id, user_id, permission_type, granted_by)")).
		WithArgs("folder", 1, "uid-alice", "read", "owner-uid").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// List notes in folder for inheritance
	now := time.Now().Format(time.RFC3339)
	noteRows := sqlmock.NewRows([]string{"note_id", "folder_id", "owner_id", "title", "content", "created_at", "updated_at"}).
		AddRow(10, 1, "owner-uid", "Note A", "", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT note_id, folder_id, owner_id, title, content, created_at, updated_at FROM notes WHERE folder_id = ?")).
		WithArgs(1).
		WillReturnRows(noteRows)

	// CreatePermission for note 10: check existing
	mock.ExpectQuery(regexp.QuoteMeta(permSelectQ)).
		WithArgs("note", 10, "uid-alice").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO permissions (asset_type, asset_id, user_id, permission_type, granted_by)")).
		WithArgs("note", 10, "uid-alice", "read", "owner-uid").
		WillReturnResult(sqlmock.NewResult(2, 1))

	err := svc.ShareAsset(dtorequest.ShareAssetRequest{
		Email:          "alice@example.com",
		FolderID:       1,
		PermissionType: "read",
	}, "owner-uid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShareAsset_UserNotFound(t *testing.T) {
	svc, mock, cleanup := newSharingSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("ghost@example.com").
		WillReturnError(sql.ErrNoRows)

	err := svc.ShareAsset(dtorequest.ShareAssetRequest{
		Email:          "ghost@example.com",
		NoteID:         10,
		PermissionType: "read",
	}, "owner-uid")
	if err == nil {
		t.Fatal("expected error: user not found")
	}
}

func TestShareAsset_NoteNotFound(t *testing.T) {
	svc, mock, cleanup := newSharingSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnRows(userRow("uid-alice", "alice", "alice@example.com"))

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	err := svc.ShareAsset(dtorequest.ShareAssetRequest{
		Email:          "alice@example.com",
		NoteID:         999,
		PermissionType: "read",
	}, "owner-uid")
	if err == nil {
		t.Fatal("expected error: note not found")
	}
}

func TestShareAsset_NoAssetID(t *testing.T) {
	svc, mock, cleanup := newSharingSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnRows(userRow("uid-alice", "alice", "alice@example.com"))

	err := svc.ShareAsset(dtorequest.ShareAssetRequest{
		Email:          "alice@example.com",
		PermissionType: "read",
	}, "owner-uid")
	if err == nil {
		t.Fatal("expected error: note_id or folder_id required")
	}
}

// ── RevokeAccess ──────────────────────────────────────────────────────────────

func TestRevokeAccess_Note_Success(t *testing.T) {
	svc, mock, cleanup := newSharingSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnRows(userRow("uid-alice", "alice", "alice@example.com"))

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("note", 10, "uid-alice").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.RevokeAccess(dtorequest.RevokeAccessRequest{
		Email:  "alice@example.com",
		NoteID: 10,
	}, "owner-uid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRevokeAccess_Folder_WithCascade(t *testing.T) {
	svc, mock, cleanup := newSharingSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnRows(userRow("uid-alice", "alice", "alice@example.com"))

	// Revoke folder permission
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("folder", 1, "uid-alice").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// List notes for cascade
	now := time.Now().Format(time.RFC3339)
	noteRows := sqlmock.NewRows([]string{"note_id", "folder_id", "owner_id", "title", "content", "created_at", "updated_at"}).
		AddRow(10, 1, "owner-uid", "Note A", "", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT note_id, folder_id, owner_id, title, content, created_at, updated_at FROM notes WHERE folder_id = ?")).
		WithArgs(1).
		WillReturnRows(noteRows)

	// Revoke note permission
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("note", 10, "uid-alice").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.RevokeAccess(dtorequest.RevokeAccessRequest{
		Email:    "alice@example.com",
		FolderID: 1,
	}, "owner-uid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRevokeAccess_UserNotFound(t *testing.T) {
	svc, mock, cleanup := newSharingSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("ghost@example.com").
		WillReturnError(sql.ErrNoRows)

	err := svc.RevokeAccess(dtorequest.RevokeAccessRequest{
		Email:  "ghost@example.com",
		NoteID: 10,
	}, "owner-uid")
	if err == nil {
		t.Fatal("expected error: user not found")
	}
}

func TestRevokeAccess_NoAssetID(t *testing.T) {
	svc, mock, cleanup := newSharingSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnRows(userRow("uid-alice", "alice", "alice@example.com"))

	err := svc.RevokeAccess(dtorequest.RevokeAccessRequest{
		Email: "alice@example.com",
	}, "owner-uid")
	if err == nil {
		t.Fatal("expected error: note_id or folder_id required")
	}
}

func TestRevokeAccess_PermissionNotFound(t *testing.T) {
	svc, mock, cleanup := newSharingSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT user_id, username, email, password_hash, role, created_at FROM users WHERE email = ?")).
		WithArgs("alice@example.com").
		WillReturnRows(userRow("uid-alice", "alice", "alice@example.com"))

	// No row deleted
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("note", 10, "uid-alice").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := svc.RevokeAccess(dtorequest.RevokeAccessRequest{
		Email:  "alice@example.com",
		NoteID: 10,
	}, "owner-uid")
	if !errors.Is(err, repository.ErrPermissionNotFound) {
		t.Errorf("expected ErrPermissionNotFound, got %v", err)
	}
}
