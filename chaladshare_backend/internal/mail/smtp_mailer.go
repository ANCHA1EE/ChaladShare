package mail

import (
	"fmt"
	"net/smtp"
)

type Mailer struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

func NewMailer(host string, port int, user, pass, from string) *Mailer {
	return &Mailer{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
		From: from,
	}
}

func (m *Mailer) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)

	msg := ""
	msg += fmt.Sprintf("From: %s\r\n", m.From)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	msg += "\r\n" + body

	auth := smtp.PlainAuth("", m.User, m.Pass, m.Host)
	return smtp.SendMail(addr, auth, m.User, []string{to}, []byte(msg))
}
