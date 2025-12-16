package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/api/middleware"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/service"
)

// InvitationHandler exposes HTTP endpoints for invitation flows.
type InvitationHandler struct {
	svc service.InvitationService
}


type createWorkspaceInvitationReq struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=ADMIN MEMBER VIEWER"`
}

type createProjectInvitationReq struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required,oneof=LEAD MEMBER VIEWER"`
}

type acceptInvitationReq struct {
	Token string `json:"token" binding:"required"`
}

func NewInvitationHandler(svc service.InvitationService) *InvitationHandler {
	return &InvitationHandler{svc: svc}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// DTOs

type createInvitationReq struct {
	WorkspaceID   string                      `json:"workspace_id"`
	Email         string                      `json:"email"`
	Type          repository.InvitationType  `json:"type,omitempty"`
	TargetID      string                      `json:"target_id,omitempty"`
	TargetName    string                      `json:"target_name,omitempty"`
	Role          repository.WorkspaceRole    `json:"role,omitempty"`
	Permission    repository.PermissionLevel  `json:"permission,omitempty"`
	InvitedByID   string                      `json:"invited_by_id,omitempty"`
	InvitedByName string                      `json:"invited_by_name,omitempty"`
	Message       *string                     `json:"message,omitempty"`
	ExpiresInDays *int                        `json:"expires_in_days,omitempty"`
	MaxUses       *int                        `json:"max_uses,omitempty"`
	Method        repository.InvitationMethod `json:"method,omitempty"`
}

type acceptReq struct {
	UserID string `json:"user_id,omitempty"`
}

type createLinkReq struct {
	WorkspaceID       string                     `json:"workspace_id"`
	Type              repository.InvitationType `json:"type,omitempty"`
	TargetID          string                     `json:"target_id,omitempty"`
	DefaultRole       repository.WorkspaceRole   `json:"default_role,omitempty"`
	DefaultPermission repository.PermissionLevel `json:"default_permission,omitempty"`
	IsActive          *bool                      `json:"is_active,omitempty"`
	RequiresApproval  *bool                      `json:"requires_approval,omitempty"`
	AllowedDomains    []string                   `json:"allowed_domains,omitempty"`
	BlockedDomains    []string                   `json:"blocked_domains,omitempty"`
	MaxUses           *int                       `json:"max_uses,omitempty"`
	ExpiresInDays     *int                       `json:"expires_in_days,omitempty"`
	CreatedByID       string                     `json:"created_by_id,omitempty"`
}

type accessRequestReq struct {
	WorkspaceID string                     `json:"workspace_id"`
	RequesterID string                     `json:"requester_id"`
	Email       string                     `json:"email"`
	Type        repository.InvitationType `json:"type"`
	TargetID    string                     `json:"target_id"`
	Message     *string                    `json:"message,omitempty"`
}

// Handlers

// POST /api/invitations
func (h *InvitationHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	var req createInvitationReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	inv := &repository.Invitation{
		WorkspaceID:   req.WorkspaceID,
		Email:         req.Email,
		Type:          req.Type,
		TargetID:      req.TargetID,
		TargetName:    req.TargetName,
		Role:          req.Role,
		Permission:    req.Permission,
		InvitedByID:   req.InvitedByID,
		InvitedByName: req.InvitedByName,
		Message:       req.Message,
		Method:        req.Method,
		MaxUses:       req.MaxUses,
	}
	if req.ExpiresInDays != nil {
		t := time.Now().Add(time.Duration(*req.ExpiresInDays) * 24 * time.Hour)
		inv.ExpiresAt = &t
	}
	if err := h.svc.CreateInvitation(r.Context(), inv); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, inv)
}

