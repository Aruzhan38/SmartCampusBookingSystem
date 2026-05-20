package mail

import (
	"fmt"
	"net/smtp"
)

type SMTPSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPSender(host, port, username, password, from string) *SMTPSender {
	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SMTPSender) Send(to, subject, body string) error {
	if s.username == "" || s.password == "" || s.from == "" || to == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}

	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	message := []byte(
		"From: " + s.from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
			"\r\n" +
			body,
	)

	addr := s.host + ":" + s.port
	return smtp.SendMail(addr, auth, s.from, []string{to}, message)
}
