// internal/service/invitation_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Marga-Ghale/ora-scrum-backend/internal/email"
	"github.com/Marga-Ghale/ora-scrum-backend/internal/repository"
)

// ============================================
// Invitation Service - ClickUp Style
// ============================================

// InvitationService defines invitation operations
type InvitationService interface {
	// Primary invitation method - invite to workspace
	CreateWorkspaceInvitation(ctx context.Context, workspaceID, email, role, invitedByID string) (*InvitationResponse, error)

	// Project invitation - user must be workspace member OR will be added to workspace too
	CreateProjectInvitation(ctx context.Context, projectID, email, role, invitedByID string) (*InvitationResponse, error)

	// Accept any invitation type
	AcceptInvitation(ctx context.Context, token, userID string) (*AcceptInvitationResponse, error)

	// Get pending invitations for current user
	GetPendingInvitations(ctx context.Context, email string) ([]*InvitationWithDetails, error)

	// Get pending invitations for a workspace (admin view)
	GetPendingWorkspaceInvitations(ctx context.Context, workspaceID string) ([]*InvitationWithDetails, error)

	// Get pending invitations for a project (admin view)
	GetPendingProjectInvitations(ctx context.Context, projectID string) ([]*InvitationWithDetails, error)

	// Cancel/delete an invitation
	CancelInvitation(ctx context.Context, id, userID string) error

	// Resend invitation email
	ResendInvitation(ctx context.Context, id, userID string) error
}

// InvitationResponse is the API response for created invitations
type InvitationResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Type        string    `json:"type"`
	TargetID    string    `json:"targetId"`
	TargetName  string    `json:"targetName"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	InvitedBy   string    `json:"invitedBy"`
	InviterName string    `json:"inviterName"`
	ExpiresAt   time.Time `json:"expiresAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

// InvitationWithDetails includes target info for display
type InvitationWithDetails struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Type         string    `json:"type"`
	TargetID     string    `json:"targetId"`
	TargetName   string    `json:"targetName"`
	TargetIcon   string    `json:"targetIcon,omitempty"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	InvitedBy    string    `json:"invitedBy"`
	InviterName  string    `json:"inviterName"`
	InviterEmail string    `json:"inviterEmail"`
	ExpiresAt    time.Time `json:"expiresAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

// AcceptInvitationResponse shows what happened when accepting
type AcceptInvitationResponse struct {
	Message     string `json:"message"`
	Type        string `json:"type"`
	TargetID    string `json:"targetId"`
	TargetName  string `json:"targetName"`
	WorkspaceID string `json:"workspaceId,omitempty"`
}

type invitationService struct {
	invitationRepo repository.InvitationRepository
	workspaceRepo  repository.WorkspaceRepository
	teamRepo       repository.TeamRepository
	projectRepo    repository.ProjectRepository
	userRepo       repository.UserRepository
	spaceRepo      repository.SpaceRepository
	emailSvc       *email.Service
}

// NewInvitationService creates a new invitation service
func NewInvitationService(
	invitationRepo repository.InvitationRepository,
	workspaceRepo repository.WorkspaceRepository,
	teamRepo repository.TeamRepository,
	projectRepo repository.ProjectRepository,
	userRepo repository.UserRepository,
	spaceRepo repository.SpaceRepository, // NEW parameter
	emailSvc *email.Service,
) InvitationService {
	return &invitationService{
		invitationRepo: invitationRepo,
		workspaceRepo:  workspaceRepo,
		teamRepo:       teamRepo,
		projectRepo:    projectRepo,
		userRepo:       userRepo,
		spaceRepo:      spaceRepo,
		emailSvc:       emailSvc,
	}
}

