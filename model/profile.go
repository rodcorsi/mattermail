package model

import (
	"bytes"
	"fmt"
	"text/template"
)

const (
	defaultMailTemplate    = ":incoming_envelope: _From: **{{.From}}**_\n>_{{.Subject}}_\n\n{{.Message}}"
	defaultLinesToPreview  = 10
	defaultRedirectChannel = true
	defaultAttachment      = true
	defaultDisabled        = false
)

// Profile type with general service settings
type Profile struct {
	Name            string
	MailTemplate    *string `json:",omitempty"`
	LinesToPreview  *int    `json:",omitempty"`
	RedirectChannel *bool   `json:",omitempty"`
	Attachment      *bool   `json:",omitempty"`
	Disabled        *bool   `json:",omitempty"`
	Email           *Email
	Mattermost      *Mattermost
	Filter          *Filter `json:",omitempty"`
}

// NewProfile creates new Profile
func NewProfile() *Profile {
	return &Profile{
		Email:      &Email{},
		Mattermost: &Mattermost{},
	}
}

// Validate set default value for config and check if valid return err
func (c *Profile) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("Field 'Name' is empty set a name for help in log")
	}

	if c.LinesToPreview != nil && *c.LinesToPreview <= 0 {
		return fmt.Errorf("Field 'LinesToPreview' need to be greater than 0")
	}

	if c.Email == nil {
		return fmt.Errorf("Field 'Email' is empty set Email configuration")
	}

	if c.Email != nil {
		if err := c.Email.Validate(); err != nil {
			return err
		}
	}

	if c.Mattermost == nil {
		return fmt.Errorf("Field 'Mattermost' is empty set Mattermost configuration")
	}

	if c.Mattermost != nil {
		if err := c.Mattermost.Validate(); err != nil {
			return err
		}
	}

	if c.Filter != nil {
		if err := c.Filter.Validate(); err != nil {
			return fmt.Errorf("Error in Filter:%v", err)
		}
	}

	return nil
}

// Fix fields and using default if is necessary
func (c *Profile) Fix() {
	if c.MailTemplate == nil {
		x := defaultMailTemplate
		c.MailTemplate = &x
	}
	if c.LinesToPreview == nil {
		x := defaultLinesToPreview
		c.LinesToPreview = &x
	}
	if c.RedirectChannel == nil {
		x := defaultRedirectChannel
		c.RedirectChannel = &x
	}
	if c.Attachment == nil {
		x := defaultAttachment
		c.Attachment = &x
	}
	if c.Disabled == nil {
		x := defaultDisabled
		c.Disabled = &x
	}

	c.Email.Fix()
	c.Mattermost.Fix()

	if c.Filter != nil {
		c.Filter.Fix()
	}
}

// FormatMailTemplate formats MailTemplate using fields
func (c *Profile) FormatMailTemplate(from, subject, message string) (string, error) {
	t, err := template.New("").Parse(*c.MailTemplate)
	if err != nil {
		return "", fmt.Errorf("Error on parse MailTemplate %v err:%v", c.MailTemplate, err.Error())
	}

	r := &struct {
		From    string
		Subject string
		Message string
	}{
		From:    from,
		Subject: subject,
		Message: message,
	}

	buff := &bytes.Buffer{}
	if err = t.Execute(buff, r); err != nil {
		return "", fmt.Errorf("Error on execute MailTemplate %v err:%v", c.MailTemplate, err.Error())
	}

	return buff.String(), nil
}