// Package email provides email sending functionality used by the invitation system and other parts of the app.
package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Config holds SMTP/email configuration.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	FromName string
	UseTLS   bool // implicit TLS (SMTPS) when true, otherwise STARTTLS when available
	Debug    bool // when true, print emails instead of sending
	Timeout  time.Duration
	BaseURL  string // used to build invitation links in templates
}

// Service is the main email service used across the app.
type Service struct {
	conf      *Config
	templates map[string]*template.Template
	queue     *EmailQueue
	sender    EmailSender
}

// Email represents a single email message.
type Email struct {
	To       []string
	CC       []string
	BCC      []string
	Subject  string
	Body     string // plain text
	HTMLBody string // html version (preferred)
}

// EmailSender is a small interface allowing pluggable senders (SMTP, provider API, debug).
type EmailSender interface {
	Send(ctx context.Context, from, to string, subject, plainBody, htmlBody string) error
}

// NewService creates and configures the email Service. config.From and Host/Port must be set for sending;
// if config.Debug is true and SMTP not configured, service will operate in debug mode (printing emails).
func NewService(conf *Config) (*Service, error) {
	if conf == nil {
		return nil, errors.New("email config required")
	}
	if strings.TrimSpace(conf.From) == "" {
		return nil, errors.New("email From address required")
	}
	if conf.Timeout == 0 {
		conf.Timeout = 10 * time.Second
	}

	s := &Service{
		conf:      conf,
		templates: make(map[string]*template.Template),
	}

	// choose sender
	if conf.Debug {
		s.sender = &debugSender{}
	} else if conf.Host != "" && conf.Port != 0 {
		s.sender = &smtpSender{
			cfg: SMTPConfig{
				Host:     conf.Host,
				Port:     conf.Port,
				Username: conf.User,
				Password: conf.Password,
				UseTLS:   conf.UseTLS,
				Timeout:  conf.Timeout,
			},
			timeout: conf.Timeout,
		}
	} else {
		// fallback to debug sender if no smtp configured
		s.sender = &debugSender{}
	}

	// load built-in templates
	s.loadTemplates()

	// optional queue (small default workers)
	s.queue = NewEmailQueue(s, 2)

	return s, nil
}

// Close stops the internal queue workers.
func (s *Service) Close() {
	if s.queue != nil {
		s.queue.Stop()
	}
}

// BuildInviteURL builds a front-end link for acceptance using BaseURL.
func (s *Service) BuildInviteURL(token string) string {
	base := strings.TrimRight(s.conf.BaseURL, "/")
	if base == "" {
		// default path if BaseURL not provided
		return fmt.Sprintf("/invitations/accept?token=%s", token)
	}
	return fmt.Sprintf("%s/invitations/accept?token=%s", base, token)
}

// SendInvitation composes and sends (best-effort) an invitation email.
// NOTE: signature matches calls from invitation service: (workspaceName, toEmail, fromName, token string) error
func (s *Service) SendInvitation(workspaceName, toEmail, fromName, token string) error {
	if strings.TrimSpace(toEmail) == "" {
		return errors.New("toEmail required")
	}
	// Use built-in workspace_invitation template
	data := map[string]interface{}{
		"InviterName":   fromName,
		"WorkspaceName": workspaceName,
		"Role":          "",
		"InviteURL":     s.BuildInviteURL(token),
	}
	subject := fmt.Sprintf("[ORA] Invitation to join %s", workspaceName)
	// Use context with configured timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.conf.Timeout)
	defer cancel()
	return s.sendWithTemplateContext(ctx, []string{toEmail}, subject, "workspace_invitation", data)
}