// CreateWorkspaceInvitation - Primary invitation method
func (s *invitationService) CreateWorkspaceInvitation(ctx context.Context, workspaceID, inviteEmail, role, invitedByID string) (*InvitationResponse, error) {
	// 1. Verify workspace exists
	workspace, err := s.workspaceRepo.FindByID(ctx, workspaceID)
	if err != nil || workspace == nil {
		return nil, ErrNotFound
	}

	// 2. Verify inviter has permission (must be owner or admin)
	inviterMember, _ := s.workspaceRepo.FindMember(ctx, workspaceID, invitedByID)
	if inviterMember == nil {
		return nil, ErrForbidden
	}
	if inviterMember.Role != "owner" && inviterMember.Role != "admin" {
		return nil, ErrForbidden
	}

	// 3. Normalize email
	inviteEmail = strings.ToLower(strings.TrimSpace(inviteEmail))

	// 4. Check if user already exists and is a member
	existingUser, _ := s.userRepo.FindByEmail(ctx, inviteEmail)
	if existingUser != nil {
		member, _ := s.workspaceRepo.FindMember(ctx, workspaceID, existingUser.ID)
		if member != nil {
			return nil, errors.New("user is already a workspace member")
		}
	}

	// 5. Check if invitation already pending
	existingInvites, _ := s.invitationRepo.FindByEmail(ctx, inviteEmail)
	for _, inv := range existingInvites {
		if inv.TargetID == workspaceID && inv.Type == "workspace" && inv.Status == "pending" {
			return nil, errors.New("invitation already sent to this email")
		}
	}

	// 6. Create invitation
	invitation := &repository.Invitation{
		Email:     inviteEmail,
		Type:      "workspace",
		TargetID:  workspaceID,
		Role:      role,
		InvitedBy: invitedByID,
		Status:    "pending",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	// 7. Get inviter info for response
	inviter, _ := s.userRepo.FindByID(ctx, invitedByID)
	inviterName := "Someone"
	if inviter != nil {
		inviterName = inviter.Name
	}

	// 8. Send invitation email
	if s.emailSvc != nil {
		s.emailSvc.SendWorkspaceInvitation(inviteEmail, email.WorkspaceInvitationData{
			InviterName:   inviterName,
			WorkspaceName: workspace.Name,
			Role:          role,
			InviteURL:     fmt.Sprintf("/invite/%s", invitation.Token),
		})
	}

	return &InvitationResponse{
		ID:          invitation.ID,
		Email:       invitation.Email,
		Type:        invitation.Type,
		TargetID:    invitation.TargetID,
		TargetName:  workspace.Name,
		Role:        invitation.Role,
		Status:      invitation.Status,
		InvitedBy:   invitation.InvitedBy,
		InviterName: inviterName,
		ExpiresAt:   invitation.ExpiresAt,
		CreatedAt:   invitation.CreatedAt,
	}, nil
}

// CreateProjectInvitation - ClickUp style: also adds to workspace if needed
func (s *invitationService) CreateProjectInvitation(ctx context.Context, projectID, inviteEmail, role, invitedByID string) (*InvitationResponse, error) {
	// 1. Verify project exists
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil || project == nil {
		return nil, ErrNotFound
	}

	// 2. Get workspace through space
	space, err := s.spaceRepo.FindByID(ctx, project.SpaceID)
	if err != nil || space == nil {
		return nil, ErrNotFound
	}
	workspaceID := space.WorkspaceID

	// 3. Verify inviter has permission on project
	inviterMember, _ := s.projectRepo.FindMember(ctx, projectID, invitedByID)
	if inviterMember == nil {
		// Check workspace level permission
		wsInviter, _ := s.workspaceRepo.FindMember(ctx, workspaceID, invitedByID)
		if wsInviter == nil || (wsInviter.Role != "owner" && wsInviter.Role != "admin") {
			return nil, ErrForbidden
		}
	} else if inviterMember.Role != "owner" && inviterMember.Role != "admin" && inviterMember.Role != "lead" {
		return nil, ErrForbidden
	}

	// 4. Normalize email
	inviteEmail = strings.ToLower(strings.TrimSpace(inviteEmail))

	// 5. Check if user already exists
	existingUser, _ := s.userRepo.FindByEmail(ctx, inviteEmail)
	if existingUser != nil {
		// Check if already a project member
		projectMember, _ := s.projectRepo.FindMember(ctx, projectID, existingUser.ID)
		if projectMember != nil {
			return nil, errors.New("user is already a project member")
		}
	}

	// 6. Check if invitation already pending
	existingInvites, _ := s.invitationRepo.FindByEmail(ctx, inviteEmail)
	for _, inv := range existingInvites {
		if inv.TargetID == projectID && inv.Type == "project" && inv.Status == "pending" {
			return nil, errors.New("invitation already sent to this email")
		}
	}

	// 7. Create invitation (store workspace ID in metadata or handle in accept)
	invitation := &repository.Invitation{
		Email:     inviteEmail,
		Type:      "project",
		TargetID:  projectID,
		Role:      role,
		InvitedBy: invitedByID,
		Status:    "pending",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	// 8. Get inviter info
	inviter, _ := s.userRepo.FindByID(ctx, invitedByID)
	inviterName := "Someone"
	if inviter != nil {
		inviterName = inviter.Name
	}

	// 9. Send invitation email
	if s.emailSvc != nil {
		workspace, _ := s.workspaceRepo.FindByID(ctx, workspaceID)
		workspaceName := ""
		if workspace != nil {
			workspaceName = workspace.Name
		}

		s.emailSvc.SendProjectInvitation(inviteEmail, email.ProjectInvitationData{
			InviterName:   inviterName,
			ProjectName:   project.Name,
			WorkspaceName: workspaceName,
			Role:          role,
			InviteURL:     fmt.Sprintf("/invite/%s", invitation.Token),
		})
	}

	return &InvitationResponse{
		ID:          invitation.ID,
		Email:       invitation.Email,
		Type:        invitation.Type,
		TargetID:    invitation.TargetID,
		TargetName:  project.Name,
		Role:        invitation.Role,
		Status:      invitation.Status,
		InvitedBy:   invitation.InvitedBy,
		InviterName: inviterName,
		ExpiresAt:   invitation.ExpiresAt,
		CreatedAt:   invitation.CreatedAt,
	}, nil
}

// AcceptInvitation - ClickUp style: handles workspace membership automatically
func (s *invitationService) AcceptInvitation(ctx context.Context, token, userID string) (*AcceptInvitationResponse, error) {
	// 1. Find invitation by token
	invitation, err := s.invitationRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, ErrNotFound
	}

	// 2. Check status
	if invitation.Status != "pending" {
		return nil, errors.New("invitation already " + invitation.Status)
	}

	// 3. Check expiration
	if time.Now().After(invitation.ExpiresAt) {
		invitation.Status = "expired"
		s.invitationRepo.Update(ctx, invitation)
		return nil, errors.New("invitation has expired")
	}

	// 4. Verify user email matches invitation
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}
	if strings.ToLower(user.Email) != strings.ToLower(invitation.Email) {
		return nil, errors.New("this invitation was sent to a different email address")
	}

	var response *AcceptInvitationResponse

	// 5. Handle based on invitation type
	switch invitation.Type {
	case "workspace":
		workspace, _ := s.workspaceRepo.FindByID(ctx, invitation.TargetID)
		if workspace == nil {
			return nil, ErrNotFound
		}

		// Check if already a member
		existing, _ := s.workspaceRepo.FindMember(ctx, invitation.TargetID, userID)
		if existing != nil {
			invitation.Status = "accepted"
			s.invitationRepo.Update(ctx, invitation)
			return &AcceptInvitationResponse{
				Message:    "You are already a member of this workspace",
				Type:       "workspace",
				TargetID:   invitation.TargetID,
				TargetName: workspace.Name,
			}, nil
		}

		member := &repository.WorkspaceMember{
			WorkspaceID: invitation.TargetID,
			UserID:      userID,
			Role:        invitation.Role,
		}
		if err := s.workspaceRepo.AddMember(ctx, member); err != nil {
			return nil, err
		}

		response = &AcceptInvitationResponse{
			Message:     fmt.Sprintf("You have joined %s", workspace.Name),
			Type:        "workspace",
			TargetID:    invitation.TargetID,
			TargetName:  workspace.Name,
			WorkspaceID: invitation.TargetID,
		}

	case "project":
		project, _ := s.projectRepo.FindByID(ctx, invitation.TargetID)
		if project == nil {
			return nil, ErrNotFound
		}

		// Get workspace through space
		space, _ := s.spaceRepo.FindByID(ctx, project.SpaceID)
		if space == nil {
			return nil, ErrNotFound
		}
		workspaceID := space.WorkspaceID

		// CLICKUP STYLE: First ensure user is workspace member
		wsMember, _ := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
		if wsMember == nil {
			// Auto-add to workspace as "member" role
			wsMember := &repository.WorkspaceMember{
				WorkspaceID: workspaceID,
				UserID:      userID,
				Role:        "member", // Default workspace role
			}
			if err := s.workspaceRepo.AddMember(ctx, wsMember); err != nil {
				return nil, fmt.Errorf("failed to add to workspace: %w", err)
			}
		}

		// Now add to project
		existing, _ := s.projectRepo.FindMember(ctx, invitation.TargetID, userID)
		if existing != nil {
			invitation.Status = "accepted"
			s.invitationRepo.Update(ctx, invitation)
			return &AcceptInvitationResponse{
				Message:     "You are already a member of this project",
				Type:        "project",
				TargetID:    invitation.TargetID,
				TargetName:  project.Name,
				WorkspaceID: workspaceID,
			}, nil
		}

		projectMember := &repository.ProjectMember{
			ProjectID: invitation.TargetID,
			UserID:    userID,
			Role:      invitation.Role,
		}
		if err := s.projectRepo.AddMember(ctx, projectMember); err != nil {
			return nil, err
		}

		response = &AcceptInvitationResponse{
			Message:     fmt.Sprintf("You have joined %s", project.Name),
			Type:        "project",
			TargetID:    invitation.TargetID,
			TargetName:  project.Name,
			WorkspaceID: workspaceID,
		}

	case "team":
		team, _ := s.teamRepo.FindByID(ctx, invitation.TargetID)
		if team == nil {
			return nil, ErrNotFound
		}

		// Ensure workspace membership first
		wsMember, _ := s.workspaceRepo.FindMember(ctx, team.WorkspaceID, userID)
		if wsMember == nil {
			wsMember := &repository.WorkspaceMember{
				WorkspaceID: team.WorkspaceID,
				UserID:      userID,
				Role:        "member",
			}
			if err := s.workspaceRepo.AddMember(ctx, wsMember); err != nil {
				return nil, fmt.Errorf("failed to add to workspace: %w", err)
			}
		}

		teamMember := &repository.TeamMember{
			TeamID: invitation.TargetID,
			UserID: userID,
			Role:   invitation.Role,
		}
		if err := s.teamRepo.AddMember(ctx, teamMember); err != nil {
			return nil, err
		}

		response = &AcceptInvitationResponse{
			Message:     fmt.Sprintf("You have joined team %s", team.Name),
			Type:        "team",
			TargetID:    invitation.TargetID,
			TargetName:  team.Name,
			WorkspaceID: team.WorkspaceID,
		}

	default:
		return nil, errors.New("unknown invitation type")
	}

	// 6. Update invitation status
	invitation.Status = "accepted"
	s.invitationRepo.Update(ctx, invitation)

	return response, nil
}

