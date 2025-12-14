package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ============================================
// Space Handler
// ============================================

type SpaceHandler struct {
	spaceService service.SpaceService
}

func (h *SpaceHandler) ListByWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")

	spaces, err := h.spaceService.ListByWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch spaces"})
		return
	}

	response := make([]models.SpaceResponse, len(spaces))
	for i, s := range spaces {
		response[i] = toSpaceResponse(s)
	}

	c.JSON(http.StatusOK, response)
}

func (h *SpaceHandler) Create(c *gin.Context) {
	workspaceID := c.Param("id")

	var req models.CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	space, err := h.spaceService.Create(c.Request.Context(), workspaceID, req.Name, req.Description, req.Icon, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create space"})
		return
	}

	c.JSON(http.StatusCreated, toSpaceResponse(space))
}

func (h *SpaceHandler) Get(c *gin.Context) {
	id := c.Param("id")

	space, err := h.spaceService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Space not found"})
		return
	}

	c.JSON(http.StatusOK, toSpaceResponse(space))
}

func (h *SpaceHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	space, err := h.spaceService.Update(c.Request.Context(), id, req.Name, req.Description, req.Icon, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update space"})
		return
	}

	c.JSON(http.StatusOK, toSpaceResponse(space))
}

func (h *SpaceHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.spaceService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete space"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
