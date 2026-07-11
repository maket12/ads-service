package yandexmail

import (
	"context"
	"fmt"
	"net/smtp"
)

type SmtpClient struct {
	host                string
	port                string
	source              string
	password            string
	verificationBaseURL string
}

func NewSmtpClient(host, port, source, password, verfBaseURL string) *SmtpClient {
	return &SmtpClient{
		host:                host,
		port:                port,
		source:              source,
		password:            password,
		verificationBaseURL: verfBaseURL,
	}
}

func (c *SmtpClient) SendVerificationEmail(_ context.Context,
	toEmail, token string,
) error {
	link := fmt.Sprintf("%s?token=%s", c.verificationBaseURL, token)

	subject := "Subject: Confirm your email address\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf("<h2>Welcome!</h2><p>Please click <a href='%s'>here</a> to verify your email.</p>", link)

	msg := []byte(subject + mime + body)
	auth := smtp.PlainAuth("", c.source, c.password, c.host)

	addr := fmt.Sprintf("%s:%s", c.host, c.port)
	if err := smtp.SendMail(addr, auth, c.source, []string{toEmail}, msg); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}
