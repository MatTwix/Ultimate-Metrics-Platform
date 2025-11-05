package notifier

import (
	"fmt"
	"net/smtp"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/services/notification-service/internal/metrics"
)

type EmailNotifier struct {
	smtpHost string
	smtpPort string
	from     string
	username string
	password string
	to       string
	metrics  *metrics.Metrics
}

func NewEmailNotifier(host, port, from, username, password, to string, metrics *metrics.Metrics) Notifier {
	return &EmailNotifier{
		smtpHost: host,
		smtpPort: port,
		from:     from,
		username: username,
		password: password,
		to:       to,
		metrics:  metrics,
	}
}

func (n *EmailNotifier) NotifyStarInrcease(repo string, oldStars, newStars int) error {
	start := time.Now()

	defer func() {
		duration := time.Since(start).Seconds()
		n.metrics.NotificationsSendDuration.WithLabelValues("email").Observe(duration)
	}()

	subject := fmt.Sprintf("GitHub Stars Increased for %s", repo)
	body := fmt.Sprintf("Stars: %d -> %d (+%d)", oldStars, newStars, newStars-oldStars)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", n.from, n.to, subject, body)
	auth := smtp.PlainAuth("", n.username, n.password, n.smtpHost)

	if err := smtp.SendMail(n.smtpHost+":"+n.smtpPort, auth, n.from, []string{n.to}, []byte(msg)); err != nil {
		n.metrics.NotificationsErrorsTotal.WithLabelValues("email").Inc()
		return fmt.Errorf("failed to send mail: %w", err)
	}

	n.metrics.NotificationsSentTotal.WithLabelValues("email").Inc()

	return nil
}
