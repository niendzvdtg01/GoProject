package service

import (
	"backend/internal/model"
	"backend/internal/respository"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type FolderService struct {
	folder *respository.FolderRepository
	users  *respository.UserRepository
}

func NewFolderService(folder *respository.FolderRepository, users *respository.UserRepository) *FolderService {
	return &FolderService{
		folder: folder,
		users:  users,
	}
}

func (fs *FolderService) CreateFolder(ownerName, folderName string) (model.Folder, error) {
	folderName = strings.TrimSpace(folderName)
	if folderName == "" {
		return model.Folder{}, errors.New("folder name is required")
	}

	owner, ownerErr := fs.users.GetUserByUsername(ownerName)
	if ownerErr != nil {
		return model.Folder{}, fmt.Errorf("get folder owner: %w", ownerErr)
	}

	folderID, folderErr := fs.folder.CreateFolder(owner.UserID, folderName)
	if folderErr != nil {
		return model.Folder{}, fmt.Errorf("create folder: %w", folderErr)
	}

	folderIDStr := strconv.Itoa(folderID)
	return model.Folder{
		ID:      folderIDStr,
		OwnerID: owner.UserID,
		Name:    folderName,
	}, nil
}

func (fs *FolderService) GetFolderByID(folderID string) (model.Folder, error) {
	folder, err := fs.folder.GetFolderByID(folderID)
	if err != nil {
		return model.Folder{}, err
	}
	return folder, nil
}

func (fs *FolderService) ListFoldersByOwner(ownerName string) ([]model.Folder, error) {
	owner, err := fs.users.GetUserByUsername(ownerName)
	if err != nil {
		return nil, fmt.Errorf("get folder owner: %w", err)
	}

	return fs.folder.ListFoldersByOwner(owner.UserID)
}

func (fs *FolderService) UpdateFolder(folderID string, ownerName, newName string) (model.Folder, error) {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return model.Folder{}, errors.New("folder name is required")
	}

	owner, err := fs.users.GetUserByUsername(ownerName)
	if err != nil {
		return model.Folder{}, fmt.Errorf("get folder owner: %w", err)
	}

	current, err := fs.folder.GetFolderByID(folderID)
	if err != nil {
		return model.Folder{}, err
	}
	if current.OwnerID != owner.UserID {
		return model.Folder{}, errors.New("only the folder owner can update the folder")
	}

	updated, err := fs.folder.UpdateFolder(folderID, newName)
	if err != nil {
		return model.Folder{}, err
	}
	return updated, nil
}

func (fs *FolderService) DeleteFolder(folderID string, ownerName string) error {
	owner, err := fs.users.GetUserByUsername(ownerName)
	if err != nil {
		return fmt.Errorf("get folder owner: %w", err)
	}

	current, err := fs.folder.GetFolderByID(folderID)
	if err != nil {
		return err
	}
	if current.OwnerID != owner.UserID {
		return errors.New("only the folder owner can delete the folder")
	}

	return fs.folder.DeleteFolder(folderID)
}