// GetPendingInvitations gets all pending invitations for a user's email
func (s *invitationService) GetPendingInvitations(ctx context.Context, userEmail string) ([]*InvitationWithDetails, error) {
	invitations, err := s.invitationRepo.FindByEmail(ctx, strings.ToLower(userEmail))
	if err != nil {
		return nil, err
	}

	result := make([]*InvitationWithDetails, 0, len(invitations))
	for _, inv := range invitations {
		detail := s.enrichInvitation(ctx, inv)
		result = append(result, detail)
	}

	return result, nil
}

// GetPendingWorkspaceInvitations gets pending invitations for a workspace
func (s *invitationService) GetPendingWorkspaceInvitations(ctx context.Context, workspaceID string) ([]*InvitationWithDetails, error) {
	invitations, err := s.invitationRepo.FindPendingByTarget(ctx, "workspace", workspaceID)
	if err != nil {
		return nil, err
	}

	result := make([]*InvitationWithDetails, 0, len(invitations))
	for _, inv := range invitations {
		detail := s.enrichInvitation(ctx, inv)
		result = append(result, detail)
	}

	return result, nil
}

// GetPendingProjectInvitations gets pending invitations for a project
func (s *invitationService) GetPendingProjectInvitations(ctx context.Context, projectID string) ([]*InvitationWithDetails, error) {
	invitations, err := s.invitationRepo.FindPendingByTarget(ctx, "project", projectID)
	if err != nil {
		return nil, err
	}

	result := make([]*InvitationWithDetails, 0, len(invitations))
	for _, inv := range invitations {
		detail := s.enrichInvitation(ctx, inv)
		result = append(result, detail)
	}

	return result, nil
}

