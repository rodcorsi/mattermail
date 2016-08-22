package mmail

import (
	"fmt"
	"strings"
)

type Config struct {
	Name              string
	Server            string
	Team              string
	Channel           string
	MattermostUser    string
	MattermostPass    string
	ImapServer        string
	StartTLS          bool
	TLSAcceptAllCerts bool
	Email             string
	EmailPass         string
	MailTemplate      string
	Debug             bool
	Disabled          bool
	LinesToPreview    int
	NoRedirectChannel bool
	Filter            *Filter
	NoAttachments	  bool
}

const defLinesToPreview = 10

// Valid set default value for config and check if valid return err
func (c *Config) Valid() error {
	if c.LinesToPreview <= 0 {
		c.LinesToPreview = defLinesToPreview
	}

	c.Channel = strings.TrimSpace(c.Channel)
	c.Channel = strings.ToLower(c.Channel)

	if len(c.Channel) > 0 {
		if !strings.HasPrefix(c.Channel, "#") && !strings.HasPrefix(c.Channel, "@") {
			c.Channel = "#" + c.Channel
		}
	}

	if c.Filter != nil {
		if err := c.Filter.Valid(); err != nil {
			return fmt.Errorf("Error in Filter:%v", err)
		}
	}

	return nil
}
