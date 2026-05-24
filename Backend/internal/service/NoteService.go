package service

import (
	"backend/internal/cache"
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
	assetCache  *cache.AssetMetadata
	aclCache    *cache.ACL
}

func NewNoteService(notes *repository.NoteRepository, folder *repository.FolderRepository, users *repository.UserRepository, permission *repository.PermissionRepository, teamMembers *repository.TeamMemberRepository) *NoteService {
	noop := cache.NewNoopCache()
	return &NoteService{
		notes:       notes,
		folder:      folder,
		users:       users,
		permission:  permission,
		teamMembers: teamMembers,
		publisher:   event.NewNoopPublisher(),
		assetCache:  cache.NewAssetMetadata(noop),
		aclCache:    cache.NewACL(noop),
	}
}

func (s *NoteService) WithPublisher(p event.Publisher) *NoteService {
	if p != nil {
		s.publisher = p
	}
	return s
}

// WithCache injects the metadata + ACL caches; defaults to noop in tests.
func (s *NoteService) WithCache(assets *cache.AssetMetadata, acl *cache.ACL) *NoteService {
	if assets != nil {
		s.assetCache = assets
	}
	if acl != nil {
		s.aclCache = acl
	}
	return s
}

// lookupPermission mirrors FolderService.lookupPermission: read-through cache
// with negative caching so "not permitted" answers don't repeatedly hit the DB.
func (s *NoteService) lookupPermission(ctx context.Context, assetType string, assetID int, userID string) (string, error) {
	if cached, ok := s.aclCache.Get(ctx, assetType, assetID, userID); ok {
		return cached, nil
	}
	perm, err := s.permission.GetPermissionByAssetAndUser(assetType, assetID, userID)
	if err == nil {
		_ = s.aclCache.Set(ctx, assetType, assetID, userID, perm.PermissionType)
		return perm.PermissionType, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	_ = s.aclCache.Set(ctx, assetType, assetID, userID, cache.PermissionNone)
	return cache.PermissionNone, nil
}

// canRead checks: owner → note permission → folder permission (inherited) → team manager.
func (s *NoteService) canRead(ctx context.Context, requesterID string, note model.Note) (bool, error) {
	if requesterID == note.OwnerID {
		return true, nil
	}

	perm, err := s.lookupPermission(ctx, repository.AssetTypeNote, note.ID, requesterID)
	if err != nil {
		return false, err
	}
	if perm != "" && perm != cache.PermissionNone {
		return true, nil
	}

	// Folder permission satisfies note read access (inherited from ShareAsset propagation).
	folderPerm, err := s.lookupPermission(ctx, repository.AssetTypeFolder, note.FolderID, requesterID)
	if err != nil {
		return false, err
	}
	if folderPerm != "" && folderPerm != cache.PermissionNone {
		return true, nil
	}

	folder, err := s.getFolderCached(ctx, note.FolderID)
	if err != nil {
		return false, err
	}

	return s.teamMembers.IsManagerOf(requesterID, folder.OwnerID)
}

// canWrite checks: owner → note write permission → folder write permission (managers excluded, same as FolderService.canWrite).
func (s *NoteService) canWrite(ctx context.Context, requesterID string, note model.Note) (bool, error) {
	if requesterID == note.OwnerID {
		return true, nil
	}

	perm, err := s.lookupPermission(ctx, repository.AssetTypeNote, note.ID, requesterID)
	if err != nil {
		return false, err
	}
	if perm == "write" {
		return true, nil
	}
	if perm != cache.PermissionNone && perm != "" {
		// non-empty, non-none, non-write → read grant, insufficient for write
		return false, nil
	}

	folderPerm, err := s.lookupPermission(ctx, repository.AssetTypeFolder, note.FolderID, requesterID)
	if err != nil {
		return false, err
	}
	return folderPerm == "write", nil
}

func (s *NoteService) getFolderCached(ctx context.Context, folderID int) (model.Folder, error) {
	if cached, ok := s.assetCache.GetFolder(ctx, folderID); ok {
		return cached, nil
	}
	folder, err := s.folder.GetFolderByID(folderID)
	if err != nil {
		return model.Folder{}, err
	}
	_ = s.assetCache.SetFolder(ctx, folder)
	return folder, nil
}

func (s *NoteService) getNoteCached(ctx context.Context, noteID int) (model.Note, error) {
	if cached, ok := s.assetCache.GetNote(ctx, noteID); ok {
		return cached, nil
	}
	note, err := s.notes.GetNoteByID(noteID)
	if err != nil {
		return model.Note{}, err
	}
	_ = s.assetCache.SetNote(ctx, note)
	return note, nil
}

// CreateNote creates a note in a folder; only the folder owner may add notes (shared write access is insufficient).
func (s *NoteService) CreateNote(ctx context.Context, folderID int, ownerID, title, content string) (model.Note, error) {
	folder, err := s.getFolderCached(ctx, folderID)
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

	_ = s.assetCache.SetNote(ctx, note)

	s.publisher.PublishAssetEvent(ctx, event.NoteCreated, ownerID, map[string]any{
		"note_id":   note.ID,
		"folder_id": note.FolderID,
		"title":     note.Title,
	})
	return note, nil
}

func (s *NoteService) GetNote(ctx context.Context, noteID int, requesterID string) (model.Note, error) {
	note, err := s.getNoteCached(ctx, noteID)
	if err != nil {
		return model.Note{}, err
	}

	ok, err := s.canRead(ctx, requesterID, note)
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
	folder, err := s.getFolderCached(ctx, folderID)
	if err != nil {
		return nil, err
	}

	if requesterID == folder.OwnerID {
		return s.notes.ListNotesByFolder(folderID)
	}

	folderPerm, err := s.lookupPermission(ctx, repository.AssetTypeFolder, folderID, requesterID)
	if err != nil {
		return nil, err
	}
	if folderPerm != "" && folderPerm != cache.PermissionNone {
		return s.notes.ListNotesByFolder(folderID)
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
	note, err := s.getNoteCached(ctx, noteID)
	if err != nil {
		return model.Note{}, err
	}

	ok, err := s.canWrite(ctx, requesterID, note)
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

	_ = s.assetCache.SetNote(ctx, updated)

	s.publisher.PublishAssetEvent(ctx, event.NoteUpdated, requesterID, map[string]any{
		"note_id":   updated.ID,
		"folder_id": updated.FolderID,
		"title":     updated.Title,
		"owner_id":  updated.OwnerID,
	})
	return updated, nil
}

func (s *NoteService) DeleteNote(ctx context.Context, noteID int, requesterID string) error {
	note, err := s.getNoteCached(ctx, noteID)
	if err != nil {
		return err
	}

	ok, err := s.canWrite(ctx, requesterID, note)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}

	if err := s.notes.DeleteNote(noteID); err != nil {
		return err
	}

	_ = s.assetCache.Invalidate(ctx, repository.AssetTypeNote, noteID)

	s.publisher.PublishAssetEvent(ctx, event.NoteDeleted, requesterID, map[string]any{
		"note_id":   note.ID,
		"folder_id": note.FolderID,
		"title":     note.Title,
		"owner_id":  note.OwnerID,
	})
	return nil
}
