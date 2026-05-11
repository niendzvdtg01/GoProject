package service

import (
	"backend/internal/respository"
	"backend/package/dtorequest"
	"database/sql"
	"errors"
)

const (
	readPermission  = "read"
	writePermission = "write"
)

type SharingService struct {
	notes      *respository.NoteRepository
	folder     *respository.FolderRepository
	users      *respository.UserRepository
	permission *respository.PermissionRepository
}

func NewSharing(notes *respository.NoteRepository, folder *respository.FolderRepository, users *respository.UserRepository, permission *respository.PermissionRepository) *SharingService {
	return &SharingService{
		notes:      notes,
		folder:     folder,
		users:      users,
		permission: permission,
	}
}

func (s *SharingService) shareAsset(
	assetType string,
	assetID string,
	targetUserID string,
	permissionType string,
	ownerUserID string,
) error {

	if _, err := s.users.GetUserByID(targetUserID); err != nil {

		if errors.Is(err, respository.ErrUserNotFound) {
			return errors.New("target user not found")
		}

		return err
	}

	return s.permission.CreatePermission(
		assetType,
		assetID,
		targetUserID,
		permissionType,
		ownerUserID,
	)
}

func (s *SharingService) ShareNote(request dtorequest.ShareAssetRequest, ownerUserID string) error {

	note, err := s.notes.GetNoteByID(request.AssetID)
	if err != nil {
		return err
	}

	folder, err := s.folder.GetFolderByID(note.FolderID)
	if err != nil {
		return err
	}

	if folder.OwnerID != ownerUserID {
		return errors.New("only the folder owner can share notes in this folder")
	}
	return s.shareAsset(
		respository.AssetTypeNote,
		request.AssetID,
		request.UserID,
		request.PermissionType,
		ownerUserID,
	)
}

func (s *SharingService) ShareFolder(request dtorequest.ShareAssetRequest, ownerUserID string) error {

	folder, err := s.folder.GetFolderByID(request.AssetID)
	if err != nil {
		return err
	}

	if folder.OwnerID != ownerUserID {
		return errors.New("only the folder owner can share this folder")
	}
	return s.shareAsset(
		respository.AssetTypeFolder,
		request.AssetID,
		request.UserID,
		request.PermissionType,
		ownerUserID,
	)
}

func (s *SharingService) ReadNotePermission(noteID string, userID string) (bool, error) {
	folder, folderErr := s.folder.GetFolderByNoteID(noteID)
	if folderErr != nil {
		return false, folderErr
	}

	if folder.OwnerID == userID {
		return true, nil
	}

	permission, err := s.permission.GetPermissionByAssetAndUser(
		respository.AssetTypeNote,
		noteID,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return permission.PermissionType == readPermission || permission.PermissionType == writePermission, nil
}

func (s *SharingService) WriteNotePermission(noteID string, userID string) (bool, error) {
	folder, folderErr := s.folder.GetFolderByNoteID(noteID)
	if folderErr != nil {
		return false, folderErr
	}

	if folder.OwnerID == userID {
		return true, nil
	}

	permission, err := s.permission.GetPermissionByAssetAndUser(
		respository.AssetTypeNote,
		noteID,
		userID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return permission.PermissionType == writePermission, nil
}
