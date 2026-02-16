package mailer

import (
	"fmt"
	"html"
	"net/smtp"
	"strings"
)

// Config holds SMTP connection settings.
type Config struct {
	Host         string // SMTP server hostname
	Port         string // SMTP server port (e.g. "587")
	Username     string // SMTP auth username
	Password     string // SMTP auth password
	FromAddress  string // From address for outgoing emails
	ContactEmail string // Recipient for contact form notifications
}

// Mailer sends emails via SMTP.
type Mailer struct {
	cfg  Config
	auth smtp.Auth
	addr string
}

// New creates a Mailer from the given config.
func New(cfg Config) *Mailer {
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	addr := cfg.Host + ":" + cfg.Port
	return &Mailer{cfg: cfg, auth: auth, addr: addr}
}

// sanitizeHeader removes \r and \n from a string to prevent email header injection.
func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

// send dispatches an HTML email.
func (m *Mailer) send(to, subject, htmlBody string) error {
	to = sanitizeHeader(to)
	subject = sanitizeHeader(subject)

	headers := []string{
		"From: " + m.cfg.FromAddress,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
	}
	msg := []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + htmlBody)
	return smtp.SendMail(m.addr, m.auth, m.cfg.FromAddress, []string{to}, msg)
}

// SendContactNotification sends the form data to the site owner.
func (m *Mailer) SendContactNotification(name, email, subject, message string) error {
	subjectLine := fmt.Sprintf("[Website Contact] %s", subject)

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#161618;color:#d8d8d8;font-family:Inter,Arial,sans-serif;">
  <div style="max-width:600px;margin:0 auto;padding:32px;">
    <h2 style="color:#e2b384;margin-top:0;">New Contact Form Submission</h2>
    <table style="width:100%%;border-collapse:collapse;">
      <tr>
        <td style="padding:8px 12px;color:#a0a0a0;vertical-align:top;white-space:nowrap;">Name</td>
        <td style="padding:8px 12px;color:#d8d8d8;">%s</td>
      </tr>
      <tr>
        <td style="padding:8px 12px;color:#a0a0a0;vertical-align:top;white-space:nowrap;">Email</td>
        <td style="padding:8px 12px;"><a href="mailto:%s" style="color:#e2b384;">%s</a></td>
      </tr>
      <tr>
        <td style="padding:8px 12px;color:#a0a0a0;vertical-align:top;white-space:nowrap;">Subject</td>
        <td style="padding:8px 12px;color:#d8d8d8;">%s</td>
      </tr>
      <tr>
        <td style="padding:8px 12px;color:#a0a0a0;vertical-align:top;white-space:nowrap;">Message</td>
        <td style="padding:8px 12px;color:#d8d8d8;white-space:pre-wrap;">%s</td>
      </tr>
    </table>
  </div>
</body>
</html>`,
		html.EscapeString(name),
		html.EscapeString(email),
		html.EscapeString(email),
		html.EscapeString(subject),
		html.EscapeString(message),
	)

	return m.send(m.cfg.ContactEmail, subjectLine, body)
}

// SendConfirmation sends an auto-reply to the person who submitted the form.
func (m *Mailer) SendConfirmation(toEmail, toName string) error {
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="margin:0;padding:0;background-color:#161618;color:#d8d8d8;font-family:Inter,Arial,sans-serif;">
  <div style="max-width:600px;margin:0 auto;padding:32px;">
    <h2 style="color:#e2b384;margin-top:0;">Thanks for reaching out, %s!</h2>
    <p style="line-height:1.6;color:#d8d8d8;">
      I've received your message and will get back to you as soon as possible.
    </p>
    <p style="line-height:1.6;color:#a0a0a0;font-size:14px;">
      This is an automated confirmation. Please do not reply to this email.
    </p>
  </div>
</body>
</html>`, html.EscapeString(toName))

	return m.send(toEmail, "Thanks for reaching out!", body)
}
