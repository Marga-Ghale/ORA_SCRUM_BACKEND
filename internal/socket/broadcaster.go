package socket

import (
	"fmt"
	"log"
)

// Broadcaster provides high-level methods for broadcasting events
type Broadcaster struct {
	hub *Hub
}

// NewBroadcaster creates a new Broadcaster
func NewBroadcaster(hub *Hub) *Broadcaster {
	return &Broadcaster{hub: hub}
}

// ============================================
// Notification Broadcasting
// ============================================

// SendNotification sends a notification to a specific user
func (b *Broadcaster) SendNotification(userID string, notification map[string]interface{}) {
	b.hub.SendToUser(userID, MessageNotification, notification)
}

// SendNotificationCount updates notification count for a user
func (b *Broadcaster) SendNotificationCount(userID string, total, unread int) {
	b.hub.SendToUser(userID, MessageNotificationCount, map[string]interface{}{
		"total":  total,
		"unread": unread,
	})
}

// ============================================
// Task Broadcasting
// ============================================

// BroadcastTaskCreated broadcasts task creation to project members
func (b *Broadcaster) BroadcastTaskCreated(projectID string, task map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("project:%s", projectID)
	b.hub.SendToRoom(room, MessageTaskCreated, task, excludeUserID)
}

// BroadcastTaskUpdated broadcasts task updates to project members
func (b *Broadcaster) BroadcastTaskUpdated(
	projectID string,
	task map[string]interface{},
	changes []string,
	excludeUserID string,
) {
	room := fmt.Sprintf("project:%s", projectID)

	payload := map[string]interface{}{
		"task":          task,
		"changedFields": changes,
		"changedByUser": excludeUserID,
		"projectId":     projectID,
	}

	log.Printf("📡 BroadcastTaskUpdated: room=%s, taskId=%v, exclude=%s",
		room, task["id"], excludeUserID)

	b.hub.SendToRoom(room, MessageTaskUpdated, payload, excludeUserID)
}

// BroadcastTaskDeleted broadcasts task deletion to project members
func (b *Broadcaster) BroadcastTaskDeleted(projectID, taskID, taskKey string, excludeUserID string) {
	room := fmt.Sprintf("project:%s", projectID)
	b.hub.SendToRoom(room, MessageTaskDeleted, map[string]interface{}{
		"taskId":  taskID,
		"taskKey": taskKey,
	}, excludeUserID)
}

// BroadcastTaskStatusChanged broadcasts task status change to project members
func (b *Broadcaster) BroadcastTaskStatusChanged(projectID string, task map[string]interface{}, oldStatus, newStatus string, excludeUserID string) {
	room := fmt.Sprintf("project:%s", projectID)
	b.hub.SendToRoom(room, MessageTaskStatusChanged, map[string]interface{}{
		"task":          task,
		"oldStatus":     oldStatus,
		"newStatus":     newStatus,
		"changedByUser": excludeUserID,
	}, excludeUserID)
}

// BroadcastTaskPositionChanged broadcasts task position/status change WITHOUT notifications
func (b *Broadcaster) BroadcastTaskPositionChanged(projectID string, task map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("project:%s", projectID)

	payload := map[string]interface{}{
		"task":      task,
		"projectId": projectID,
	}

	log.Printf("📡 BroadcastTaskPositionChanged: room=%s, taskId=%v, exclude=%s",
		room, task["id"], excludeUserID)

	b.hub.SendToRoom(room, MessageTaskPositionChanged, payload, excludeUserID)
}

// BroadcastTaskAssigned notifies the assigned user
func (b *Broadcaster) BroadcastTaskAssigned(assigneeID string, task map[string]interface{}, assignedBy string) {
	b.hub.SendToUser(assigneeID, MessageTaskAssigned, map[string]interface{}{
		"task":       task,
		"assignedBy": assignedBy,
	})
}

// ============================================
// Sprint Broadcasting
// ============================================

// BroadcastSprintStarted broadcasts sprint start to project members
func (b *Broadcaster) BroadcastSprintStarted(projectID string, sprint map[string]interface{}) {
	room := fmt.Sprintf("project:%s", projectID)
	b.hub.SendToRoom(room, MessageSprintStarted, sprint, "")
}

// BroadcastSprintCompleted broadcasts sprint completion to project members
func (b *Broadcaster) BroadcastSprintCompleted(projectID string, sprint map[string]interface{}, stats map[string]interface{}) {
	room := fmt.Sprintf("project:%s", projectID)
	b.hub.SendToRoom(room, MessageSprintCompleted, map[string]interface{}{
		"sprint": sprint,
		"stats":  stats,
	}, "")
}

