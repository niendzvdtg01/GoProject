package service

import (
	"backend/internal/model"
	"backend/internal/respository"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type NoteService struct {
	notes  *respository.NoteRepository
	folder *respository.FolderRepository
	users  *respository.UserRepository
}

func NewNoteService(notes *respository.NoteRepository, folder *respository.FolderRepository, users *respository.UserRepository) *NoteService {
	return &NoteService{
		notes:  notes,
		folder: folder,
		users:  users,
	}
}

func (ns *NoteService) CreateNote(ownerName, folderID string, title, content string) (model.Notes, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" {
		return model.Notes{}, errors.New("note title is required")
	}

	owner, err := ns.users.GetUserByUsername(ownerName)
	if err != nil {
		return model.Notes{}, fmt.Errorf("get note owner: %w", err)
	}

	folderInt, err := strconv.Atoi(folderID)
	if err != nil {
		return model.Notes{}, fmt.Errorf("invalid folder id: %w", err)
	}

	folderRecord, err := ns.folder.GetFolderByID(folderInt)
	if err != nil {
		return model.Notes{}, err
	}
	if folderRecord.OwnerID != owner.UserID {
		return model.Notes{}, errors.New("only the folder owner can create notes in this folder")
	}

	noteID, err := ns.notes.CreateNote(folderID, title, content)
	if err != nil {
		return model.Notes{}, fmt.Errorf("create note: %w", err)
	}

	return model.Notes{
		ID:       strconv.Itoa(noteID),
		FolderID: folderID,
		Title:    title,
		Content:  content,
	}, nil
}

func (ns *NoteService) GetNoteByID(noteID int) (model.Notes, error) {
	return ns.notes.GetNoteByID(noteID)
}

func (ns *NoteService) ListNotesByFolder(folderID int) ([]model.Notes, error) {
	return ns.notes.ListNotesByFolder(folderID)
}

func (ns *NoteService) UpdateNote(noteID int, ownerName, title, content string) (model.Notes, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" {
		return model.Notes{}, errors.New("note title is required")
	}

	updated, err := ns.notes.UpdateNote(noteID, title, content)
	if err != nil {
		return model.Notes{}, err
	}
	return updated, nil
}

func (ns *NoteService) DeleteNote(noteID int, ownerName string) error {
	owner, err := ns.users.GetUserByUsername(ownerName)
	if err != nil {
		return fmt.Errorf("get note owner: %w", err)
	}
	folder, err := ns.folder.GetFolderByNoteID(noteID)
	if err != nil {
		return fmt.Errorf("get note folder: %w", err)
	}
	if folder.OwnerID != owner.UserID {
		return errors.New("only the folder owner can delete notes in this folder")
	}
	return ns.notes.DeleteNote(noteID)
}
