package service

import (
	"backend/internal/repository"
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func newFolderSetup(t *testing.T) (*FolderService, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	svc := NewFolderService(
		repository.NewFolderRepository(db),
		repository.NewUserRepository(db),
		repository.NewPermissionRepository(db),
		repository.NewTeamMemberRepository(db),
	)
	return svc, mock, func() { db.Close() }
}

func folderRow(id int, ownerID, name string) *sqlmock.Rows {
	now := time.Now().Format(time.RFC3339)
	return sqlmock.NewRows([]string{"folder_id", "owner_id", "name", "created_at", "updated_at"}).
		AddRow(id, ownerID, name, now, now)
}

const folderSelectQ = "SELECT folder_id, owner_id, name, created_at, updated_at FROM folders WHERE folder_id = ?"

// ── CreateFolder ──────────────────────────────────────────────────────────────

func TestCreateFolder_Success(t *testing.T) {
	svc, mock, cleanup := newFolderSetup(t)
	defer cleanup()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO folders (owner_id, name)")).
		WithArgs("uid-1", "My Folder").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "uid-1", "My Folder"))

	folder, err := svc.CreateFolder(context.Background(), "uid-1", "My Folder")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if folder.Name != "My Folder" {
		t.Errorf("expected name 'My Folder', got %s", folder.Name)
	}
	if folder.OwnerID != "uid-1" {
		t.Errorf("expected ownerID uid-1, got %s", folder.OwnerID)
	}
}

// ── GetFolder ─────────────────────────────────────────────────────────────────

func TestGetFolder_ByOwner(t *testing.T) {
	svc, mock, cleanup := newFolderSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "uid-1", "Docs"))

	folder, err := svc.GetFolder(context.Background(), 1, "uid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if folder.ID != 1 {
		t.Errorf("expected ID=1, got %d", folder.ID)
	}
}

func TestGetFolder_BySharedUser_WithPermission(t *testing.T) {
	svc, mock, cleanup := newFolderSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "owner-uid", "Docs"))

	// Permission check — shared user has permission
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("folder", 1, "shared-uid").
		WillReturnRows(sqlmock.NewRows([]string{"id", "asset_type", "asset_id", "user_id", "permission_type", "granted_by", "created_at"}).
			AddRow(1, "folder", 1, "shared-uid", "read", "owner-uid", time.Now()))

	_, err := svc.GetFolder(context.Background(), 1, "shared-uid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetFolder_Forbidden(t *testing.T) {
	svc, mock, cleanup := newFolderSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "owner-uid", "Private"))

	// No permission
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("folder", 1, "other-uid").
		WillReturnError(sql.ErrNoRows)

	// Not a manager either
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("other-uid", "owner-uid").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	_, err := svc.GetFolder(context.Background(), 1, "other-uid")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestGetFolder_ByTeamManager(t *testing.T) {
	svc, mock, cleanup := newFolderSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "owner-uid", "Docs"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("folder", 1, "mgr-uid").
		WillReturnError(sql.ErrNoRows)

	// Manager check returns count > 0
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("mgr-uid", "owner-uid").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(1))

	_, err := svc.GetFolder(context.Background(), 1, "mgr-uid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetFolder_NotFound(t *testing.T) {
	svc, mock, cleanup := newFolderSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.GetFolder(context.Background(), 99, "uid-1")
	if !errors.Is(err, repository.ErrFolderNotFound) {
		t.Errorf("expected ErrFolderNotFound, got %v", err)
	}
}

// ── ListUserFolders ────────────────────────────────────────────────────────────

func TestListUserFolders_Success(t *testing.T) {
	svc, mock, cleanup := newFolderSetup(t)
	defer cleanup()

	now := time.Now().Format(time.RFC3339)
	rows := sqlmock.NewRows([]string{"folder_id", "owner_id", "name", "created_at", "updated_at"}).
		AddRow(1, "uid-1", "Docs", now, now).
		AddRow(2, "uid-1", "Images", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT folder_id, owner_id, name, created_at, updated_at FROM folders WHERE owner_id = ?")).
		WithArgs("uid-1").
		WillReturnRows(rows)

	folders, err := svc.ListUserFolders(context.Background(), "uid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(folders) != 2 {
		t.Errorf("expected 2, got %d", len(folders))
	}
}

// ── UpdateFolder ──────────────────────────────────────────────────────────────

func TestUpdateFolder_ByOwner(t *testing.T) {
	svc, mock, cleanup := newFolderSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "uid-1", "OldName"))

	mock.ExpectExec(regexp.QuoteMeta("UPDATE folders SET name = ?, updated_at = NOW() WHERE folder_id = ?")).
		WithArgs("NewName", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "uid-1", "NewName"))

	folder, err := svc.UpdateFolder(context.Background(), 1, "uid-1", "NewName")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if folder.Name != "NewName" {
		t.Errorf("expected NewName, got %s", folder.Name)
	}
}

func TestUpdateFolder_Forbidden(t *testing.T) {
	svc, mock, cleanup := newFolderSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "owner-uid", "Private"))

	// No write permission (read-only)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("folder", 1, "other-uid").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.UpdateFolder(context.Background(), 1, "other-uid", "Hacked")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// ── DeleteFolder ──────────────────────────────────────────────────────────────

func TestDeleteFolder_ByOwner(t *testing.T) {
	svc, mock, cleanup := newFolderSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "uid-1", "ToDelete"))

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM folders WHERE folder_id = ?")).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.DeleteFolder(context.Background(), 1, "uid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteFolder_Forbidden(t *testing.T) {
	svc, mock, cleanup := newFolderSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "owner-uid", "Private"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("folder", 1, "other-uid").
		WillReturnError(sql.ErrNoRows)

	err := svc.DeleteFolder(context.Background(), 1, "other-uid")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
