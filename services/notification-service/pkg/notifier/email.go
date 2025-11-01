package notifier

import (
	"fmt"
	"net/smtp"
)

type EmailNotifier struct {
	smtpHost string
	smtpPort string
	from     string
	username string
	password string
	to       string
}

func NewEmailNotifier(host, port, from, username, password, to string) Notifier {
	return &EmailNotifier{
		smtpHost: host,
		smtpPort: port,
		from:     from,
		password: password,
		to:       to,
	}
}

func (n *EmailNotifier) NotifyStarInrcease(repo string, oldStars, newStars int) error {
	subject := fmt.Sprintf("GitHub Stars Increased for %s", repo)
	body := fmt.Sprintf("Stars: %d -> %d (+%d)", oldStars, newStars, newStars-oldStars)

	msg := fmt.Sprintf("From %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", n.from, n.to, subject, body)
	auth := smtp.PlainAuth("", n.username, n.password, n.smtpHost)

	return smtp.SendMail(n.smtpHost+":"+n.smtpPort, auth, n.from, []string{n.to}, []byte(msg))
}
