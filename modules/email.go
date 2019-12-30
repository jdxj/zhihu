package modules

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	addr = "smtp.qq.com:587"
	host = "smtp.qq.com"
)

type EmailMsg struct {
	From    string
	To      string
	Subject string
	Content string
}

func NewEmailSender(config *EmailConfig) (*EmailSender, error) {
	if config == nil {
		return nil, fmt.Errorf("invalid email config")
	}

	es := &EmailSender{
		emailConfig: config,
		sender:      email.NewEmail(),
	}
	return es, nil
}

type EmailSender struct {
	emailConfig *EmailConfig
	sender      *email.Email
}

func (es *EmailSender) SendEmail(msg *EmailMsg) error {
	if msg == nil {
		return fmt.Errorf("invalid email message")
	}

	sender := es.sender
	config := es.emailConfig

	sender.From = msg.From
	sender.To = []string{msg.To}
	sender.Subject = msg.Subject
	sender.Text = []byte(msg.Content)

	return sender.Send(addr, smtp.PlainAuth("", config.User, config.Token, host))
}
