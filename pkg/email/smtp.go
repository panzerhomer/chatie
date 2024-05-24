package smtp

import (
	"github.com/go-gomail/gomail"
	"github.com/pkg/errors"
)

type SMTPServer struct {
	from string
	pass string
	host string
	port int
}

func NewSMTPSender(from, pass, host string, port int) (*SMTPServer, error) {
	if !IsEmailValid(from) {
		return nil, errors.New("invalid from email")
	}

	return &SMTPServer{from: from, pass: pass, host: host, port: port}, nil
}

func (s *SMTPServer) Send(input SendEmailInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", s.from)
	msg.SetHeader("To", input.To)
	msg.SetHeader("Subject", input.Subject)
	msg.SetBody("text/html", input.Body)

	dialer := gomail.NewDialer(s.host, s.port, s.from, s.pass)
	if err := dialer.DialAndSend(msg); err != nil {
		return errors.Wrap(err, "failed to sent email via smtp")
	}

	return nil
}