// ============================================
// Comment Broadcasting
// ============================================

// BroadcastCommentAdded broadcasts new comment to task watchers
func (b *Broadcaster) BroadcastCommentAdded(projectID, taskID string, comment map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("project:%s", projectID)
	b.hub.SendToRoom(room, MessageCommentAdded, map[string]interface{}{
		"taskId":  taskID,
		"comment": comment,
	}, excludeUserID)
}

// BroadcastCommentUpdated broadcasts comment update
func (b *Broadcaster) BroadcastCommentUpdated(projectID, taskID string, comment map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("project:%s", projectID)
	b.hub.SendToRoom(room, MessageCommentUpdated, map[string]interface{}{
		"taskId":  taskID,
		"comment": comment,
	}, excludeUserID)
}

// BroadcastCommentDeleted broadcasts comment deletion
func (b *Broadcaster) BroadcastCommentDeleted(projectID, taskID, commentID string, excludeUserID string) {
	room := fmt.Sprintf("project:%s", projectID)
	b.hub.SendToRoom(room, MessageCommentDeleted, map[string]interface{}{
		"taskId":    taskID,
		"commentId": commentID,
	}, excludeUserID)
}

// ============================================
// Team Broadcasting
// ============================================

// BroadcastTeamCreated broadcasts team creation to workspace members
func (b *Broadcaster) BroadcastTeamCreated(workspaceID string, team map[string]interface{}) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	b.hub.SendToRoom(room, MessageTeamCreated, team, "")
}

// BroadcastTeamUpdated broadcasts team updates
func (b *Broadcaster) BroadcastTeamUpdated(workspaceID string, team map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	b.hub.SendToRoom(room, MessageTeamUpdated, team, excludeUserID)
}

// BroadcastTeamDeleted broadcasts team deletion
func (b *Broadcaster) BroadcastTeamDeleted(workspaceID, teamID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	b.hub.SendToRoom(room, MessageTeamDeleted, map[string]interface{}{
		"teamId": teamID,
	}, "")
}

// BroadcastTeamMemberAdded broadcasts team member addition
func (b *Broadcaster) BroadcastTeamMemberAdded(workspaceID, teamID string, member map[string]interface{}) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	b.hub.SendToRoom(room, MessageTeamMemberAdded, map[string]interface{}{
		"teamId": teamID,
		"member": member,
	}, "")
}

// BroadcastTeamMemberRemoved broadcasts team member removal
func (b *Broadcaster) BroadcastTeamMemberRemoved(workspaceID, teamID, userID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	b.hub.SendToRoom(room, MessageTeamMemberRemoved, map[string]interface{}{
		"teamId": teamID,
		"userId": userID,
	}, "")
}

// ============================================
// ✅ NEW: Workspace CRUD Broadcasting
// ============================================

// BroadcastWorkspaceCreated broadcasts workspace creation to the creator
// Note: New workspaces don't have other members yet, so we send to the creator
func (b *Broadcaster) BroadcastWorkspaceCreated(creatorID string, workspace map[string]interface{}) {
	log.Printf("📡 BroadcastWorkspaceCreated: creator=%s, workspace=%v", creatorID, workspace["id"])
	b.hub.SendToUser(creatorID, MessageWorkspaceCreated, workspace)
}

// BroadcastWorkspaceUpdated broadcasts workspace update to all workspace members
func (b *Broadcaster) BroadcastWorkspaceUpdated(workspaceID string, workspace map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	log.Printf("📡 BroadcastWorkspaceUpdated: room=%s, exclude=%s", room, excludeUserID)
	b.hub.SendToRoom(room, MessageWorkspaceUpdated, workspace, excludeUserID)
}

// BroadcastWorkspaceDeleted broadcasts workspace deletion to all workspace members
func (b *Broadcaster) BroadcastWorkspaceDeleted(workspaceID string, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	log.Printf("📡 BroadcastWorkspaceDeleted: room=%s, exclude=%s", room, excludeUserID)
	b.hub.SendToRoom(room, MessageWorkspaceDeleted, map[string]interface{}{
		"id":          workspaceID,
		"workspaceId": workspaceID,
	}, excludeUserID)
}

// ============================================
// ✅ NEW: Space CRUD Broadcasting
// ============================================

