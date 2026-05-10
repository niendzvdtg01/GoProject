package service

import (
	"backend/internal/respository"
	"backend/package/dtorequest"
	"errors"
	"fmt"
	"strconv"
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

func (s *SharingService) ShareNote(req dtorequest.ShareNoteRequest, ownerUserID string) error {
	noteID, err := strconv.Atoi(req.NoteID)
	if err != nil {
		return fmt.Errorf("invalid note id: %w", err)
	}

	note, err := s.notes.GetNoteByID(noteID)
	if err != nil {
		return err
	}

	folderID, err := strconv.Atoi(note.FolderID)
	if err != nil {
		return fmt.Errorf("invalid note folder id: %w", err)
	}

	folder, err := s.folder.GetFolderByID(folderID)
	if err != nil {
		return err
	}
	if folder.OwnerID != ownerUserID {
		return errors.New("only the folder owner can share the note")
	}

	if _, err := s.users.GetUserByID(req.UserID); err != nil {
		if errors.Is(err, respository.ErrUserNotFound) {
			return errors.New("target user not found")
		}
		return fmt.Errorf("get share target: %w", err)
	}

	return s.permission.CreatePermission(respository.AssetTypeNote, req.NoteID, req.UserID, req.PermissionType, ownerUserID)
}
