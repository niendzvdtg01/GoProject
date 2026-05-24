package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/package/event"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type NoteService struct {
	notes       *repository.NoteRepository
	folder      *repository.FolderRepository
	users       *repository.UserRepository
	permission  *repository.PermissionRepository
	teamMembers *repository.TeamMemberRepository
	publisher   event.Publisher
}

func NewNoteService(notes *repository.NoteRepository, folder *repository.FolderRepository, users *repository.UserRepository, permission *repository.PermissionRepository, teamMembers *repository.TeamMemberRepository) *NoteService {
	return &NoteService{
		notes:       notes,
		folder:      folder,
		users:       users,
		permission:  permission,
		teamMembers: teamMembers,
		publisher:   event.NewNoopPublisher(),
	}
}

func (s *NoteService) WithPublisher(p event.Publisher) *NoteService {
	if p != nil {
		s.publisher = p
	}
	return s
}

// canRead checks: owner → note permission → folder permission (inherited) → team manager.
func (s *NoteService) canRead(requesterID string, note model.Note) (bool, error) {
	if requesterID == note.OwnerID {
		return true, nil
	}

	perm, err := s.permission.GetPermissionByAssetAndUser(repository.AssetTypeNote, note.ID, requesterID)
	if err == nil {
		_ = perm
		return true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	// Folder permission satisfies note read access (inherited from ShareAsset propagation).
	folderPerm, err := s.permission.GetPermissionByAssetAndUser(repository.AssetTypeFolder, note.FolderID, requesterID)
	if err == nil {
		_ = folderPerm
		return true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	folder, err := s.folder.GetFolderByID(note.FolderID)
	if err != nil {
		return false, err
	}

	isManager, err := s.teamMembers.IsManagerOf(requesterID, folder.OwnerID)
	if err != nil {
		return false, err
	}
	return isManager, nil
}

// canWrite checks: owner → note write permission → folder write permission (managers excluded, same as FolderService.canWrite).
func (s *NoteService) canWrite(requesterID string, note model.Note) (bool, error) {
	if requesterID == note.OwnerID {
		return true, nil
	}

	perm, err := s.permission.GetPermissionByAssetAndUser(repository.AssetTypeNote, note.ID, requesterID)
	if err == nil {
		return perm.PermissionType == "write", nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	folderPerm, err := s.permission.GetPermissionByAssetAndUser(repository.AssetTypeFolder, note.FolderID, requesterID)
	if err == nil {
		return folderPerm.PermissionType == "write", nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return false, nil
}

// CreateNote creates a note in a folder; only the folder owner may add notes (shared write access is insufficient).
func (s *NoteService) CreateNote(ctx context.Context, folderID int, ownerID, title, content string) (model.Note, error) {
	folder, err := s.folder.GetFolderByID(folderID)
	if err != nil {
		return model.Note{}, fmt.Errorf("folder not found: %w", err)
	}

	if folder.OwnerID != ownerID {
		return model.Note{}, ErrForbidden
	}

	id, err := s.notes.CreateNote(folderID, ownerID, title, content)
	if err != nil {
		return model.Note{}, fmt.Errorf("create note: %w", err)
	}

	note, err := s.notes.GetNoteByID(id)
	if err != nil {
		return model.Note{}, err
	}

	s.publisher.PublishAssetEvent(ctx, event.NoteCreated, ownerID, map[string]any{
		"note_id":   note.ID,
		"folder_id": note.FolderID,
		"title":     note.Title,
	})
	return note, nil
}

func (s *NoteService) GetNote(ctx context.Context, noteID int, requesterID string) (model.Note, error) {
	note, err := s.notes.GetNoteByID(noteID)
	if err != nil {
		return model.Note{}, err
	}

	ok, err := s.canRead(requesterID, note)
	if err != nil {
		return model.Note{}, err
	}
	if !ok {
		return model.Note{}, ErrForbidden
	}

	return note, nil
}

// ListFolderNotes returns all notes in a folder; access is checked at the folder level (ownership, explicit permission, or manager).
func (s *NoteService) ListFolderNotes(ctx context.Context, folderID int, requesterID string) ([]model.Note, error) {
	folder, err := s.folder.GetFolderByID(folderID)
	if err != nil {
		return nil, err
	}

	if requesterID == folder.OwnerID {
		return s.notes.ListNotesByFolder(folderID)
	}

	perm, err := s.permission.GetPermissionByAssetAndUser(repository.AssetTypeFolder, folderID, requesterID)
	if err == nil {
		_ = perm
		return s.notes.ListNotesByFolder(folderID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	isManager, err := s.teamMembers.IsManagerOf(requesterID, folder.OwnerID)
	if err != nil {
		return nil, err
	}
	if isManager {
		return s.notes.ListNotesByFolder(folderID)
	}

	return nil, ErrForbidden
}

func (s *NoteService) UpdateNote(ctx context.Context, noteID int, requesterID, title, content string) (model.Note, error) {
	note, err := s.notes.GetNoteByID(noteID)
	if err != nil {
		return model.Note{}, err
	}

	ok, err := s.canWrite(requesterID, note)
	if err != nil {
		return model.Note{}, err
	}
	if !ok {
		return model.Note{}, ErrForbidden
	}

	updated, err := s.notes.UpdateNote(noteID, title, content)
	if err != nil {
		return model.Note{}, err
	}

	s.publisher.PublishAssetEvent(ctx, event.NoteUpdated, requesterID, map[string]any{
		"note_id":   updated.ID,
		"folder_id": updated.FolderID,
		"title":     updated.Title,
		"owner_id":  updated.OwnerID,
	})
	return updated, nil
}

func (s *NoteService) DeleteNote(ctx context.Context, noteID int, requesterID string) error {
	note, err := s.notes.GetNoteByID(noteID)
	if err != nil {
		return err
	}

	ok, err := s.canWrite(requesterID, note)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}

	if err := s.notes.DeleteNote(noteID); err != nil {
		return err
	}

	s.publisher.PublishAssetEvent(ctx, event.NoteDeleted, requesterID, map[string]any{
		"note_id":   note.ID,
		"folder_id": note.FolderID,
		"title":     note.Title,
		"owner_id":  note.OwnerID,
	})
	return nil
}
