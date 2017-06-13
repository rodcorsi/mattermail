package model

import (
	"fmt"
	"strings"
)

const defaultDebug = true

// Config type to parse config.json
type Config struct {
	Profiles []*Profile
	Debug    *bool `json:",omitempty"`
}

// Validate set default value for config and check if valid return err
func (c *Config) Validate() error {
	if c.Profiles == nil || len(c.Profiles) == 0 {
		return fmt.Errorf("Field 'Profiles' is empty set Profiles configuration")
	}

	for _, p := range c.Profiles {
		if err := p.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Fix fields and using default if is necessary
func (c *Config) Fix() {
	if c.Debug == nil {
		x := defaultDebug
		c.Debug = &x
	}

	for _, p := range c.Profiles {
		p.Fix()
	}
}

// MigrateFromV1 migrates config from version 1 to actual
func MigrateFromV1(v1 ConfigV1) *Config {
	config := &Config{}

	for _, c := range v1 {
		// Email
		email := &Email{
			ImapServer: c.ImapServer,
			Address:    c.Email,
			Password:   c.EmailPass,
		}

		if c.StartTLS != defaultStartTLS {
			email.StartTLS = &c.StartTLS
		}

		if c.TLSAcceptAllCerts != defaultTLSAcceptAllCerts {
			email.TLSAcceptAllCerts = &c.TLSAcceptAllCerts
		}

		// Mattermost
		mattermost := &Mattermost{
			Server:   c.Server,
			Team:     c.Team,
			Channels: []string{c.Channel},
			User:     c.MattermostUser,
			Password: c.MattermostPass,
		}

		// Profile
		profile := &Profile{
			Name:       c.Name,
			Email:      email,
			Mattermost: mattermost,
			Filter:     c.Filter,
		}

		mailtemplate := strings.Replace(c.MailTemplate, "%v", "{{.From}}", 1)
		mailtemplate = strings.Replace(mailtemplate, "%v", "{{.Subject}}", 1)
		mailtemplate = strings.Replace(mailtemplate, "%v", "{{.Message}}", 1)

		if mailtemplate != defaultMailTemplate {
			profile.MailTemplate = &mailtemplate
		}

		if c.LinesToPreview != defaultLinesToPreview {
			profile.LinesToPreview = &c.LinesToPreview
		}

		profile.RedirectChannel = &c.NoRedirectChannel
		*profile.RedirectChannel = !*profile.RedirectChannel

		if *profile.RedirectChannel == defaultRedirectChannel {
			profile.RedirectChannel = nil
		}

		profile.Attachment = &c.NoAttachment
		*profile.Attachment = !*profile.Attachment

		if *profile.Attachment == defaultAttachment {
			profile.Attachment = nil
		}

		if c.Disabled != defaultDisabled {
			profile.Disabled = &c.Disabled
		}

		config.Profiles = append(config.Profiles, profile)
	}

	return config
}
