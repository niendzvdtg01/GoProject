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

var ErrForbidden = errors.New("forbidden: insufficient permissions")

type FolderService struct {
	folder      *repository.FolderRepository
	users       *repository.UserRepository
	permission  *repository.PermissionRepository
	teamMembers *repository.TeamMemberRepository
	publisher   event.Publisher
	assetCache  *cache.AssetMetadata
	aclCache    *cache.ACL
}

func NewFolderService(folder *repository.FolderRepository, users *repository.UserRepository, permission *repository.PermissionRepository, teamMembers *repository.TeamMemberRepository) *FolderService {
	noop := cache.NewNoopCache()
	return &FolderService{
		folder:      folder,
		users:       users,
		permission:  permission,
		teamMembers: teamMembers,
		publisher:   event.NewNoopPublisher(),
		assetCache:  cache.NewAssetMetadata(noop),
		aclCache:    cache.NewACL(noop),
	}
}

func (s *FolderService) WithPublisher(p event.Publisher) *FolderService {
	if p != nil {
		s.publisher = p
	}
	return s
}

// WithCache injects the metadata + ACL caches. Defaults to noop so unit tests
// and cache-less deployments keep working unchanged.
func (s *FolderService) WithCache(assets *cache.AssetMetadata, acl *cache.ACL) *FolderService {
	if assets != nil {
		s.assetCache = assets
	}
	if acl != nil {
		s.aclCache = acl
	}
	return s
}

// canRead checks: owner -> explicit permission (cached) -> team manager (read-only oversight).
func (s *FolderService) canRead(ctx context.Context, requesterID string, folder model.Folder) (bool, error) {
	if requesterID == folder.OwnerID {
		return true, nil
	}

	perm, err := s.lookupPermission(ctx, repository.AssetTypeFolder, folder.ID, requesterID)
	if err != nil {
		return false, err
	}
	if perm != "" && perm != cache.PermissionNone {
		return true, nil
	}

	isManager, err := s.teamMembers.IsManagerOf(requesterID, folder.OwnerID)
	if err != nil {
		return false, err
	}
	return isManager, nil
}

// canWrite checks: owner -> explicit write permission (cached). Read grant is
// insufficient and managers have no write access.
func (s *FolderService) canWrite(ctx context.Context, requesterID string, folder model.Folder) (bool, error) {
	if requesterID == folder.OwnerID {
		return true, nil
	}

	perm, err := s.lookupPermission(ctx, repository.AssetTypeFolder, folder.ID, requesterID)
	if err != nil {
		return false, err
	}
	return perm == "write", nil
}

// lookupPermission resolves the user's permission_type through the ACL cache,
// falling back to the DB on miss. The DB miss is cached as PermissionNone so
// repeated unauthorised probes don't keep hitting the database.
func (s *FolderService) lookupPermission(ctx context.Context, assetType string, assetID int, userID string) (string, error) {
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

func (s *FolderService) CreateFolder(ctx context.Context, ownerID, name string) (model.Folder, error) {
	id, err := s.folder.CreateFolder(ownerID, name)
	if err != nil {
		return model.Folder{}, fmt.Errorf("create folder: %w", err)
	}

	folder, err := s.folder.GetFolderByID(id)
	if err != nil {
		return model.Folder{}, err
	}

	_ = s.assetCache.SetFolder(ctx, folder)

	s.publisher.PublishAssetEvent(ctx, event.FolderCreated, ownerID, map[string]any{
		"folder_id": folder.ID,
		"name":      folder.Name,
	})
	return folder, nil
}

// GetFolder reads metadata through the cache before falling back to the DB.
// The ACL check runs after we have the folder so we can authorise against the
// owner without a second query.
func (s *FolderService) GetFolder(ctx context.Context, folderID int, requesterID string) (model.Folder, error) {
	folder, err := s.getFolderCached(ctx, folderID)
	if err != nil {
		return model.Folder{}, err
	}

	ok, err := s.canRead(ctx, requesterID, folder)
	if err != nil {
		return model.Folder{}, err
	}
	if !ok {
		return model.Folder{}, ErrForbidden
	}

	return folder, nil
}

func (s *FolderService) getFolderCached(ctx context.Context, folderID int) (model.Folder, error) {
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

func (s *FolderService) ListUserFolders(ctx context.Context, userID string) ([]model.Folder, error) {
	return s.folder.ListFoldersByOwner(userID)
}

func (s *FolderService) UpdateFolder(ctx context.Context, folderID int, requesterID, name string) (model.Folder, error) {
	folder, err := s.getFolderCached(ctx, folderID)
	if err != nil {
		return model.Folder{}, err
	}

	ok, err := s.canWrite(ctx, requesterID, folder)
	if err != nil {
		return model.Folder{}, err
	}
	if !ok {
		return model.Folder{}, ErrForbidden
	}

	updated, err := s.folder.UpdateFolder(folderID, name)
	if err != nil {
		return model.Folder{}, err
	}

	_ = s.assetCache.SetFolder(ctx, updated)

	s.publisher.PublishAssetEvent(ctx, event.FolderUpdated, requesterID, map[string]any{
		"folder_id": updated.ID,
		"name":      updated.Name,
		"owner_id":  updated.OwnerID,
	})
	return updated, nil
}

func (s *FolderService) DeleteFolder(ctx context.Context, folderID int, requesterID string) error {
	folder, err := s.getFolderCached(ctx, folderID)
	if err != nil {
		return err
	}

	ok, err := s.canWrite(ctx, requesterID, folder)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}

	if err := s.folder.DeleteFolder(folderID); err != nil {
		return err
	}

	_ = s.assetCache.Invalidate(ctx, repository.AssetTypeFolder, folderID)

	s.publisher.PublishAssetEvent(ctx, event.FolderDeleted, requesterID, map[string]any{
		"folder_id": folder.ID,
		"name":      folder.Name,
		"owner_id":  folder.OwnerID,
	})
	return nil
}
