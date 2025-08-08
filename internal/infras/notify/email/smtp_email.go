package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
)

type SMTPSender struct {
	addr   string
	auth   smtp.Auth
	from   string
	useTLS bool
}

func NewSMTPSender(host string, port int, username, password, from string, useTLS bool) *SMTPSender {
	var auth smtp.Auth
	if username != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	return &SMTPSender{
		addr:   fmt.Sprintf("%s:%d", host, port),
		auth:   auth,
		from:   from,
		useTLS: useTLS,
	}
}

func (s *SMTPSender) Send(ctx context.Context, to, subject, body string) error {
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n" +
		body + "\r\n")

	if s.useTLS {
		// Best-effort TLS: connect, then use smtp.NewClient
		conn, err := tls.Dial("tcp", s.addr, &tls.Config{InsecureSkipVerify: false})
		if err != nil {
			return err
		}
		defer conn.Close()
		c, err := smtp.NewClient(conn, s.addr)
		if err != nil {
			return err
		}
		defer c.Close()
		if s.auth != nil {
			if err := c.Auth(s.auth); err != nil {
				return err
			}
		}
		if err := c.Mail(s.from); err != nil {
			return err
		}
		if err := c.Rcpt(to); err != nil {
			return err
		}
		wc, err := c.Data()
		if err != nil {
			return err
		}
		if _, err := wc.Write(msg); err != nil {
			wc.Close()
			return err
		}
		return wc.Close()
	}
	return smtp.SendMail(s.addr, s.auth, s.from, []string{to}, msg)
}
