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
		"task":           task,
		"changedFields":  changes,
		"changedByUser":  excludeUserID, // ✅ Keep track of who made the change
		"projectId":      projectID,
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
		"task":      task,
		"oldStatus": oldStatus,
		"newStatus": newStatus,
	}, excludeUserID)
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
// Project Broadcasting
// ============================================

// BroadcastProjectUpdated broadcasts project updates to members
func (b *Broadcaster) BroadcastProjectUpdated(projectID string, project map[string]interface{}, excludeUserID string) {
	room := fmt.Sprintf("project:%s", projectID)
	b.hub.SendToRoom(room, MessageProjectUpdated, project, excludeUserID)
}

// BroadcastMemberAdded broadcasts new member addition
func (b *Broadcaster) BroadcastMemberAdded(projectID string, member map[string]interface{}) {
	room := fmt.Sprintf("project:%s", projectID)
	b.hub.SendToRoom(room, MessageMemberAdded, member, "")
}

// BroadcastMemberRemoved broadcasts member removal
func (b *Broadcaster) BroadcastMemberRemoved(projectID, userID string) {
	room := fmt.Sprintf("project:%s", projectID)
	b.hub.SendToRoom(room, MessageMemberRemoved, map[string]interface{}{
		"userId": userID,
	}, "")
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
// Workspace Broadcasting
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
