package model

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

const (
	defaultMailTemplate      = ":incoming_envelope: _From: **{{.From}}**_\n>_{{.Subject}}_\n\n{{.Message}}"
	defaultLinesToPreview    = 10
	defaultRedirectBySubject = true
	defaultAttachment        = true
	defaultDisabled          = false
)

// Profile type with general service settings
type Profile struct {
	Name              string
	Channels          []string
	MailTemplate      *string `json:",omitempty"`
	LinesToPreview    *int    `json:",omitempty"`
	RedirectBySubject *bool   `json:",omitempty"`
	Attachment        *bool   `json:",omitempty"`
	Disabled          *bool   `json:",omitempty"`
	Email             *Email
	Mattermost        *Mattermost
	Filter            *Filter `json:",omitempty"`
}

// NewProfile creates new Profile with default values
func NewProfile() *Profile {
	profile := &Profile{
		MailTemplate:      new(string),
		LinesToPreview:    new(int),
		RedirectBySubject: new(bool),
		Attachment:        new(bool),
		Disabled:          new(bool),
		Email:             NewEmail(),
		Mattermost:        NewMattermost(),
	}
	*profile.MailTemplate = defaultMailTemplate
	*profile.LinesToPreview = defaultLinesToPreview
	*profile.RedirectBySubject = defaultRedirectBySubject
	*profile.Attachment = defaultAttachment
	*profile.Disabled = defaultDisabled

	return profile
}

// Validate set default value for config and check if valid return err
func (c *Profile) Validate() error {
	if c.Name == "" {
		return errors.New("Field 'Name' is empty set a name for help in log")
	}

	if len(c.Channels) == 0 {
		return errors.New("Field 'Channels' need to set at least one channel or user for destination")
	}

	for _, channel := range c.Channels {
		if channel != "" && !validateChannel(channel) {
			return errors.Errorf("Field 'Channels' contains invalid chars, make sure if you are using url channel name or username. This field need to start with # for channel or @ for username: %v", channel)
		}
	}

	if c.LinesToPreview != nil && *c.LinesToPreview <= 0 {
		return errors.New("Field 'LinesToPreview' need to be greater than 0")
	}

	if c.Email == nil {
		return errors.New("Field 'Email' is empty set Email configuration")
	}

	if c.Email != nil {
		if err := c.Email.Validate(); err != nil {
			return err
		}
	}

	if c.Mattermost == nil {
		return errors.New("Field 'Mattermost' is empty set Mattermost configuration")
	}

	if c.Mattermost != nil {
		if err := c.Mattermost.Validate(); err != nil {
			return err
		}
	}

	if c.Filter != nil {
		if err := c.Filter.Validate(); err != nil {
			return errors.Errorf("Error in Filter:%v", err)
		}
	}

	return nil
}

// Fix fields and using default if is necessary
func (c *Profile) Fix() {
	for i, channel := range c.Channels {
		channel = strings.TrimSpace(channel)
		channel = strings.ToLower(channel)

		if !strings.HasPrefix(channel, "#") && !strings.HasPrefix(channel, "@") {
			channel = "#" + channel
		}
		c.Channels[i] = channel
	}

	if c.MailTemplate == nil {
		x := defaultMailTemplate
		c.MailTemplate = &x
	}
	if c.LinesToPreview == nil {
		x := defaultLinesToPreview
		c.LinesToPreview = &x
	}
	if c.RedirectBySubject == nil {
		x := defaultRedirectBySubject
		c.RedirectBySubject = &x
	}
	if c.Attachment == nil {
		x := defaultAttachment
		c.Attachment = &x
	}
	if c.Disabled == nil {
		x := defaultDisabled
		c.Disabled = &x
	}

	if c.Email != nil {
		c.Email.Fix()
	}

	if c.Mattermost != nil {
		c.Mattermost.Fix()
	}

	if c.Filter != nil {
		c.Filter.Fix()
	}
}

// FormatMailTemplate formats MailTemplate using fields
func (c *Profile) FormatMailTemplate(from, subject, message string) (string, error) {
	t, err := template.New("").Parse(*c.MailTemplate)
	if err != nil {
		return "", errors.Wrapf(err, "parse MailTemplate %v", c.MailTemplate)
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
		return "", errors.Wrapf(err, "Error on execute MailTemplate %v", c.MailTemplate)
	}

	return buff.String(), nil
}
