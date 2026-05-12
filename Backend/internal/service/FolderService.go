package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"fmt"
)

type FolderService struct {
	folder *repository.FolderRepository
	users  *repository.UserRepository
}

func NewFolderService(folder *repository.FolderRepository, users *repository.UserRepository) *FolderService {
	return &FolderService{
		folder: folder,
		users:  users,
	}
}

func (s *FolderService) CreateFolder(ctx context.Context, ownerID, name string) (model.Folder, error) {
	id, err := s.folder.CreateFolder(ownerID, name)
	if err != nil {
		return model.Folder{}, fmt.Errorf("create folder: %w", err)
	}

	return s.folder.GetFolderByID(id)
}

func (s *FolderService) GetFolder(ctx context.Context, folderID int) (model.Folder, error) {
	return s.folder.GetFolderByID(folderID)
}

func (s *FolderService) ListUserFolders(ctx context.Context, userID string) ([]model.Folder, error) {
	return s.folder.ListFoldersByOwner(userID)
}
