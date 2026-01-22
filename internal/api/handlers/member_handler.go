package handlers

import (
	"log"
	"net/http"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/models"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	memberService service.MemberService
}

func NewMemberHandler(memberService service.MemberService) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
	}
}

// ListDirectMembers lists only direct members
func (h *MemberHandler) ListDirectMembers(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID := c.Param("entityId")

	members, err := h.memberService.ListDirectMembers(c.Request.Context(), entityType, entityID)
	if err != nil {
		if err == service.ErrInvalidEntityType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity type"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}

	response := make([]models.UnifiedMemberResponse, len(members))
	for i, m := range members {
		response[i] = toUnifiedMemberResponse(m)
	}

	c.JSON(http.StatusOK, response)
}

// ListEffectiveMembers lists direct + inherited members
func (h *MemberHandler) ListEffectiveMembers(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID := c.Param("entityId")

	members, err := h.memberService.ListEffectiveMembers(c.Request.Context(), entityType, entityID)
	if err != nil {
		if err == service.ErrInvalidEntityType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity type"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}

	response := make([]models.UnifiedMemberResponse, len(members))
	for i, m := range members {
		response[i] = toUnifiedMemberResponse(m)
	}

	c.JSON(http.StatusOK, response)
}

// AddMember adds a member by user ID
func (h *MemberHandler) AddMember(c *gin.Context) {
    entityType := c.Param("entityType")
    entityID := c.Param("entityId")
    inviterID, ok := middleware.RequireUserID(c)
    if !ok {
        return
    }

    var req models.AddMemberRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    err := h.memberService.AddMember(c.Request.Context(), entityType, entityID, req.UserID, req.Role, inviterID)
    if err != nil {
        // ✅ ADD THIS LOGGING
        log.Printf("[MemberHandler][AddMember] entityType=%s entityID=%s userID=%s inviterID=%s error=%v", 
            entityType, entityID, req.UserID, inviterID, err)
        
        if err == service.ErrUserNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
            return
        }
        if err == service.ErrConflict {
            c.JSON(http.StatusConflict, gin.H{"error": "User is already a member"})
            return
        }
        if err == service.ErrUnauthorized {  // ✅ ADD THIS CHECK
            c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to add members"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}

// InviteMemberByEmail invites by email
func (h *MemberHandler) InviteMemberByEmail(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID := c.Param("entityId")
	inviterID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.memberService.InviteMemberByEmail(c.Request.Context(), entityType, entityID, req.Email, req.Role, inviterID)
	if err != nil {
		// ✅ ADD LOGGING
		log.Printf("[MemberHandler][InviteMemberByEmail] entityType=%s entityID=%s email=%s inviterID=%s error=%v", 
			entityType, entityID, req.Email, inviterID, err)
		
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found"})
			return
		}
		if err == service.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "User is already a member"})
			return
		}
		// ✅ ADD THIS CHECK
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to add members"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to invite member"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}

// UpdateMemberRole updates a member's role
// func (h *MemberHandler) UpdateMemberRole(c *gin.Context) {
// 	entityType := c.Param("entityType")
// 	entityID := c.Param("entityId")
// 	userID := c.Param("userId")

// 	var req models.UpdateMemberRoleRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	err := h.memberService.UpdateMemberRole(c.Request.Context(), entityType, entityID, userID, req.Role)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
// }


func (h *MemberHandler) UpdateMemberRole(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID := c.Param("entityId")
	userID := c.Param("userId")
	
	// ✅ Get requester ID
	requesterID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	var req models.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ✅ Pass requesterID to service
	err := h.memberService.UpdateMemberRole(c.Request.Context(), entityType, entityID, userID, req.Role, requesterID)
	if err != nil {
		log.Printf("[MemberHandler][UpdateMemberRole] error=%v", err)
		
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this member's role"})
			return
		}
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
			return
		}
		if err == service.ErrLastOwner {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot demote the last owner"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

// RemoveMember removes a member
// func (h *MemberHandler) RemoveMember(c *gin.Context) {
// 	entityType := c.Param("entityType")
// 	entityID := c.Param("entityId")
// 	userID := c.Param("userId")

// 	err := h.memberService.RemoveMember(c.Request.Context(), entityType, entityID, userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
// 		return
// 	}

// 	c.JSON(http.StatusNoContent, nil)
// }



func (h *MemberHandler) RemoveMember(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID := c.Param("entityId")
	userID := c.Param("userId")
	
	// ✅ Get requester ID
	requesterID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	// ✅ Pass requesterID to service
	err := h.memberService.RemoveMember(c.Request.Context(), entityType, entityID, userID, requesterID)
	if err != nil {
		log.Printf("[MemberHandler][RemoveMember] error=%v", err)
		
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to remove this member"})
			return
		}
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
			return
		}
		if err == service.ErrLastOwner {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove the last owner"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// CheckAccess checks if user has access (direct or inherited)
func (h *MemberHandler) CheckAccess(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID := c.Param("entityId")
	userID := c.Query("userId")
	
	if userID == "" {
		// Check current user
		var ok bool
		userID, ok = middleware.RequireUserID(c)
		if !ok {
			return
		}
	}

	hasAccess, inheritedFrom, err := h.memberService.HasEffectiveAccess(c.Request.Context(), entityType, entityID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hasAccess":     hasAccess,
		"isDirect":      inheritedFrom == "",
		"inheritedFrom": inheritedFrom,
	})
}

// GetAccessLevel gets user's role and where it comes from
func (h *MemberHandler) GetAccessLevel(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID := c.Param("entityId")
	userID := c.Query("userId")
	
	if userID == "" {
		var ok bool
		userID, ok = middleware.RequireUserID(c)
		if !ok {
			return
		}
	}

	role, inheritedFrom, err := h.memberService.GetAccessLevel(c.Request.Context(), entityType, entityID, userID)
	if err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "No access"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access level"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"role":          role,
		"isDirect":      inheritedFrom == "",
		"inheritedFrom": inheritedFrom,
	})
}