// Send is a low-level convenience to send an email immediately (synchronous).
func (s *Service) Send(ctx context.Context, email *Email) error {
	if email == nil {
		return errors.New("email required")
	}
	plain := email.Body
	html := email.HTMLBody
	// pick a single "to" for the lower-level sender interface; sender.Send handles single recipient
	if len(email.To) > 0 {
	} else {
		return errors.New("at least one recipient required")
	}
	// Use context with service timeout if ctx has no closer deadline
	sendCtx := ctx
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > s.conf.Timeout {
		var cancel context.CancelFunc
		sendCtx, cancel = context.WithTimeout(ctx, s.conf.Timeout)
		defer cancel()
	}

	// For simplicity we send to each recipient individually (best-effort).
	recipients := append([]string{}, email.To...)
	recipients = append(recipients, email.CC...)
	recipients = append(recipients, email.BCC...)

	var lastErr error
	for _, rcpt := range recipients {
		if err := s.sender.Send(sendCtx, fmt.Sprintf("%s <%s>", s.conf.FromName, s.conf.From), rcpt, email.Subject, plain, html); err != nil {
			lastErr = err
			log.Printf("email send to %s failed: %v", rcpt, err)
		}
	}
	return lastErr
}

// SendWithTemplate renders a template and sends the resulting HTML to the provided recipients.
func (s *Service) SendWithTemplate(to []string, subject, templateName string, data interface{}) error {
	return s.sendWithTemplateContext(context.Background(), to, subject, templateName, data)
}

func (s *Service) sendWithTemplateContext(ctx context.Context, to []string, subject, templateName string, data interface{}) error {
	tmpl, ok := s.templates[templateName]
	if !ok {
		return fmt.Errorf("template not found: %s", templateName)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("template execute: %w", err)
	}

	email := &Email{
		To:       to,
		Subject:  subject,
		HTMLBody: buf.String(),
		Body:     stripHTMLForPlain(buf.String()),
	}
	return s.Send(ctx, email)
}

// EnqueueWithTemplate enqueues an email to be sent asynchronously using the service's queue.
func (s *Service) EnqueueWithTemplate(to []string, subject, templateName string, data interface{}) {
	if s.queue == nil {
		_ = s.SendWithTemplate(to, subject, templateName, data)
		return
	}
	s.queue.Enqueue(to, subject, templateName, data)
}

// ========== Templates (use the templates you already have) ==========
func (s *Service) loadTemplates() {
	// load the templates you provided into s.templates.
	// For brevity here, load the workspace_invitation template and simple placeholders.
	s.templates["workspace_invitation"] = template.Must(template.New("workspace_invitation").Parse(`
<!DOCTYPE html>
<html>
<body>
  <h1>🎉 Workspace Invitation</h1>
  <p><strong>{{.InviterName}}</strong> invited you to join <strong>{{.WorkspaceName}}</strong>.</p>
  <p><a href="{{.InviteURL}}">Accept Invitation</a></p>
</body>
</html>
`))

	// (You should copy the full templates from your earlier snippet into s.templates["..."] as needed)
}

// stripHTMLForPlain attempts to produce a plain-text fallback from HTML content.
func stripHTMLForPlain(html string) string {
	plain := strings.ReplaceAll(html, "\n", " ")
	plain = strings.ReplaceAll(plain, "<br>", "\n")
	plain = strings.ReplaceAll(plain, "<br/>", "\n")
	plain = strings.ReplaceAll(plain, "<br />", "\n")
	for {
		start := strings.Index(plain, "<")
		if start == -1 {
			break
		}
		end := strings.Index(plain[start:], ">")
		if end == -1 {
			break
		}
		plain = plain[:start] + plain[start+end+1:]
	}
	return strings.TrimSpace(plain)
}

// ========== Default SMTP sender implementation ==========

type smtpSender struct {
	cfg     SMTPConfig
	timeout time.Duration
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
	Timeout  time.Duration
}

