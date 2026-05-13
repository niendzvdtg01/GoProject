package handler

import (
	"backend/internal/repository"
	"backend/internal/service"
	"backend/package/dtorequest"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type NoteHandler struct {
	noteService *service.NoteService
}

func NewNoteHandler(noteService *service.NoteService) *NoteHandler {
	return &NoteHandler{noteService: noteService}
}

func (h *NoteHandler) CreateNote(c *gin.Context) {
	folderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder id"})
		return
	}

	var req dtorequest.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ownerID, _ := c.Get("user_id")

	note, err := h.noteService.CreateNote(c.Request.Context(), folderID, ownerID.(string), req.Title, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrFolderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create note"})
		}
		return
	}

	c.JSON(http.StatusCreated, note)
}

func (h *NoteHandler) ListNotes(c *gin.Context) {
	folderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid folder id"})
		return
	}

	requesterID, _ := c.Get("user_id")

	notes, err := h.noteService.ListFolderNotes(c.Request.Context(), folderID, requesterID.(string))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrFolderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "folder not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notes"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"notes": notes})
}

func (h *NoteHandler) GetNote(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	requesterID, _ := c.Get("user_id")

	note, err := h.noteService.GetNote(c.Request.Context(), id, requesterID.(string))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrNoteNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get note"})
		}
		return
	}

	c.JSON(http.StatusOK, note)
}

func (h *NoteHandler) UpdateNote(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	var req dtorequest.UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requesterID, _ := c.Get("user_id")

	note, err := h.noteService.UpdateNote(c.Request.Context(), id, requesterID.(string), req.Title, req.Content)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrNoteNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update note"})
		}
		return
	}

	c.JSON(http.StatusOK, note)
}

func (h *NoteHandler) DeleteNote(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	requesterID, _ := c.Get("user_id")

	if err := h.noteService.DeleteNote(c.Request.Context(), id, requesterID.(string)); err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		case errors.Is(err, repository.ErrNoteNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete note"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "note deleted"})
}
