package actions

import (
	"net/smtp"
)

type SMTPMail struct {
	SMTPHost string
	SMTPPort string
	SendFrom string
	SendTo string
	Subject string
	Password string
}

func (mail SMTPMail) SendMail(message string) (error) {
	auth := smtp.PlainAuth("", mail.SendFrom, mail.Password, mail.SMTPHost)
	byte_message := []byte("Subject:"+mail.Subject+"\r\n\r\n"+message)
	err := smtp.SendMail(
		mail.SMTPHost + ":" + mail.SMTPPort,
		auth,
		mail.SendFrom,
		[]string{mail.SendTo},
		byte_message,
	)
	return err
}

func (mail SMTPMail) Act(message string) error {
	err := mail.SendMail(message)
	return err
}