// BroadcastSpaceCreated broadcasts space creation to workspace members
func (b *Broadcaster) BroadcastSpaceCreated(workspaceID string, space map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	log.Printf("📡 BroadcastSpaceCreated: room=%s, spaceId=%v, exclude=%s", room, space["id"], excludeUserID)
	b.hub.SendToRoom(room, MessageSpaceCreated, space, excludeUserID)
}

// BroadcastSpaceUpdated broadcasts space update to workspace members
func (b *Broadcaster) BroadcastSpaceUpdated(workspaceID string, space map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	log.Printf("📡 BroadcastSpaceUpdated: room=%s, spaceId=%v, exclude=%s", room, space["id"], excludeUserID)
	b.hub.SendToRoom(room, MessageSpaceUpdated, space, excludeUserID)
}

// BroadcastSpaceDeleted broadcasts space deletion to workspace members
func (b *Broadcaster) BroadcastSpaceDeleted(workspaceID, spaceID string, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	log.Printf("📡 BroadcastSpaceDeleted: room=%s, spaceId=%s, exclude=%s", room, spaceID, excludeUserID)
	b.hub.SendToRoom(room, MessageSpaceDeleted, map[string]interface{}{
		"id":          spaceID,
		"spaceId":     spaceID,
		"workspaceId": workspaceID,
	}, excludeUserID)
}

// ============================================
// ✅ NEW: Folder CRUD Broadcasting
// ============================================

// BroadcastFolderCreated broadcasts folder creation to workspace members
func (b *Broadcaster) BroadcastFolderCreated(workspaceID, spaceID string, folder map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	log.Printf("📡 BroadcastFolderCreated: room=%s, folderId=%v, spaceId=%s, exclude=%s", room, folder["id"], spaceID, excludeUserID)

	// Add spaceId to payload for frontend filtering
	folder["spaceId"] = spaceID
	folder["workspaceId"] = workspaceID

	b.hub.SendToRoom(room, MessageFolderCreated, folder, excludeUserID)
}

// BroadcastFolderUpdated broadcasts folder update to workspace members
func (b *Broadcaster) BroadcastFolderUpdated(workspaceID, spaceID string, folder map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	log.Printf("📡 BroadcastFolderUpdated: room=%s, folderId=%v, exclude=%s", room, folder["id"], excludeUserID)

	folder["spaceId"] = spaceID
	folder["workspaceId"] = workspaceID

	b.hub.SendToRoom(room, MessageFolderUpdated, folder, excludeUserID)
}

// BroadcastFolderDeleted broadcasts folder deletion to workspace members
func (b *Broadcaster) BroadcastFolderDeleted(workspaceID, spaceID, folderID string, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	log.Printf("📡 BroadcastFolderDeleted: room=%s, folderId=%s, exclude=%s", room, folderID, excludeUserID)
	b.hub.SendToRoom(room, MessageFolderDeleted, map[string]interface{}{
		"id":          folderID,
		"folderId":    folderID,
		"spaceId":     spaceID,
		"workspaceId": workspaceID,
	}, excludeUserID)
}

// ============================================
// ✅ NEW: Project CRUD Broadcasting
// ============================================

// BroadcastProjectCreated broadcasts project creation to workspace members
func (b *Broadcaster) BroadcastProjectCreated(workspaceID, spaceID string, folderID *string, project map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	log.Printf("📡 BroadcastProjectCreated: room=%s, projectId=%v, spaceId=%s, exclude=%s", room, project["id"], spaceID, excludeUserID)

	// Add hierarchy IDs to payload for frontend filtering
	project["spaceId"] = spaceID
	project["workspaceId"] = workspaceID
	if folderID != nil {
		project["folderId"] = *folderID
	}

	b.hub.SendToRoom(room, MessageProjectCreated, project, excludeUserID)
}

// BroadcastProjectUpdated broadcasts project update to workspace members
func (b *Broadcaster) BroadcastProjectUpdated(workspaceID string, project map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	log.Printf("📡 BroadcastProjectUpdated: room=%s, projectId=%v, exclude=%s", room, project["id"], excludeUserID)

	project["workspaceId"] = workspaceID

	b.hub.SendToRoom(room, MessageProjectUpdated, project, excludeUserID)
}

