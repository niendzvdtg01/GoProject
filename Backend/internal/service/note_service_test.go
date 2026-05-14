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

func newNoteSetup(t *testing.T) (*NoteService, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	svc := NewNoteService(
		repository.NewNoteRepository(db),
		repository.NewFolderRepository(db),
		repository.NewUserRepository(db),
		repository.NewPermissionRepository(db),
		repository.NewTeamMemberRepository(db),
	)
	return svc, mock, func() { db.Close() }
}

func noteRow(id, folderID int, ownerID, title, content string) *sqlmock.Rows {
	now := time.Now().Format(time.RFC3339)
	return sqlmock.NewRows([]string{"note_id", "folder_id", "owner_id", "title", "content", "created_at", "updated_at"}).
		AddRow(id, folderID, ownerID, title, content, now, now)
}

const noteSelectQ = "SELECT note_id, folder_id, owner_id, title, content, created_at, updated_at FROM notes WHERE note_id = ?"

// ── CreateNote ────────────────────────────────────────────────────────────────

func TestCreateNote_Success(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "uid-1", "Docs"))

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO notes (folder_id, owner_id, title, content)")).
		WithArgs(1, "uid-1", "Hello", "World").
		WillReturnResult(sqlmock.NewResult(10, 1))

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "uid-1", "Hello", "World"))

	note, err := svc.CreateNote(context.Background(), 1, "uid-1", "Hello", "World")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.Title != "Hello" {
		t.Errorf("expected title Hello, got %s", note.Title)
	}
}

func TestCreateNote_NotFolderOwner(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "real-owner", "Docs"))

	_, err := svc.CreateNote(context.Background(), 1, "other-uid", "Title", "Content")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestCreateNote_FolderNotFound(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	_, err := svc.CreateNote(context.Background(), 99, "uid-1", "Title", "Content")
	if err == nil {
		t.Fatal("expected error for missing folder")
	}
}

// ── GetNote ───────────────────────────────────────────────────────────────────

func TestGetNote_ByOwner(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "uid-1", "Hello", "World"))

	note, err := svc.GetNote(context.Background(), 10, "uid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.ID != 10 {
		t.Errorf("expected ID=10, got %d", note.ID)
	}
}

func TestGetNote_ByNotePermission(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "owner-uid", "Hello", "World"))

	// Direct note permission exists
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("note", 10, "shared-uid").
		WillReturnRows(sqlmock.NewRows([]string{"id", "asset_type", "asset_id", "user_id", "permission_type", "granted_by", "created_at"}).
			AddRow(1, "note", 10, "shared-uid", "read", "owner-uid", time.Now()))

	_, err := svc.GetNote(context.Background(), 10, "shared-uid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetNote_ByFolderPermission(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "owner-uid", "Hello", "World"))

	// No direct note permission
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("note", 10, "shared-uid").
		WillReturnError(sql.ErrNoRows)

	// Folder permission exists
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("folder", 1, "shared-uid").
		WillReturnRows(sqlmock.NewRows([]string{"id", "asset_type", "asset_id", "user_id", "permission_type", "granted_by", "created_at"}).
			AddRow(2, "folder", 1, "shared-uid", "read", "owner-uid", time.Now()))

	_, err := svc.GetNote(context.Background(), 10, "shared-uid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetNote_Forbidden(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "owner-uid", "Secret", "Content"))

	// No note permission
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("note", 10, "other-uid").
		WillReturnError(sql.ErrNoRows)

	// No folder permission
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("folder", 1, "other-uid").
		WillReturnError(sql.ErrNoRows)

	// Get folder to check team manager
	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "owner-uid", "Docs"))

	// Not a manager
	mock.ExpectQuery("SELECT COUNT").
		WithArgs("other-uid", "owner-uid").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	_, err := svc.GetNote(context.Background(), 10, "other-uid")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// ── UpdateNote ────────────────────────────────────────────────────────────────

func TestUpdateNote_ByOwner(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "uid-1", "Old", "OldContent"))

	mock.ExpectExec(regexp.QuoteMeta("UPDATE notes SET title = ?, content = ?, updated_at = NOW() WHERE note_id = ?")).
		WithArgs("New", "NewContent", 10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "uid-1", "New", "NewContent"))

	note, err := svc.UpdateNote(context.Background(), 10, "uid-1", "New", "NewContent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.Title != "New" {
		t.Errorf("expected title New, got %s", note.Title)
	}
}

func TestUpdateNote_Forbidden(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "owner-uid", "Secret", "Content"))

	// No note permission
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("note", 10, "other-uid").
		WillReturnError(sql.ErrNoRows)

	// No folder permission
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("folder", 1, "other-uid").
		WillReturnError(sql.ErrNoRows)

	_, err := svc.UpdateNote(context.Background(), 10, "other-uid", "Hacked", "Content")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateNote_WritePermissionOnly(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "owner-uid", "Title", "Content"))

	// Read-only permission on note — canWrite returns false
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("note", 10, "readonly-uid").
		WillReturnRows(sqlmock.NewRows([]string{"id", "asset_type", "asset_id", "user_id", "permission_type", "granted_by", "created_at"}).
			AddRow(1, "note", 10, "readonly-uid", "read", "owner-uid", time.Now()))

	_, err := svc.UpdateNote(context.Background(), 10, "readonly-uid", "Title", "Content")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden for read-only user, got %v", err)
	}
}

// ── DeleteNote ────────────────────────────────────────────────────────────────

func TestDeleteNote_ByOwner(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "uid-1", "ToDelete", ""))

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notes WHERE note_id = ?")).
		WithArgs(10).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.DeleteNote(context.Background(), 10, "uid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteNote_Forbidden(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(noteSelectQ)).
		WithArgs(10).
		WillReturnRows(noteRow(10, 1, "owner-uid", "Private", ""))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("note", 10, "other-uid").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("folder", 1, "other-uid").
		WillReturnError(sql.ErrNoRows)

	err := svc.DeleteNote(context.Background(), 10, "other-uid")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// ── ListFolderNotes ────────────────────────────────────────────────────────────

func TestListFolderNotes_ByOwner(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "uid-1", "Docs"))

	now := time.Now().Format(time.RFC3339)
	noteRows := sqlmock.NewRows([]string{"note_id", "folder_id", "owner_id", "title", "content", "created_at", "updated_at"}).
		AddRow(1, 1, "uid-1", "Note A", "", now, now).
		AddRow(2, 1, "uid-1", "Note B", "", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT note_id, folder_id, owner_id, title, content, created_at, updated_at FROM notes WHERE folder_id = ?")).
		WithArgs(1).
		WillReturnRows(noteRows)

	notes, err := svc.ListFolderNotes(context.Background(), 1, "uid-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(notes))
	}
}

func TestListFolderNotes_Forbidden(t *testing.T) {
	svc, mock, cleanup := newNoteSetup(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(folderSelectQ)).
		WithArgs(1).
		WillReturnRows(folderRow(1, "owner-uid", "Docs"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, asset_type, asset_id, user_id, permission_type, granted_by, created_at FROM permissions WHERE asset_type = ? AND asset_id = ? AND user_id = ?")).
		WithArgs("folder", 1, "other-uid").
		WillReturnError(sql.ErrNoRows)

	mock.ExpectQuery("SELECT COUNT").
		WithArgs("other-uid", "owner-uid").
		WillReturnRows(sqlmock.NewRows([]string{"COUNT(*)"}).AddRow(0))

	_, err := svc.ListFolderNotes(context.Background(), 1, "other-uid")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
