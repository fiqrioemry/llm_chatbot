package lib

import (
	"log"

	"server/internal/config"

	"gopkg.in/gomail.v2"
)

 
type Mailer struct {
	dialer *gomail.Dialer
	from   string
}

func NewMailer(cfg *config.Config) *Mailer {
	d := gomail.NewDialer(
		cfg.Mail.Host,
		cfg.Mail.Port,
		cfg.Mail.User,
		cfg.Mail.Password,
	)

	if conn, err := d.Dial(); err != nil {
		log.Printf("⚠️  Mailer dial failed (will retry on send): %v", err)
	} else {
		conn.Close()
		log.Println("✅ Mailer connected")
	}

	return &Mailer{dialer: d, from: cfg.Mail.From}
}

// SendVerificationLink mengirim email dengan link verifikasi / reset password.
func (m *Mailer) SendVerificationLink(to, subject, link string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", buildLinkEmailBody(link, subject))
	return m.dialer.DialAndSend(msg)
}

func buildLinkEmailBody(link, subject string) string {
	return `<!DOCTYPE html>
<html>
<body style="font-family:sans-serif;padding:32px;">
  <h2>` + subject + `</h2>
  <p>Click the link below to continue:</p>
  <a href="` + link + `" style="display:inline-block;padding:12px 24px;background:#4F46E5;color:#fff;border-radius:6px;text-decoration:none;">
    Open Link
  </a>
  <p style="margin-top:16px;color:#6B7280;font-size:12px;">
    This link expires in 24 hours. If you did not request this, please ignore this email.
  </p>
</body>
</html>`
}