// POST /api/invitations/with-permissions
func (h *InvitationHandler) CreateWithPermissions(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Invitation  *repository.Invitation           `json:"invitation"`
		Permissions *repository.InvitationPermissions `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if payload.Invitation == nil {
		writeError(w, http.StatusBadRequest, "invitation required")
		return
	}
	if err := h.svc.CreateWithPermissions(r.Context(), payload.Invitation, payload.Permissions); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, payload.Invitation)
}

// GET /api/invitations/{id}
func (h *InvitationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	inv, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if inv == nil {
		writeError(w, http.StatusNotFound, "invitation not found")
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

// GET /api/invitations/token/{token}
func (h *InvitationHandler) GetByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	inv, err := h.svc.GetByToken(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if inv == nil {
		writeError(w, http.StatusNotFound, "invitation not found")
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

// POST /api/invitations/{id}/accept
func (h *InvitationHandler) AcceptByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req acceptReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	inv, err := h.svc.AcceptByID(r.Context(), id, req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if inv == nil {
		writeError(w, http.StatusNotFound, "invitation not found")
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

// POST /api/invitations/token/{token}/accept
func (h *InvitationHandler) AcceptByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	var req acceptReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	inv, err := h.svc.AcceptByToken(r.Context(), token, req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if inv == nil {
		writeError(w, http.StatusNotFound, "invitation not found")
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

// POST /api/invitations/{id}/decline
func (h *InvitationHandler) DeclineByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	inv, err := h.svcDeclineByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if inv == nil {
		writeError(w, http.StatusNotFound, "invitation not found")
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

// wrapper to handle potential naming collision
func (h *InvitationHandler) svcDeclineByID(ctx context.Context, id string) (*repository.Invitation, error) {
	return h.svc.DeclineByID(ctx, id)
}

// POST /api/invitations/{id}/resend
func (h *InvitationHandler) ResendByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	actor := r.Header.Get("X-Actor-Id")
	var actorPtr *string
	if actor != "" {
		actorPtr = &actor
	}
	inv, err := h.svc.ResendInvitation(r.Context(), id, actorPtr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if inv == nil {
		writeError(w, http.StatusNotFound, "invitation not found")
		return
	}
	writeJSON(w, http.StatusOK, inv)
}

// GET /api/workspaces/{workspace_id}/invitations?limit=&offset=
func (h *InvitationHandler) ListByWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspace_id")
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	invs, total, err := h.svc.ListByWorkspace(r.Context(), workspaceID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := map[string]interface{}{
		"count": len(invs),
		"total": total,
		"data":  invs,
	}
	writeJSON(w, http.StatusOK, resp)
}

// CreateLinkSettings POST /api/invitation-links
func (h *InvitationHandler) CreateLinkSettings(w http.ResponseWriter, r *http.Request) {
	var req createLinkReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	settings := &repository.InvitationLinkSettings{
		WorkspaceID:       req.WorkspaceID,
		Type:              req.Type,
		TargetID:          req.TargetID,
		DefaultRole:       req.DefaultRole,
		DefaultPermission: req.DefaultPermission,
		IsActive:          true,
		RequiresApproval:  false,
		CreatedByID:       req.CreatedByID,
	}
	if req.IsActive != nil {
		settings.IsActive = *req.IsActive
	}
	if req.RequiresApproval != nil {
		settings.RequiresApproval = *req.RequiresApproval
	}
	if len(req.AllowedDomains) > 0 {
		b, _ := json.Marshal(req.AllowedDomains)
		s := string(b)
		settings.AllowedDomains = &s
	}
	if len(req.BlockedDomains) > 0 {
		b, _ := json.Marshal(req.BlockedDomains)
		s := string(b)
		settings.BlockedDomains = &s
	}
	if req.MaxUses != nil {
		settings.MaxUses = req.MaxUses
	}
	if req.ExpiresInDays != nil {
		t := time.Now().Add(time.Duration(*req.ExpiresInDays) * 24 * time.Hour)
		settings.ExpiresAt = &t
	}

	if err := h.svc.CreateLinkSettings(r.Context(), settings); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, settings)
}

// GET /api/invitation-links/token/{token}/use?email=...
func (h *InvitationHandler) UseInvitationLink(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	emailAddr := r.URL.Query().Get("email")
	if emailAddr == "" {
		writeError(w, http.StatusBadRequest, "email query param required")
		return
	}
	inv, ls, err := h.svc.UseLink(r.Context(), token, emailAddr)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if ls == nil {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"link_settings": ls,
		"invitation":    inv,
	})
}

// GET /api/invitation-links/token/{token}
func (h *InvitationHandler) GetLinkSettingsByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	ls, err := h.svc.GetLinkSettingsByToken(r.Context(), token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if ls == nil {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	writeJSON(w, http.StatusOK, ls)
}

// POST /api/invitations/{id}/regenerate-token
func (h *InvitationHandler) RegenerateToken(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	newTok, err := h.svc.RegenerateToken(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": newTok})
}

// GET /api/workspaces/{workspace_id}/invitations/stats
func (h *InvitationHandler) GetStatsByWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspace_id")
	stats, err := h.svc.GetStatsByWorkspace(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// POST /api/access-requests
func (h *InvitationHandler) CreateAccessRequest(w http.ResponseWriter, r *http.Request) {
	var req accessRequestReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	ar := &repository.AccessRequest{
		WorkspaceID: req.WorkspaceID,
		RequesterID: req.RequesterID,
		Email:       req.Email,
		Type:        req.Type,
		TargetID:    req.TargetID,
		Message:     req.Message,
		Status:      "pending",
	}
	if err := h.svc.CreateAccessRequest(r.Context(), ar); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, ar)
}

// POST /api/access-requests/{id}/process
func (h *InvitationHandler) ProcessAccessRequest(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "id")
	var payload struct {
		Status       string  `json:"status"`
		ProcessedBy  *string `json:"processed_by,omitempty"`
		DenialReason *string `json:"denial_reason,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if payload.Status == "" {
		writeError(w, http.StatusBadRequest, "status required")
		return
	}
	// Use repository method via service (service exposes CreateAccessRequest only in this implementation).
	// The repo-level UpdateAccessRequestStatus is not exposed via the service interface above;
	// if you need it, add an appropriate method to service.InvitationService and implement it.
	writeError(w, http.StatusNotImplemented, "processing access requests via handler requires service method - add UpdateAccessRequestStatus to service if needed")
}