func (s *smtpSender) Send(ctx context.Context, from, to, subject, plainBody, htmlBody string) error {
	boundary := "INVITE-MSG-BOUNDARY"
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", to))
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary))
	b.WriteString("\r\n")
	b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(plainBody + "\r\n")
	b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	b.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(htmlBody + "\r\n")
	b.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	msg := b.Bytes()
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	// if implicit TLS (SMTPS)
	if s.cfg.UseTLS {
		dialer := &net.Dialer{Timeout: s.timeout}
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: s.cfg.Host})
		if err != nil {
			return fmt.Errorf("tls dial failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.cfg.Host)
		if err != nil {
			return fmt.Errorf("smtp client failed: %w", err)
		}
		defer client.Quit()

		if s.cfg.Username != "" && s.cfg.Password != "" {
			auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
			if ok, _ := client.Extension("AUTH"); ok {
				if err := client.Auth(auth); err != nil {
					return fmt.Errorf("auth failed: %w", err)
				}
			}
		}

		fromAddr := from
		if strings.Contains(from, "<") && strings.Contains(from, ">") {
			start := strings.Index(from, "<")
			end := strings.Index(from, ">")
			if start >= 0 && end > start {
				fromAddr = strings.TrimSpace(from[start+1 : end])
			}
		}
		if err := client.Mail(fromAddr); err != nil {
			return fmt.Errorf("mail from failed: %w", err)
		}
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("rcpt failed: %w", err)
		}
		wc, err := client.Data()
		if err != nil {
			return fmt.Errorf("data failed: %w", err)
		}
		if _, err := wc.Write(msg); err != nil {
			_ = wc.Close()
			return fmt.Errorf("write failed: %w", err)
		}
		if err := wc.Close(); err != nil {
			return fmt.Errorf("close failed: %w", err)
		}
		return nil
	}

	// Non-TLS: Dial and use STARTTLS when available
	var conn net.Conn
	var err error
	if dl := s.timeout; dl == 0 {
		dl = 10 * time.Second
	}
	dl := s.timeout
	dialer := &net.Dialer{Timeout: dl}
	conn, err = dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}
	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("smtp new client failed: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.cfg.Host}); err != nil {
			return fmt.Errorf("starttls failed: %w", err)
		}
	}

	if s.cfg.Username != "" && s.cfg.Password != "" {
		auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("smtp auth failed: %w", err)
			}
		}
	}

	fromAddr := from
	if strings.Contains(from, "<") && strings.Contains(from, ">") {
		start := strings.Index(from, "<")
		end := strings.Index(from, ">")
		if start >= 0 && end > start {
			fromAddr = strings.TrimSpace(from[start+1 : end])
		}
	}

	if err := client.Mail(fromAddr); err != nil {
		return fmt.Errorf("mail from failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt failed: %w", err)
	}
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("data failed: %w", err)
	}
	if _, err := wc.Write(msg); err != nil {
		_ = wc.Close()
		return fmt.Errorf("write failed: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}
	_ = client.Quit()
	return nil
}

// ========== Debug sender ==========

type debugSender struct{}

func (d *debugSender) Send(ctx context.Context, from, to, subject, plainBody, htmlBody string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	log.Printf("DEBUG EMAIL\nFrom: %s\nTo: %s\nSubject: %s\n\n%s\n\n(HTML body omitted)\n", from, to, subject, plainBody)
	return nil
}

// ========== Simple in-memory async queue (same as before) ==========

type queuedEmail struct {
	to           []string
	subject      string
	templateName string
	data         interface{}
	retries      int
}

type EmailQueue struct {
	service *Service
	queue   chan *queuedEmail
	done    chan struct{}
}

func NewEmailQueue(service *Service, workers int) *EmailQueue {
	q := &EmailQueue{
		service: service,
		queue:   make(chan *queuedEmail, 1000),
		done:    make(chan struct{}),
	}
	if workers <= 0 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		go q.worker()
	}
	return q
}

func (q *EmailQueue) worker() {
	for {
		select {
		case e := <-q.queue:
			if e == nil {
				continue
			}
			err := q.service.SendWithTemplate(e.to, e.subject, e.templateName, e.data)
			if err != nil {
				log.Printf("email send error: %v (template=%s)", err, e.templateName)
				if e.retries < 3 {
					e.retries++
					time.Sleep(time.Second * time.Duration(e.retries*2))
					select {
					case q.queue <- e:
					default:
						log.Printf("email queue full, dropping email (subject=%s)", e.subject)
					}
				}
			}
		case <-q.done:
			return
		}
	}
}

func (q *EmailQueue) Enqueue(to []string, subject, templateName string, data interface{}) {
	select {
	case q.queue <- &queuedEmail{to: to, subject: subject, templateName: templateName, data: data}:
	default:
		log.Printf("email queue full, dropping email (subject=%s)", subject)
	}
}

func (q *EmailQueue) Stop() {
	close(q.done)
}