// CancelInvitation cancels an invitation (only inviter or workspace admin can cancel)
func (s *invitationService) CancelInvitation(ctx context.Context, id, userID string) error {
	invitation, err := s.invitationRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if invitation == nil {
		return ErrNotFound
	}

	// Check permission: inviter can always cancel
	if invitation.InvitedBy != userID {
		// Check if user is workspace admin
		var workspaceID string
		switch invitation.Type {
		case "workspace":
			workspaceID = invitation.TargetID
		case "project":
			project, _ := s.projectRepo.FindByID(ctx, invitation.TargetID)
			if project != nil {
				space, _ := s.spaceRepo.FindByID(ctx, project.SpaceID)
				if space != nil {
					workspaceID = space.WorkspaceID
				}
			}
		case "team":
			team, _ := s.teamRepo.FindByID(ctx, invitation.TargetID)
			if team != nil {
				workspaceID = team.WorkspaceID
			}
		}

		if workspaceID != "" {
			member, _ := s.workspaceRepo.FindMember(ctx, workspaceID, userID)
			if member == nil || (member.Role != "owner" && member.Role != "admin") {
				return ErrForbidden
			}
		} else {
			return ErrForbidden
		}
	}

	return s.invitationRepo.Delete(ctx, id)
}

// ResendInvitation resends the invitation email
func (s *invitationService) ResendInvitation(ctx context.Context, id, userID string) error {
	invitation, err := s.invitationRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if invitation == nil {
		return ErrNotFound
	}

	if invitation.Status != "pending" {
		return errors.New("can only resend pending invitations")
	}

	// Check permission
	if invitation.InvitedBy != userID {
		return ErrForbidden
	}

	// Extend expiration
	invitation.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	if err := s.invitationRepo.Update(ctx, invitation); err != nil {
		return err
	}

	// Resend email
	if s.emailSvc != nil {
		inviter, _ := s.userRepo.FindByID(ctx, userID)
		inviterName := "Someone"
		if inviter != nil {
			inviterName = inviter.Name
		}

		switch invitation.Type {
		case "workspace":
			workspace, _ := s.workspaceRepo.FindByID(ctx, invitation.TargetID)
			if workspace != nil {
				s.emailSvc.SendWorkspaceInvitation(invitation.Email, email.WorkspaceInvitationData{
					InviterName:   inviterName,
					WorkspaceName: workspace.Name,
					Role:          invitation.Role,
					InviteURL:     fmt.Sprintf("/invite/%s", invitation.Token),
				})
			}
		case "project":
			project, _ := s.projectRepo.FindByID(ctx, invitation.TargetID)
			if project != nil {
				s.emailSvc.SendProjectInvitation(invitation.Email, email.ProjectInvitationData{
					InviterName: inviterName,
					ProjectName: project.Name,
					Role:        invitation.Role,
					InviteURL:   fmt.Sprintf("/invite/%s", invitation.Token),
				})
			}
		}
	}

	return nil
}

