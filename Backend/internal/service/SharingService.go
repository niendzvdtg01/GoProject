package service

import (
	"backend/internal/repository"
	"backend/package/dtorequest"
	"errors"
	"fmt"
)

type SharingService struct {
	notes      *repository.NoteRepository
	folder     *repository.FolderRepository
	users      *repository.UserRepository
	permission *repository.PermissionRepository
}

func NewSharing(notes *repository.NoteRepository, folder *repository.FolderRepository, users *repository.UserRepository, permission *repository.PermissionRepository) *SharingService {
	return &SharingService{
		notes:      notes,
		folder:     folder,
		users:      users,
		permission: permission,
	}
}

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

	return s.permission.CreatePermission(assetType, assetID, user.UserID, req.PermissionType, grantedBy)
}
