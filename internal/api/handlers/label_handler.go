package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ============================================
// Label Handler
// ============================================

type LabelHandler struct {
	labelService service.LabelService
}

func (h *LabelHandler) ListByProject(c *gin.Context) {
	projectID := c.Param("id")

	labels, err := h.labelService.ListByProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch labels"})
		return
	}

	response := make([]models.LabelResponse, len(labels))
	for i, l := range labels {
		response[i] = toLabelResponse(l)
	}

	c.JSON(http.StatusOK, response)
}

func (h *LabelHandler) Create(c *gin.Context) {
	projectID := c.Param("id")

	var req models.CreateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	label, err := h.labelService.Create(c.Request.Context(), projectID, req.Name, req.Color)
	if err != nil {
		if err == service.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Label with this name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create label"})
		return
	}

	c.JSON(http.StatusCreated, toLabelResponse(label))
}

func (h *LabelHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	label, err := h.labelService.Update(c.Request.Context(), id, req.Name, req.Color)
	if err != nil {
		if err == service.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "Label with this name already exists"})
			return
		}
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Label not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update label"})
		return
	}

	c.JSON(http.StatusOK, toLabelResponse(label))
}

func (h *LabelHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.labelService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete label"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
