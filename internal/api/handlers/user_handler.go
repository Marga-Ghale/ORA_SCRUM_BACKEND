package handlers

import (
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

// ============================================
// User Handler
// ============================================

type UserHandler struct {
	userService service.UserService
}

func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

func (h *UserHandler) UpdateCurrentUser(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Update(c.Request.Context(), userID, req.Name, req.Avatar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

// SearchUsers searches for users by email or name
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if len(query) < 2 {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}

	users, err := h.userService.Search(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}

	response := make([]models.UserResponse, len(users))
	for i, u := range users {
		response[i] = toUserResponse(u)
	}

	c.JSON(http.StatusOK, response)
}
