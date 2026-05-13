package service

import (
	"backend/internal/model"
	"backend/internal/repository"
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
}

func NewFolderService(folder *repository.FolderRepository, users *repository.UserRepository, permission *repository.PermissionRepository, teamMembers *repository.TeamMemberRepository) *FolderService {
	return &FolderService{
		folder:      folder,
		users:       users,
		permission:  permission,
		teamMembers: teamMembers,
	}
}

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

	return s.folder.GetFolderByID(id)
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

	return s.folder.UpdateFolder(folderID, name)
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

	return s.folder.DeleteFolder(folderID)
}
