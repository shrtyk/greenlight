package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

type MailWriter interface {
	Send(recipient, templateFile string, data any) error
}

type MailData struct {
	UserName        string `json:"userName"`
	ActivationToken string `json:"activationToken"`
}

type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) MailWriter {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 10 * time.Second

	return &Mailer{
		dialer: dialer,
		sender: sender,
	}
}

func (m *Mailer) Send(recipient, templateFile string, data any) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(plainBody, "plainBody", data); err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data); err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/plain", htmlBody.String())

	if err = m.dialer.DialAndSend(msg); err != nil {
		return err
	}

	return nil
}

type MockMailer struct {
	tokens *[]MailData
	sender string
}

func NewMockMailer(tokens *[]MailData) MailWriter {
	return &MockMailer{
		tokens: tokens,
		sender: "mock sender",
	}
}

func (m *MockMailer) Send(recipient, templateFile string, data any) error {
	val, ok := data.(MailData)
	if !ok {
		return fmt.Errorf("wrong type")
	}
	*m.tokens = append(*m.tokens, val)
	return nil
}
