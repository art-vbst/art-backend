package mailer

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/art-vbst/art-backend/internal/platform/config"
)

type Mailer interface {
	SendEmail(to, subject, body string) error
}

func New(config *config.Config) Mailer {
	return &Mailgun{config: config}
}

type Mailgun struct {
	config *config.Config
}

func (m *Mailgun) SendEmail(to, subject, body string) error {
	baseURL := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", m.config.MailgunDomain)

	data := url.Values{}
	data.Set("from", fmt.Sprintf("%s <noreply@%s>", m.config.EmailFromName, m.config.MailgunDomain))
	data.Set("to", m.getSafeTo(to))
	data.Set("subject", subject)
	data.Set("text", body)

	if config.IsDebug() {
		m.logEmail(to, subject, body)
		return nil
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("email new request err: %w", err)
	}

	req.SetBasicAuth("api", m.config.MailgunApiKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("email api request err: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (m *Mailgun) getSafeTo(intended string) string {
	if !config.IsDebug() {
		return intended
	}

	testEmail := m.config.TestEmail
	if testEmail == "" {
		log.Fatal("DEBUG mode enabled and no test email set")
	}

	return testEmail
}

func (m *Mailgun) logEmail(to, subject, body string) {
	log.Println("--------------------------------")
	log.Println("To: ", to)
	log.Println("Subject: ", subject)
	log.Println("Body: ", body)
	log.Println("--------------------------------")
}