// Add this helper function at the top of the file (after imports)
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Replace your enrichInvitation function with this:
func (s *invitationService) enrichInvitation(ctx context.Context, inv *repository.Invitation) *InvitationWithDetails {
	detail := &InvitationWithDetails{
		ID:        inv.ID,
		Email:     inv.Email,
		Type:      inv.Type,
		TargetID:  inv.TargetID,
		Role:      inv.Role,
		Status:    inv.Status,
		InvitedBy: inv.InvitedBy,
		ExpiresAt: inv.ExpiresAt,
		CreatedAt: inv.CreatedAt,
	}

	// Get target name and icon
	switch inv.Type {
	case "workspace":
		workspace, _ := s.workspaceRepo.FindByID(ctx, inv.TargetID)
		if workspace != nil {
			detail.TargetName = workspace.Name
			detail.TargetIcon = derefString(workspace.Icon)
		}
	case "project":
		project, _ := s.projectRepo.FindByID(ctx, inv.TargetID)
		if project != nil {
			detail.TargetName = project.Name
			detail.TargetIcon = derefString(project.Icon)
		}
	case "team":
		team, _ := s.teamRepo.FindByID(ctx, inv.TargetID)
		if team != nil {
			detail.TargetName = team.Name
		}
	}

	// Get inviter info
	inviter, _ := s.userRepo.FindByID(ctx, inv.InvitedBy)
	if inviter != nil {
		detail.InviterName = inviter.Name
		detail.InviterEmail = inviter.Email
	}

	return detail
}
