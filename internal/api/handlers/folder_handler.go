package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type FolderHandler struct {
	folderService service.FolderService
}

func NewFolderHandler(folderService service.FolderService) *FolderHandler {
	return &FolderHandler{
		folderService: folderService,
	}
}

// ListBySpace - List folders in a specific space
// GET /spaces/:id/folders
func (h *FolderHandler) ListBySpace(c *gin.Context) {
	spaceID := c.Param("id") // Changed from "spaceId" to "id"

	folders, err := h.folderService.ListBySpace(c.Request.Context(), spaceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch folders"})
		return
	}

	response := make([]models.FolderResponse, len(folders))
	for i, folder := range folders {
		response[i] = toFolderResponse(folder)
	}

	c.JSON(http.StatusOK, response)
}

// ListByUser - List folders accessible by a user
// GET /folders/my
func (h *FolderHandler) ListByUser(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	folders, err := h.folderService.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch folders"})
		return
	}

	response := make([]models.FolderResponse, len(folders))
	for i, folder := range folders {
		response[i] = toFolderResponse(folder)
	}

	c.JSON(http.StatusOK, response)
}

// Create - Create a new folder in a space
// POST /spaces/:id/folders
func (h *FolderHandler) Create(c *gin.Context) {
	// Get spaceID from URL parameter
	spaceID := c.Param("id") // Changed from "spaceId" to "id"
	if spaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "spaceId is required"})
		return
	}

	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.folderService.Create(
		c.Request.Context(),
		spaceID,         // spaceID from URL parameter
		userID,          // creatorID from auth
		req.Name,        // name from body
		req.Description, // description from body
		req.Icon,        // icon from body
		req.Color,       // color from body
	)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Space not found"})
			return
		}
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to create folder in this space"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, toFolderResponse(folder))
}

// Get - Get a folder by ID
// GET /folders/:id
func (h *FolderHandler) Get(c *gin.Context) {
	id := c.Param("id")

	folder, err := h.folderService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch folder"})
		return
	}

	c.JSON(http.StatusOK, toFolderResponse(folder))
}

// Update - Update a folder
// PUT /folders/:id
func (h *FolderHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.folderService.Update(
		c.Request.Context(),
		id,
		req.Name,
		req.Description,
		req.Icon,
		req.Color,
		req.Visibility,
		req.AllowedUsers,
		req.AllowedTeams,
	)
	if err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update folder"})
		return
	}

	c.JSON(http.StatusOK, toFolderResponse(folder))
}

// Delete - Delete a folder
// DELETE /folders/:id
func (h *FolderHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.folderService.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// UpdateVisibility - Update folder visibility settings
// PATCH /folders/:id/visibility
func (h *FolderHandler) UpdateVisibility(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Visibility   string   `json:"visibility" binding:"required"`
		AllowedUsers []string `json:"allowedUsers"`
		AllowedTeams []string `json:"allowedTeams"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.folderService.UpdateVisibility(
		c.Request.Context(),
		id,
		req.Visibility,
		req.AllowedUsers,
		req.AllowedTeams,
	); err != nil {
		if err == service.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update visibility"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Visibility updated successfully"})
}
