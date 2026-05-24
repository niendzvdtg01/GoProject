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

var ErrForbidden = errors.New("forbidden: insufficient permissions")

type FolderService struct {
	folder      *repository.FolderRepository
	users       *repository.UserRepository
	permission  *repository.PermissionRepository
	teamMembers *repository.TeamMemberRepository
	publisher   event.Publisher
}

func NewFolderService(folder *repository.FolderRepository, users *repository.UserRepository, permission *repository.PermissionRepository, teamMembers *repository.TeamMemberRepository) *FolderService {
	return &FolderService{
		folder:      folder,
		users:       users,
		permission:  permission,
		teamMembers: teamMembers,
		publisher:   event.NewNoopPublisher(),
	}
}

func (s *FolderService) WithPublisher(p event.Publisher) *FolderService {
	if p != nil {
		s.publisher = p
	}
	return s
}

// canRead checks: owner -> explicit permission -> team manager (read-only oversight).
func (s *FolderService) canRead(requesterID string, folder model.Folder) (bool, error) {
	if requesterID == folder.OwnerID {
		return true, nil
	}

	perm, err := s.permission.GetPermissionByAssetAndUser(repository.AssetTypeFolder, folder.ID, requesterID)
	if err == nil {
		_ = perm
		return true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	isManager, err := s.teamMembers.IsManagerOf(requesterID, folder.OwnerID)
	if err != nil {
		return false, err
	}
	return isManager, nil
}

// canWrite checks: owner -> explicit write permission (read grant is insufficient; managers have no write access).
func (s *FolderService) canWrite(requesterID string, folder model.Folder) (bool, error) {
	if requesterID == folder.OwnerID {
		return true, nil
	}

	perm, err := s.permission.GetPermissionByAssetAndUser(repository.AssetTypeFolder, folder.ID, requesterID)
	if err == nil {
		return perm.PermissionType == "write", nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	return false, nil
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

	s.publisher.PublishAssetEvent(ctx, event.FolderCreated, ownerID, map[string]any{
		"folder_id": folder.ID,
		"name":      folder.Name,
	})
	return folder, nil
}

func (s *FolderService) GetFolder(ctx context.Context, folderID int, requesterID string) (model.Folder, error) {
	folder, err := s.folder.GetFolderByID(folderID)
	if err != nil {
		return model.Folder{}, err
	}

	ok, err := s.canRead(requesterID, folder)
	if err != nil {
		return model.Folder{}, err
	}
	if !ok {
		return model.Folder{}, ErrForbidden
	}

	return folder, nil
}

func (s *FolderService) ListUserFolders(ctx context.Context, userID string) ([]model.Folder, error) {
	return s.folder.ListFoldersByOwner(userID)
}

func (s *FolderService) UpdateFolder(ctx context.Context, folderID int, requesterID, name string) (model.Folder, error) {
	folder, err := s.folder.GetFolderByID(folderID)
	if err != nil {
		return model.Folder{}, err
	}

	ok, err := s.canWrite(requesterID, folder)
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

	s.publisher.PublishAssetEvent(ctx, event.FolderUpdated, requesterID, map[string]any{
		"folder_id": updated.ID,
		"name":      updated.Name,
		"owner_id":  updated.OwnerID,
	})
	return updated, nil
}

func (s *FolderService) DeleteFolder(ctx context.Context, folderID int, requesterID string) error {
	folder, err := s.folder.GetFolderByID(folderID)
	if err != nil {
		return err
	}

	ok, err := s.canWrite(requesterID, folder)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}

	if err := s.folder.DeleteFolder(folderID); err != nil {
		return err
	}

	s.publisher.PublishAssetEvent(ctx, event.FolderDeleted, requesterID, map[string]any{
		"folder_id": folder.ID,
		"name":      folder.Name,
		"owner_id":  folder.OwnerID,
	})
	return nil
}