// BroadcastProjectDeleted broadcasts project deletion to workspace members
func (b *Broadcaster) BroadcastProjectDeleted(workspaceID, spaceID, projectID string, folderID *string, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	log.Printf("📡 BroadcastProjectDeleted: room=%s, projectId=%s, exclude=%s", room, projectID, excludeUserID)

	payload := map[string]interface{}{
		"id":          projectID,
		"projectId":   projectID,
		"spaceId":     spaceID,
		"workspaceId": workspaceID,
	}
	if folderID != nil {
		payload["folderId"] = *folderID
	}

	b.hub.SendToRoom(room, MessageProjectDeleted, payload, excludeUserID)
}

// ============================================
// Generic Broadcasting Methods
// ============================================

// BroadcastToWorkspace broadcasts a message to all workspace members
func (b *Broadcaster) BroadcastToWorkspace(workspaceID string, msgType MessageType, payload map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("workspace:%s", workspaceID)
	b.hub.SendToRoom(room, msgType, payload, excludeUserID)
}



// ============================================
// Direct User Messaging
// ============================================

// SendToUsers sends a message to multiple specific users
func (b *Broadcaster) SendToUsers(userIDs []string, msgType MessageType, payload map[string]interface{}) {
	for _, userID := range userIDs {
		b.hub.SendToUser(userID, msgType, payload)
	}
}



// ============================================
// Member Broadcasting 
// ============================================

// BroadcastMemberAdded broadcasts member addition to the appropriate entity room
func (b *Broadcaster) BroadcastMemberAdded(entityType, entityID string, member map[string]interface{}, excludeUserID string) {
	var room string
	
	switch entityType {
	case "workspace":
		room = fmt.Sprintf("workspace:%s", entityID)
	case "space":
		room = fmt.Sprintf("workspace:%s", member["workspaceId"]) // Broadcast to workspace room
	case "folder":
		room = fmt.Sprintf("workspace:%s", member["workspaceId"]) // Broadcast to workspace room
	case "project":
		room = fmt.Sprintf("workspace:%s", member["workspaceId"]) // Broadcast to workspace room
	default:
		log.Printf("⚠️ Unknown entity type for member addition: %s", entityType)
		return
	}

	payload := map[string]interface{}{
		"entityType": entityType,
		"entityId":   entityID,
		"member":     member,
	}

	log.Printf("📡 BroadcastMemberAdded: entityType=%s, entityId=%s, room=%s, exclude=%s",
		entityType, entityID, room, excludeUserID)

	b.hub.SendToRoom(room, MessageMemberAdded, payload, excludeUserID)
}

// BroadcastMemberRemoved broadcasts member removal to the appropriate entity room
func (b *Broadcaster) BroadcastMemberRemoved(entityType, entityID, userID string, workspaceID string, excludeUserID string) {
	var room string
	
	switch entityType {
	case "workspace":
		room = fmt.Sprintf("workspace:%s", entityID)
	case "space", "folder", "project":
		room = fmt.Sprintf("workspace:%s", workspaceID) // Broadcast to workspace room
	default:
		log.Printf("⚠️ Unknown entity type for member removal: %s", entityType)
		return
	}

	payload := map[string]interface{}{
		"entityType": entityType,
		"entityId":   entityID,
		"userId":     userID,
	}

	log.Printf("📡 BroadcastMemberRemoved: entityType=%s, entityId=%s, userId=%s, room=%s, exclude=%s",
		entityType, entityID, userID, room, excludeUserID)

	b.hub.SendToRoom(room, MessageMemberRemoved, payload, excludeUserID)
}

// BroadcastMemberRoleUpdated broadcasts member role update to the appropriate entity room
func (b *Broadcaster) BroadcastMemberRoleUpdated(entityType, entityID, userID, newRole string, workspaceID string, excludeUserID string) {
	var room string
	
	switch entityType {
	case "workspace":
		room = fmt.Sprintf("workspace:%s", entityID)
	case "space", "folder", "project":
		room = fmt.Sprintf("workspace:%s", workspaceID) // Broadcast to workspace room
	default:
		log.Printf("⚠️ Unknown entity type for member role update: %s", entityType)
		return
	}

	payload := map[string]interface{}{
		"entityType": entityType,
		"entityId":   entityID,
		"userId":     userID,
		"newRole":    newRole,
	}

	log.Printf("📡 BroadcastMemberRoleUpdated: entityType=%s, entityId=%s, userId=%s, newRole=%s, room=%s, exclude=%s",
		entityType, entityID, userID, newRole, room, excludeUserID)

	b.hub.SendToRoom(room, MessageMemberRoleUpdated, payload, excludeUserID)
}