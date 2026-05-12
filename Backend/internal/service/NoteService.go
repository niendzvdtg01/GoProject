package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"fmt"
)

type NoteService struct {
	notes  *repository.NoteRepository
	folder *repository.FolderRepository
	users  *repository.UserRepository
}

func NewNoteService(notes *repository.NoteRepository, folder *repository.FolderRepository, users *repository.UserRepository) *NoteService {
	return &NoteService{
		notes:  notes,
		folder: folder,
		users:  users,
	}
}

func (s *NoteService) CreateNote(ctx context.Context, folderID int, ownerID, title, content string) (model.Note, error) {
	// Verify folder exists
	_, err := s.folder.GetFolderByID(folderID)
	if err != nil {
		return model.Note{}, fmt.Errorf("folder not found: %w", err)
	}

	id, err := s.notes.CreateNote(folderID, ownerID, title, content)
	if err != nil {
		return model.Note{}, fmt.Errorf("create note: %w", err)
	}

	return s.notes.GetNoteByID(id)
}

func (s *NoteService) GetNote(ctx context.Context, noteID int) (model.Note, error) {
	return s.notes.GetNoteByID(noteID)
}

func (s *NoteService) ListFolderNotes(ctx context.Context, folderID int) ([]model.Note, error) {
	return s.notes.ListNotesByFolder(folderID)
}

func (s *NoteService) UpdateNote(ctx context.Context, noteID int, title, content string) (model.Note, error) {
	return s.notes.UpdateNote(noteID, title, content)
}

func (s *NoteService) DeleteNote(ctx context.Context, noteID int) error {
	return s.notes.DeleteNote(noteID)
}
