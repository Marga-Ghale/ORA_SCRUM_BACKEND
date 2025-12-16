// Package email provides email sending functionality
package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"strings"
	"time"
)

// Config holds email configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	FromName string
	UseTLS   bool
}

// Service handles email sending
type Service struct {
	config    *Config
	templates map[string]*template.Template
}

// NewService creates a new email service
func NewService(config *Config) *Service {
	s := &Service{
		config:    config,
		templates: make(map[string]*template.Template),
	}
	s.loadTemplates()
	return s
}

// Email represents an email message
type Email struct {
	To       []string
	CC       []string
	BCC      []string
	Subject  string
	Body     string
	HTMLBody string
}

// InvitationEmailData holds data for invitation emails
type InvitationEmailData struct {
	WorkspaceName string
	InvitedBy     string
	InviteURL     string
}



// loadTemplates loads all email templates
func (s *Service) loadTemplates() {


	// Generic Invitation Template (used by service layer)
s.templates["invitation"] = template.Must(template.New("invitation").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #10b981; color: white; padding: 24px; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 24px; border-radius: 0 0 8px 8px; }
        .btn { display: inline-block; background: #10b981; color: white; padding: 12px 20px; text-decoration: none; border-radius: 6px; margin-top: 16px; }
        .footer { margin-top: 24px; font-size: 12px; color: #6b7280; text-align: center; }
    </style>
</head>
<body>
<div class="container">
    <div class="header">
        <h2>You're Invited to ORA Scrum</h2>
    </div>
    <div class="content">
        <p>Hello,</p>
        <p><strong>{{.InvitedBy}}</strong> invited you to join <strong>{{.WorkspaceName}}</strong>.</p>

        <a href="{{.InviteURL}}" class="btn">Accept Invitation</a>

        <p style="margin-top: 16px; font-size: 14px; color: #6b7280;">
            This invitation may expire. If you were not expecting this email, you can ignore it.
        </p>
    </div>
    <div class="footer">
        ORA Scrum • Team Collaboration Platform
    </div>
</div>
</body>
</html>
`))

	// Task Assigned Template
	s.templates["task_assigned"] = template.Must(template.New("task_assigned").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 10px 10px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 10px 10px; }
        .task-card { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .btn { display: inline-block; background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin-top: 15px; }
        .priority-high { color: #ef4444; }
        .priority-medium { color: #f59e0b; }
        .priority-low { color: #10b981; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📋 New Task Assigned</h1>
        </div>
        <div class="content">
            <p>Hi {{.AssigneeName}},</p>
            <p>You have been assigned a new task by <strong>{{.AssignerName}}</strong>.</p>

            <div class="task-card">
                <h2>{{.TaskKey}} - {{.TaskTitle}}</h2>
                <p><strong>Project:</strong> {{.ProjectName}}</p>
                <p><strong>Priority:</strong> <span class="priority-{{.Priority}}">{{.Priority}}</span></p>
                {{if .DueDate}}<p><strong>Due Date:</strong> {{.DueDate}}</p>{{end}}
                {{if .Description}}<p><strong>Description:</strong><br/>{{.Description}}</p>{{end}}
            </div>

            <a href="{{.TaskURL}}" class="btn">View Task</a>
        </div>
        <div class="footer">
            <p>This email was sent from ORA Scrum</p>
        </div>
    </div>
</body>
</html>
`))

	// Task Updated Template
	s.templates["task_updated"] = template.Must(template.New("task_updated").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 10px 10px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 10px 10px; }
        .changes { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; }
        .change-item { padding: 10px 0; border-bottom: 1px solid #e5e7eb; }
        .change-item:last-child { border-bottom: none; }
        .btn { display: inline-block; background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin-top: 15px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔄 Task Updated</h1>
        </div>
        <div class="content">
            <p>Hi {{.UserName}},</p>
            <p>Task <strong>{{.TaskKey}} - {{.TaskTitle}}</strong> has been updated by {{.UpdatedBy}}.</p>

            <div class="changes">
                <h3>Changes:</h3>
                {{range .Changes}}
                <div class="change-item">
                    <strong>{{.Field}}:</strong> {{.OldValue}} → {{.NewValue}}
                </div>
                {{end}}
            </div>

            <a href="{{.TaskURL}}" class="btn">View Task</a>
        </div>
        <div class="footer">
            <p>This email was sent from ORA Scrum</p>
        </div>
    </div>
</body>
</html>
`))

	// Workspace Invitation Template
	s.templates["workspace_invitation"] = template.Must(template.New("workspace_invitation").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #10b981 0%, #059669 100%); color: white; padding: 30px; border-radius: 10px 10px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 10px 10px; }
        .btn { display: inline-block; background: #10b981; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin-top: 15px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Workspace Invitation</h1>
        </div>
        <div class="content">
            <p>Hi there!</p>
            <p><strong>{{.InviterName}}</strong> has invited you to join the <strong>{{.WorkspaceName}}</strong> workspace on ORA Scrum.</p>
            <p>You've been assigned the role of <strong>{{.Role}}</strong>.</p>

            <a href="{{.InviteURL}}" class="btn">Accept Invitation</a>

            <p style="margin-top: 20px; color: #6b7280; font-size: 14px;">
                This invitation will expire in 7 days.
            </p>
        </div>
        <div class="footer">
            <p>If you didn't expect this invitation, you can safely ignore this email.</p>
        </div>
    </div>
</body>
</html>
`))

	// Project Invitation Template
	s.templates["project_invitation"] = template.Must(template.New("project_invitation").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%); color: white; padding: 30px; border-radius: 10px 10px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 10px 10px; }
        .project-card { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .btn { display: inline-block; background: #6366f1; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin-top: 15px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📁 Project Invitation</h1>
        </div>
        <div class="content">
            <p>Hi there!</p>
            <p><strong>{{.InviterName}}</strong> has invited you to join the project <strong>{{.ProjectName}}</strong>{{if .WorkspaceName}} in the <strong>{{.WorkspaceName}}</strong> workspace{{end}}.</p>
            <p>You've been assigned the role of <strong>{{.Role}}</strong>.</p>

            <div class="project-card">
                <h3>{{.ProjectName}}</h3>
                {{if .WorkspaceName}}<p><strong>Workspace:</strong> {{.WorkspaceName}}</p>{{end}}
                <p><strong>Your Role:</strong> {{.Role}}</p>
            </div>

            <a href="{{.InviteURL}}" class="btn">Accept Invitation</a>

            <p style="margin-top: 20px; color: #6b7280; font-size: 14px;">
                This invitation will expire in 7 days.
            </p>
        </div>
        <div class="footer">
            <p>If you didn't expect this invitation, you can safely ignore this email.</p>
        </div>
    </div>
</body>
</html>
`))

	// Team Invitation Template
	s.templates["team_invitation"] = template.Must(template.New("team_invitation").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #8b5cf6 0%, #6366f1 100%); color: white; padding: 30px; border-radius: 10px 10px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 10px 10px; }
        .btn { display: inline-block; background: #8b5cf6; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin-top: 15px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>👥 Team Invitation</h1>
        </div>
        <div class="content">
            <p>Hi {{.UserName}},</p>
            <p>You've been added to the <strong>{{.TeamName}}</strong> team in <strong>{{.WorkspaceName}}</strong>.</p>
            <p>Added by: <strong>{{.AddedBy}}</strong></p>

            <a href="{{.TeamURL}}" class="btn">View Team</a>
        </div>
        <div class="footer">
            <p>This email was sent from ORA Scrum</p>
        </div>
    </div>
</body>
</html>
`))

	// Sprint Started Template
	s.templates["sprint_started"] = template.Must(template.New("sprint_started").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); color: white; padding: 30px; border-radius: 10px 10px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 10px 10px; }
        .sprint-info { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; }
        .btn { display: inline-block; background: #f59e0b; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin-top: 15px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Sprint Started!</h1>
        </div>
        <div class="content">
            <p>Hi {{.UserName}},</p>
            <p>Sprint <strong>{{.SprintName}}</strong> has started!</p>

            <div class="sprint-info">
                <p><strong>Project:</strong> {{.ProjectName}}</p>
                <p><strong>Duration:</strong> {{.StartDate}} - {{.EndDate}}</p>
                {{if .Goal}}<p><strong>Sprint Goal:</strong><br/>{{.Goal}}</p>{{end}}
                <p><strong>Tasks:</strong> {{.TaskCount}} tasks</p>
            </div>

            <a href="{{.SprintURL}}" class="btn">View Sprint Board</a>
        </div>
        <div class="footer">
            <p>This email was sent from ORA Scrum</p>
        </div>
    </div>
</body>
</html>
`))

	// Due Date Reminder Template
	s.templates["due_date_reminder"] = template.Must(template.New("due_date_reminder").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%); color: white; padding: 30px; border-radius: 10px 10px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 10px 10px; }
        .task-list { background: white; border-radius: 8px; padding: 20px; margin: 20px 0; }
        .task-item { padding: 15px; border-bottom: 1px solid #e5e7eb; }
        .task-item:last-child { border-bottom: none; }
        .due-today { color: #ef4444; font-weight: bold; }
        .due-soon { color: #f59e0b; }
        .btn { display: inline-block; background: #ef4444; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin-top: 15px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⏰ Task Due Date Reminder</h1>
        </div>
        <div class="content">
            <p>Hi {{.UserName}},</p>
            <p>You have tasks that need your attention:</p>

            <div class="task-list">
                {{range .Tasks}}
                <div class="task-item">
                    <strong>{{.TaskKey}}</strong> - {{.TaskTitle}}<br/>
                    <span class="{{if eq .DaysRemaining 0}}due-today{{else}}due-soon{{end}}">
                        {{if eq .DaysRemaining 0}}Due Today!{{else}}Due in {{.DaysRemaining}} days{{end}}
                    </span>
                </div>
                {{end}}
            </div>

            <a href="{{.DashboardURL}}" class="btn">View My Tasks</a>
        </div>
        <div class="footer">
            <p>This email was sent from ORA Scrum</p>
        </div>
    </div>
</body>
</html>
`))

	// Comment Mention Template
	s.templates["mention"] = template.Must(template.New("mention").Parse(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%); color: white; padding: 30px; border-radius: 10px 10px 0 0; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 10px 10px; }
        .comment-box { background: white; border-left: 4px solid #3b82f6; padding: 20px; margin: 20px 0; border-radius: 0 8px 8px 0; }
        .btn { display: inline-block; background: #3b82f6; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin-top: 15px; }
        .footer { text-align: center; color: #6b7280; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>💬 You Were Mentioned</h1>
        </div>
        <div class="content">
            <p>Hi {{.UserName}},</p>
            <p><strong>{{.MentionedBy}}</strong> mentioned you in a comment on task <strong>{{.TaskKey}}</strong>:</p>

            <div class="comment-box">
                {{.CommentContent}}
            </div>

            <a href="{{.TaskURL}}" class="btn">View Comment</a>
        </div>
        <div class="footer">
            <p>This email was sent from ORA Scrum</p>
        </div>
    </div>
</body>
</html>
`))
}


// SendInvitation sends a generic invitation email (workspace/project/link)
func (s *Service) SendInvitation(
	workspaceName string,
	to string,
	invitedBy string,
	token string,
) error {

	if invitedBy == "" {
		invitedBy = "Someone"
	}

	inviteURL := fmt.Sprintf(
		"https://app.orascrum.com/invite?token=%s",
		token,
	)

	data := InvitationEmailData{
		WorkspaceName: workspaceName,
		InvitedBy:     invitedBy,
		InviteURL:     inviteURL,
	}

	return s.SendWithTemplate(
		[]string{to},
		fmt.Sprintf("[ORA] Invitation to join %s", workspaceName),
		"invitation",
		data,
	)
}


// Send sends an email
func (s *Service) Send(email *Email) error {
	if s.config.Host == "" {
		log.Println("Email not configured, skipping send")
		return nil
	}

	// Build message
	var msg bytes.Buffer

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))
	if len(email.CC) > 0 {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.CC, ", ")))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if email.HTMLBody != "" {
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(email.HTMLBody)
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(email.Body)
	}

	// Build recipient list
	recipients := append(email.To, email.CC...)
	recipients = append(recipients, email.BCC...)

	// Create auth
	auth := smtp.PlainAuth("", s.config.User, s.config.Password, s.config.Host)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	if s.config.UseTLS {
		// TLS connection
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS dial error: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.config.Host)
		if err != nil {
			return fmt.Errorf("SMTP client error: %w", err)
		}
		defer client.Close()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("auth error: %w", err)
		}

		if err = client.Mail(s.config.From); err != nil {
			return fmt.Errorf("mail error: %w", err)
		}

		for _, rcpt := range recipients {
			if err = client.Rcpt(rcpt); err != nil {
				return fmt.Errorf("rcpt error: %w", err)
			}
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("data error: %w", err)
		}

		_, err = w.Write(msg.Bytes())
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("close error: %w", err)
		}

		return client.Quit()
	}

	// Non-TLS
	return smtp.SendMail(addr, auth, s.config.From, recipients, msg.Bytes())
}

// SendWithTemplate sends an email using a template
func (s *Service) SendWithTemplate(to []string, subject, templateName string, data interface{}) error {
	tmpl, ok := s.templates[templateName]
	if !ok {
		return fmt.Errorf("template not found: %s", templateName)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("template execution error: %w", err)
	}

	return s.Send(&Email{
		To:       to,
		Subject:  subject,
		HTMLBody: body.String(),
	})
}

// ============================================
// Convenience Methods
// ============================================

// TaskAssignedData holds data for task assigned email
type TaskAssignedData struct {
	AssigneeName string
	AssignerName string
	TaskKey      string
	TaskTitle    string
	ProjectName  string
	Priority     string
	DueDate      string
	Description  string
	TaskURL      string
}

// SendTaskAssigned sends a task assigned email
func (s *Service) SendTaskAssigned(to string, data TaskAssignedData) error {
	return s.SendWithTemplate(
		[]string{to},
		fmt.Sprintf("[ORA] Task Assigned: %s - %s", data.TaskKey, data.TaskTitle),
		"task_assigned",
		data,
	)
}

// WorkspaceInvitationData holds data for workspace invitation email
type WorkspaceInvitationData struct {
	InviterName   string
	WorkspaceName string
	Role          string
	InviteURL     string
}

// ProjectInvitationData holds data for project invitation email
type ProjectInvitationData struct {
	InviterName   string
	ProjectName   string
	WorkspaceName string
	Role          string
	InviteURL     string
}

// SendProjectInvitation sends a project invitation email
func (s *Service) SendProjectInvitation(to string, data ProjectInvitationData) error {
	return s.SendWithTemplate(
		[]string{to},
		fmt.Sprintf("[ORA] Invitation to join project %s", data.ProjectName),
		"project_invitation",
		data,
	)
}

// SendWorkspaceInvitation sends a workspace invitation email
func (s *Service) SendWorkspaceInvitation(to string, data WorkspaceInvitationData) error {
	return s.SendWithTemplate(
		[]string{to},
		fmt.Sprintf("[ORA] Invitation to join %s", data.WorkspaceName),
		"workspace_invitation",
		data,
	)
}

// TeamInvitationData holds data for team invitation email
type TeamInvitationData struct {
	UserName      string
	TeamName      string
	WorkspaceName string
	AddedBy       string
	TeamURL       string
}

// SendTeamInvitation sends a team invitation email
func (s *Service) SendTeamInvitation(to string, data TeamInvitationData) error {
	return s.SendWithTemplate(
		[]string{to},
		fmt.Sprintf("[ORA] Added to team: %s", data.TeamName),
		"team_invitation",
		data,
	)
}

// SprintStartedData holds data for sprint started email
type SprintStartedData struct {
	UserName    string
	SprintName  string
	ProjectName string
	StartDate   string
	EndDate     string
	Goal        string
	TaskCount   int
	SprintURL   string
}

// SendSprintStarted sends a sprint started email
func (s *Service) SendSprintStarted(to string, data SprintStartedData) error {
	return s.SendWithTemplate(
		[]string{to},
		fmt.Sprintf("[ORA] Sprint Started: %s", data.SprintName),
		"sprint_started",
		data,
	)
}

// DueDateReminderTask holds task info for due date reminder
type DueDateReminderTask struct {
	TaskKey       string
	TaskTitle     string
	DaysRemaining int
}

// DueDateReminderData holds data for due date reminder email
type DueDateReminderData struct {
	UserName     string
	Tasks        []DueDateReminderTask
	DashboardURL string
}

// SendDueDateReminder sends a due date reminder email
func (s *Service) SendDueDateReminder(to string, data DueDateReminderData) error {
	return s.SendWithTemplate(
		[]string{to},
		"[ORA] Task Due Date Reminder",
		"due_date_reminder",
		data,
	)
}

// MentionData holds data for mention notification email
type MentionData struct {
	UserName       string
	MentionedBy    string
	TaskKey        string
	CommentContent string
	TaskURL        string
}

// SendMention sends a mention notification email
func (s *Service) SendMention(to string, data MentionData) error {
	return s.SendWithTemplate(
		[]string{to},
		fmt.Sprintf("[ORA] %s mentioned you in %s", data.MentionedBy, data.TaskKey),
		"mention",
		data,
	)
}

// ============================================
// Async Email Queue (simple in-memory)
// ============================================

// EmailQueue handles async email sending
type EmailQueue struct {
	service *Service
	queue   chan *queuedEmail
	done    chan bool
}

type queuedEmail struct {
	to           []string
	subject      string
	templateName string
	data         interface{}
	retries      int
}

// NewEmailQueue creates a new email queue
func NewEmailQueue(service *Service, workers int) *EmailQueue {
	q := &EmailQueue{
		service: service,
		queue:   make(chan *queuedEmail, 1000),
		done:    make(chan bool),
	}

	// Start workers
	for i := 0; i < workers; i++ {
		go q.worker()
	}

	return q
}

func (q *EmailQueue) worker() {
	for {
		select {
		case email := <-q.queue:
			err := q.service.SendWithTemplate(email.to, email.subject, email.templateName, email.data)
			if err != nil {
				log.Printf("Email send error: %v", err)
				// Retry logic
				if email.retries < 3 {
					email.retries++
					time.Sleep(time.Second * time.Duration(email.retries*2))
					q.queue <- email
				}
			}
		case <-q.done:
			return
		}
	}
}

// Enqueue adds an email to the queue
func (q *EmailQueue) Enqueue(to []string, subject, templateName string, data interface{}) {
	q.queue <- &queuedEmail{
		to:           to,
		subject:      subject,
		templateName: templateName,
		data:         data,
		retries:      0,
	}
}

// Stop stops the email queue workers
func (q *EmailQueue) Stop() {
	close(q.done)
}