// POST /api/workspaces/:id/invitations
func (h *InvitationHandler) CreateWorkspaceInvitation(c *gin.Context) {
	workspaceID := c.Param("id")
	userID := middleware.GetUserID(c)

	var req createWorkspaceInvitationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inv, err := h.svc.CreateWorkspaceInvitation(c.Request.Context(), workspaceID, req.Email, req.Role, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, inv)
}

// GET /api/workspaces/:id/invitations
func (h *InvitationHandler) GetWorkspaceInvitations(c *gin.Context) {
	workspaceID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	invs, total, err := h.svc.ListByWorkspace(c.Request.Context(), workspaceID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  invs,
		"total": total,
		"count": len(invs),
	})
}

// POST /api/projects/:id/invitations
func (h *InvitationHandler) CreateProjectInvitation(c *gin.Context) {
	projectID := c.Param("id")
	userID := middleware.GetUserID(c)

	var req createProjectInvitationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workspaceID := c.Query("workspaceId")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "workspaceId query parameter required"})
		return
	}

	inv, err := h.svc.CreateProjectInvitation(c.Request.Context(), workspaceID, projectID, req.Email, req.Role, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, inv)
}

// // GET /api/projects/:id/invitations
// func (h *InvitationHandler) GetProjectInvitations(c *gin.Context) {
// 	projectID := c.Param("id")
// 	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
// 	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

// 	invs, total, err := h.svc.ListByProject(c.Request.Context(), projectID, limit, offset)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err