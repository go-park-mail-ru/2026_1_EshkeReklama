package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"eshkere/internal/config"
)

type NotificationSender interface {
	SendBalanceAlert(ctx context.Context, to string, subject string, body string) error
}

type SMTPNotificationSender struct {
	host     string
	port     int
	user     string
	password string
}

func NewSMTPNotificationSender(cfg config.SMTPConfig) *SMTPNotificationSender {
	return &SMTPNotificationSender{
		host:     cfg.Host,
		port:     cfg.Port,
		user:     cfg.User,
		password: cfg.Password,
	}
}

func (s *SMTPNotificationSender) Enabled() bool {
	return s != nil && s.host != "" && s.port > 0 && s.user != "" && s.password != ""
}

func (s *SMTPNotificationSender) SendBalanceAlert(ctx context.Context, to string, subject string, body string) error {
	if !s.Enabled() {
		return fmt.Errorf("smtp is not configured")
	}

	address := fmt.Sprintf("%s:%d", s.host, s.port)
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{
		ServerName: s.host,
	})
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer client.Close()

	auth := smtp.PlainAuth("", s.user, s.password, s.host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err := client.Mail(s.user); err != nil {
		return fmt.Errorf("smtp from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}

	message := fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", to, subject, body)
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write smtp message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close smtp writer: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp quit: %w", err)
	}
	return nil
}
