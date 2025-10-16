package template

import (
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/email"
)

// Template represents an email template stored in the database
type Template struct {
	Id          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Subject     string `json:"subject" db:"subject"`
	BodyHTML    string `json:"bodyHtml" db:"body_html"`
}

// NewTemplate creates a new Template instance
func NewTemplate(id int64, name string, subject string, bodyHTML string, description string) Template {
	return Template{
		Id:          id,
		Name:        name,
		Subject:     subject,
		BodyHTML:    bodyHTML,
		Description: description,
	}
}

// Validate checks if the template is valid
func (t Template) Validate() error {
	if t.Name == "" {
		return errors.New("template name is required")
	}
	if t.Subject == "" {
		return errors.New("template subject is required")
	}
	if t.BodyHTML == "" {
		return errors.New("template body is required")
	}
	return nil
}

// ToMessage creates an email message from the template
func (t Template) ToMessage() email.Message {
	return email.NewMessage(t.Subject, "text/html", t.BodyHTML)
}
