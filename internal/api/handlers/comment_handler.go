package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ============================================
// Comment Handler
// ============================================

type CommentHandler struct {
	commentService service.CommentService
}

func (h *CommentHandler) ListByTask(c *gin.Context) {
	taskID := c.Param("id")

	comments, err := h.commentService.ListByTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch comments"})
		return
	}

	response := make([]models.CommentResponse, len(comments))
	for i, cm := range comments {
		response[i] = toCommentResponse(cm)
	}

	c.JSON(http.StatusOK, response)
}

func (h *CommentHandler) Create(c *gin.Context) {
	taskID := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.commentService.Create(c.Request.Context(), taskID, userID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, toCommentResponse(comment))
}

func (h *CommentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.commentService.Update(c.Request.Context(), id, userID, req.Content)
	if err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this comment"})
			return
		}
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update comment"})
		return
	}

	c.JSON(http.StatusOK, toCommentResponse(comment))
}

func (h *CommentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	if err := h.commentService.Delete(c.Request.Context(), id, userID); err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this comment"})
			return
		}
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
