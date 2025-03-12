package smtp

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"net/mail"

	"github.com/emersion/go-smtp"
	"github.com/meyanksingh/smtp-server/logger"
	"github.com/meyanksingh/smtp-server/models"
	"github.com/meyanksingh/smtp-server/redis"

	"errors"
)

type MailHandler struct{}

func (h *MailHandler) Login(state *smtp.Conn, username, password string) (smtp.Session, error) {
	logger.Info("[SMTP] Login attempt with username: %s", username)
	return &MailSession{}, nil
}

func (h *MailHandler) AnonymousLogin(state *smtp.Conn) (smtp.Session, error) {
	logger.Info("[SMTP] Anonymous login from: %s", state.Hostname())
	return &MailSession{}, nil
}

func (h *MailHandler) NewSession(_ *smtp.Conn) (smtp.Session, error) {
	logger.Debug("[SMTP] New session created")
	return &MailSession{}, nil
}

type MailSession struct {
	from string
	to   []string
}

func (s *MailSession) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	logger.Info("[SMTP] MAIL FROM: %s", from)
	return nil
}

func (s *MailSession) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	logger.Info("[SMTP] RCPT TO: %s", to)
	return nil
}

func (s *MailSession) Data(r io.Reader) error {
	logger.Info("[SMTP] Receiving email data from %s to %v", s.from, s.to)

	email, err := mail.ReadMessage(r)
	if err != nil {
		logger.Error("[SMTP] Error parsing email: %v", err)
		return err
	}

	bodyBytes, err := io.ReadAll(email.Body)
	if err != nil {
		logger.Error("[SMTP] Error reading email body: %v", err)
		return err
	}
	bodyStr := string(bodyBytes)

	from := email.Header.Get("From")
	to := email.Header.Get("To")
	subject := email.Header.Get("Subject")

	logger.Info("[SMTP] Email parsed - Subject: %s, From: %s, To: %s", subject, from, to)

	message := models.Message{
		From:    strings.ToLower(from),
		To:      strings.ToLower(to),
		Subject: subject,
		Body:    bodyStr,
	}

	messageStr := fmt.Sprintf("From:%s\nTo:%s\nSubject:%s\nBody:%s", message.From, message.To, message.Subject, message.Body)

	ctx := context.Background()
	err = redis.RedisClient.LPush(ctx, message.To, messageStr).Err()
	if err != nil {
		logger.Error("[SMTP] Failed to save message to Redis: %v", err)
		return errors.New("failed to store message")
	}

	redis.RedisClient.Expire(ctx, message.To, 20*time.Minute)

	logger.Info("[SMTP] Email successfully saved to Redis for recipient: %s", message.To)
	return nil
}

func (s *MailSession) Reset() {
	logger.Debug("[SMTP] Session reset")
	s.from = ""
	s.to = nil
}

func (s *MailSession) Logout() error {
	logger.Debug("[SMTP] Session logout")
	return nil
}

func StartSMTPServer(host, port string) {
	logger.Info("Initializing SMTP server on %s:%s", host, port)

	backend := &MailHandler{}
	server := smtp.NewServer(backend)

	server.Addr = ":" + port
	server.Domain = host
	server.AllowInsecureAuth = true
	server.ReadTimeout = 10 * time.Second
	server.WriteTimeout = 10 * time.Second
	server.MaxMessageBytes = 10 * 1024 * 1024
	server.MaxRecipients = 50

	logger.Info("SMTP server configured with domain: %s", server.Domain)
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		logger.Fatal("Failed to start SMTP server: %v", err)
	}
	defer listener.Close()

	logger.Info("SMTP server started and listening on %s:%s", host, port)
	err = server.Serve(listener)
	if err != nil {
		logger.Fatal("SMTP server stopped: %v", err)
	}
}
