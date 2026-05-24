package service

import (
	"backend/internal/cache"
	"backend/internal/repository"
	"backend/package/dtorequest"
	"backend/package/event"
	"context"
	"errors"
	"fmt"
)

type SharingService struct {
	notes      *repository.NoteRepository
	folder     *repository.FolderRepository
	users      *repository.UserRepository
	permission *repository.PermissionRepository
	publisher  event.Publisher
	aclCache   *cache.ACL
}

func NewSharing(notes *repository.NoteRepository, folder *repository.FolderRepository, users *repository.UserRepository, permission *repository.PermissionRepository) *SharingService {
	return &SharingService{
		notes:      notes,
		folder:     folder,
		users:      users,
		permission: permission,
		publisher:  event.NewNoopPublisher(),
		aclCache:   cache.NewACL(cache.NewNoopCache()),
	}
}

func (s *SharingService) WithPublisher(p event.Publisher) *SharingService {
	if p != nil {
		s.publisher = p
	}
	return s
}

// WithCache injects the ACL cache so share/revoke can write-through. Defaults
// to noop in tests.
func (s *SharingService) WithCache(acl *cache.ACL) *SharingService {
	if acl != nil {
		s.aclCache = acl
	}
	return s
}

// ShareAsset grants a permission to a user (by email) on a note or folder; sharing a folder propagates the permission to its current notes.
func (s *SharingService) ShareAsset(req dtorequest.ShareAssetRequest, grantedBy string) error {
	user, err := s.users.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return errors.New("error: user to share with not found")
		}
		return err
	}

	var assetType string
	var assetID int

	if req.NoteID != 0 {
		assetType = repository.AssetTypeNote
		assetID = req.NoteID
		_, err := s.notes.GetNoteByID(assetID)
		if err != nil {
			return fmt.Errorf("note not found: %w", err)
		}
	} else if req.FolderID != 0 {
		assetType = repository.AssetTypeFolder
		assetID = req.FolderID
		_, err := s.folder.GetFolderByID(assetID)
		if err != nil {
			return fmt.Errorf("folder not found: %w", err)
		}
	} else {
		return errors.New("error: note_id or folder_id is required")
	}

	if err := s.permission.CreatePermission(assetType, assetID, user.UserID, req.PermissionType, grantedBy); err != nil {
		return err
	}

	ctx := context.Background()
	// Write-through: drop any cached "no permission" entry so the next ACL
	// check sees the fresh grant. The ACL TTL alone would eventually expire
	// the negative cache, but immediate invalidation keeps the share UX snappy.
	_ = s.aclCache.InvalidateUser(ctx, assetType, assetID, user.UserID)

	if assetType == repository.AssetTypeFolder { // propagate to contained notes
		notes, err := s.notes.ListNotesByFolder(assetID)
		if err != nil {
			return fmt.Errorf("list notes for folder inheritance: %w", err)
		}
		for _, note := range notes {
			err := s.permission.CreatePermission(repository.AssetTypeNote, note.ID, user.UserID, req.PermissionType, grantedBy)
			if err != nil && !errors.Is(err, repository.ErrPermissionAlreadyExists) {
				return fmt.Errorf("share note %d: %w", note.ID, err)
			}
			_ = s.aclCache.InvalidateUser(ctx, repository.AssetTypeNote, note.ID, user.UserID)
		}
	}

	eventType := event.NoteShared
	if assetType == repository.AssetTypeFolder {
		eventType = event.FolderShared
	}
	s.publisher.PublishAssetEvent(ctx, eventType, grantedBy, map[string]any{
		"asset_type":      assetType,
		"asset_id":        assetID,
		"granted_to":      user.UserID,
		"granted_email":   req.Email,
		"permission_type": req.PermissionType,
	})

	return nil
}

// RevokeAccess removes a permission from a user (by email); revoking a folder also cascades to its contained notes.
func (s *SharingService) RevokeAccess(req dtorequest.RevokeAccessRequest, revokedBy string) error {
	user, err := s.users.GetUserByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return errors.New("error: user not found")
		}
		return err
	}

	var assetType string
	var assetID int

	if req.NoteID != 0 {
		assetType = repository.AssetTypeNote
		assetID = req.NoteID
	} else if req.FolderID != 0 {
		assetType = repository.AssetTypeFolder
		assetID = req.FolderID
	} else {
		return errors.New("error: note_id or folder_id is required")
	}

	_ = revokedBy // reserved for audit logging

	if err := s.permission.RevokePermission(assetType, assetID, user.UserID); err != nil {
		return err
	}

	ctx := context.Background()
	_ = s.aclCache.InvalidateUser(ctx, assetType, assetID, user.UserID)

	if assetType == repository.AssetTypeFolder { // cascade: revoke inherited note permissions
		notes, err := s.notes.ListNotesByFolder(assetID)
		if err != nil {
			return fmt.Errorf("list notes for folder revocation: %w", err)
		}
		for _, note := range notes {
			err := s.permission.RevokePermission(repository.AssetTypeNote, note.ID, user.UserID)
			if err != nil && !errors.Is(err, repository.ErrPermissionNotFound) {
				return fmt.Errorf("revoke note %d: %w", note.ID, err)
			}
			_ = s.aclCache.InvalidateUser(ctx, repository.AssetTypeNote, note.ID, user.UserID)
		}
	}

	return nil
}
