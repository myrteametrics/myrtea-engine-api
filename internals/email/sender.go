package email

import (
	"fmt"
	"net/smtp"
)

type Sender struct {
	auth     smtp.Auth
	username string
	host     string
	port     string
}

func NewSender(username string, password string, host string, port string) *Sender {
	auth := smtp.PlainAuth("", username, password, host)
	return &Sender{
		auth:     auth,
		username: username,
		host:     host,
		port:     port,
	}
}

func (s *Sender) Send(m Message) error {
	return smtp.SendMail(
		fmt.Sprintf("%s:%s", s.host, s.port),
		s.auth,
		s.username,
		m.To,
		m.ToBytes(),
	)
}

func (s *Sender) SendBytes(to []string, b []byte) error {
	return smtp.SendMail(
		fmt.Sprintf("%s:%s", s.host, s.port),
		s.auth,
		s.username,
		to,
		b,
	)
}
