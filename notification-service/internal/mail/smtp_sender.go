package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/smtp"
)

type Sender interface {
	SendEmail(ctx context.Context, recipientEmail, subject, body string) error
}

type SMTPSender struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func NewSMTPSender(host, port, username, password, from string) Sender {
	return &SMTPSender{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func (s *SMTPSender) SendEmail(ctx context.Context, recipientEmail, subject, body string) error {
	if recipientEmail == "" {
		return fmt.Errorf("recipient email is required")
	}
	if s.From == "" {
		return fmt.Errorf("sender email is required")
	}

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", s.From, recipientEmail, subject, body)
	addr := net.JoinHostPort(s.Host, s.Port)

	var auth smtp.Auth
	if s.Username != "" || s.Password != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}

	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: s.Host}
		if err := client.StartTLS(config); err != nil {
			return err
		}
	}

	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err := client.Mail(s.From); err != nil {
		return err
	}
	if err := client.Rcpt(recipientEmail); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	if _, err := io.WriteString(w, message); err != nil {
		return err
	}

	return client.Quit()
}
