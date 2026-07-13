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

	msg := "From: " + c.source + "\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: Confirm your email address\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		"<h2>Welcome!</h2><p>Please open the link below to verify your email.</p>\r\n" +
		link

	auth := smtp.PlainAuth("", c.source, c.password, c.host)
	addr := fmt.Sprintf("%s:%s", c.host, c.port)

	if err := smtp.SendMail(addr, auth, c.source, []string{toEmail}, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}
