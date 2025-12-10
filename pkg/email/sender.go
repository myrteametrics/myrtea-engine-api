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
	var auth smtp.Auth
	if username != "" && password != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	return &Sender{
		auth:     auth,
		username: username,
		host:     host,
		port:     port,
	}
}

func (s *Sender) Send(m Message) error {
	m.From = s.username
	return smtp.SendMail(
		fmt.Sprintf("%s:%s", s.host, s.port),
		s.auth,
		s.username,
		append(m.To, m.CC...),
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
