package model

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const (
	defaultDebug     = true
	defaultDirectory = "./data/"
)

// Config type to parse config.json
type Config struct {
	Directory string
	Debug     *bool `json:",omitempty"`
	Profiles  []*Profile
}

// NewConfig creates new Config with default values
func NewConfig() *Config {
	config := &Config{
		Debug:     new(bool),
		Directory: defaultDirectory,
	}
	*config.Debug = defaultDebug

	return config
}

// NewConfigFromFile loads config from json file
func NewConfigFromFile(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not load: %v", file)
	}

	config := NewConfig()
	if err = json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "read file '%v'", file)
	}

	config.Fix()
	return config, nil
}

// Validate set default value for config and check if valid return err
func (c *Config) Validate() error {
	if _, err := os.Stat(c.Directory); err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("Directory %v does not exists. please create the directory first", c.Directory)
		}
		return errors.Wrapf(err, "Field 'Directory':'%v' is not a valid path", c.Directory)
	}

	if c.Profiles == nil || len(c.Profiles) == 0 {
		return errors.Errorf("Field 'Profiles' is empty set Profiles configuration")
	}

	for _, p := range c.Profiles {
		if err := p.Validate(); err != nil {
			return errors.Wrap(err, "Validate config file")
		}
	}

	return nil
}

// Fix fields and using default if is necessary
func (c *Config) Fix() {
	if c.Directory == "" {
		c.Directory = defaultDirectory
	}

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
	config := &Config{
		Directory: defaultDirectory,
	}

	for _, c := range v1 {
		// Email
		email := &Email{
			ImapServer: c.ImapServer,
			Username:   c.Email,
			Password:   c.EmailPass,
		}

		if c.StartTLS != defaultStartTLS {
			email.StartTLS = new(bool)
			*email.StartTLS = c.StartTLS
		}

		if c.TLSAcceptAllCerts != defaultTLSAcceptAllCerts {
			email.TLSAcceptAllCerts = new(bool)
			*email.TLSAcceptAllCerts = c.TLSAcceptAllCerts
		}

		// Mattermost
		mattermost := &Mattermost{
			Server:   c.Server,
			Team:     c.Team,
			User:     c.MattermostUser,
			Password: c.MattermostPass,
		}

		// Filter
		filter := &Filter{} 
		for _, r := range (*c.Filter){
			*filter = append(*filter, &Rule{
				From:     r.From,
				Subject:  r.Subject,
				Channels: []string{r.Channel},
			})
		}
		// Profile
		profile := &Profile{
			Name:       c.Name,
			Channels:   []string{c.Channel},
			Email:      email,
			Mattermost: mattermost,
			Filter:     filter,
		}

		mailtemplate := strings.Replace(c.MailTemplate, "%v", "{{.From}}", 1)
		mailtemplate = strings.Replace(mailtemplate, "%v", "{{.Subject}}", 1)
		mailtemplate = strings.Replace(mailtemplate, "%v", "{{.Message}}", 1)

		if mailtemplate != defaultMailTemplate {
			profile.MailTemplate = &mailtemplate
		}

		if c.LinesToPreview != defaultLinesToPreview {
			profile.LinesToPreview = new(int)
			*profile.LinesToPreview = c.LinesToPreview
		}

		redirectBySubject := !c.NoRedirectChannel
		if redirectBySubject != defaultRedirectBySubject {
			profile.RedirectBySubject = &redirectBySubject
		}

		attachment := !c.NoAttachment
		if attachment != defaultAttachment {
			profile.Attachment = &attachment
		}

		if c.Disabled != defaultDisabled {
			profile.Disabled = new(bool)
			*profile.Disabled = c.Disabled
		}

		config.Profiles = append(config.Profiles, profile)
	}

	return config
}