// GetUserMemberships gets all entities user is member of
func (h *MemberHandler) GetUserMemberships(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	memberships, err := h.memberService.GetUserMemberships(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch memberships"})
		return
	}

	c.JSON(http.StatusOK, memberships)
}

// GetUserAllAccess gets comprehensive access map
func (h *MemberHandler) GetUserAllAccess(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	accessMap, err := h.memberService.GetUserAllAccess(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch access"})
		return
	}

	c.JSON(http.StatusOK, accessMap)
}

func toUnifiedMemberResponse(m *service.UnifiedMember) models.UnifiedMemberResponse {
	resp := models.UnifiedMemberResponse{
		ID:            m.ID,
		EntityType:    m.EntityType,
		EntityID:      m.EntityID,
		UserID:        m.UserID,
		Role:          m.Role,
		JoinedAt:      m.JoinedAt,
		IsInherited:   m.IsInherited,
		InheritedFrom: m.InheritedFrom,
	}
	if m.User != nil {
		resp.User = &models.UserResponse{
			ID:     m.User.ID,
			Email:  m.User.Email,
			Name:   m.User.Name,
			Avatar: m.User.Avatar,
			Status: m.User.Status,
		}
	}
	return resp
}



// Handlers for Accessible Entities
// ✅ REPLACE your existing GetAccessible* methods in member_handler.go with these:

// GetAccessibleWorkspaces returns all workspaces user can access
func (h *MemberHandler) GetAccessibleWorkspaces(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	workspaces, err := h.memberService.GetAccessibleWorkspaces(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accessible workspaces"})
		return
	}

	c.JSON(http.StatusOK, workspaces)
}

// GetAccessibleSpaces returns all spaces user can access (direct + inherited)
func (h *MemberHandler) GetAccessibleSpaces(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	spaces, err := h.memberService.GetAccessibleSpaces(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accessible spaces"})
		return
	}

	c.JSON(http.StatusOK, spaces)
}

// GetAccessibleFolders returns all folders user can access
func (h *MemberHandler) GetAccessibleFolders(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	folders, err := h.memberService.GetAccessibleFolders(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accessible folders"})
		return
	}

	c.JSON(http.StatusOK, folders)
}

// GetAccessibleProjects returns all projects user can access (direct + inherited)
func (h *MemberHandler) GetAccessibleProjects(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	projects, err := h.memberService.GetAccessibleProjects(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch accessible projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}


// ✅ NEW: GetVisibleSpaces returns spaces user can SEE (includes workspace members)
func (h *MemberHandler) GetVisibleSpaces(c *gin.Context) {
	userID, ok := middleware.RequireUserID(c)
	if !ok {
		return
	}

	spaces, err := h.memberService.GetVisibleSpaces(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch visible spaces"})
		return
	}

	c.JSON(http.StatusOK, spaces)
}

// ✅ NEW: GetAccessInfo returns detailed access information
func (h *MemberHandler) GetAccessInfo(c *gin.Context) {
	entityType := c.Param("entityType")
	entityID := c.Param("entityId")
	userID := c.Query("userId")
	
	if userID == "" {
		// Check current user
		var ok bool
		userID, ok = middleware.RequireUserID(c)
		if !ok {
			return
		}
	}

	accessInfo, err := h.memberService.GetAccessInfo(c.Request.Context(), entityType, entityID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check access"})
		return
	}

	c.JSON(http.StatusOK, accessInfo)
}