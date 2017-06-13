package model

import (
	"bytes"
	"fmt"
	"strings"
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
	Channels        []string
	MailTemplate    *string `json:",omitempty"`
	LinesToPreview  *int    `json:",omitempty"`
	RedirectChannel *bool   `json:",omitempty"`
	Attachment      *bool   `json:",omitempty"`
	Disabled        *bool   `json:",omitempty"`
	Email           *Email
	Mattermost      *Mattermost
	Filter          *Filter `json:",omitempty"`
}

// NewProfile creates new Profile with default values
func NewProfile() *Profile {
	profile := &Profile{
		MailTemplate:    new(string),
		LinesToPreview:  new(int),
		RedirectChannel: new(bool),
		Attachment:      new(bool),
		Disabled:        new(bool),
		Email:           NewEmail(),
		Mattermost:      NewMattermost(),
	}
	*profile.MailTemplate = defaultMailTemplate
	*profile.LinesToPreview = defaultLinesToPreview
	*profile.RedirectChannel = defaultRedirectChannel
	*profile.Attachment = defaultAttachment
	*profile.Disabled = defaultDisabled

	return profile
}

// Validate set default value for config and check if valid return err
func (c *Profile) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("Field 'Name' is empty set a name for help in log")
	}

	if len(c.Channels) == 0 {
		return fmt.Errorf("Field 'Channels' need to set at least one channel or user for destination")
	}

	for _, channel := range c.Channels {
		if channel != "" && !validateChannel(channel) {
			return fmt.Errorf("Field 'Channels' contains invalid chars, make sure if you are using url channel name or username. This field need to start with # for channel or @ for username: %v", channel)
		}
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
