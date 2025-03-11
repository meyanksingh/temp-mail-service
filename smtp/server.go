package smtp

import (
	"io"
	"net"
	"time"

	"github.com/DusanKasan/parsemail"
	"github.com/emersion/go-smtp"
	"github.com/meyanksingh/smtp-server/logger"
	"github.com/meyanksingh/smtp-server/models"

	"errors"
	"strings"

	db "github.com/meyanksingh/smtp-server/db"
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

	email, err := parsemail.Parse(r) // Use r directly as the reader
	if err != nil {
		logger.Error("[SMTP] Error parsing email: %v", err)
		return err
	}

	fromAddrs := make([]string, len(email.From))
	for i, addr := range email.From {
		fromAddrs[i] = addr.String()
	}

	toAddrs := make([]string, len(email.To))
	for i, addr := range email.To {
		toAddrs[i] = addr.String()
	}

	logger.Info("[SMTP] Email parsed - Subject: %s, From: %s, To: %s",
		email.Subject,
		strings.Join(fromAddrs, ","),
		strings.Join(toAddrs, ","))

	message := models.Message{
		From:    strings.Join(fromAddrs, ","),
		To:      strings.Join(toAddrs, ","),
		Subject: email.Subject,
		Body:    email.HTMLBody,
	}

	if db.DB.Create(&message).Error != nil {
		logger.Error("[SMTP] Failed to save message to database")
		return errors.New("failed to create message")
	}

	logger.Info("[SMTP] Email successfully saved to database (ID: %d)", message.ID)
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
