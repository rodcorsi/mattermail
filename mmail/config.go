package mmail

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// MailConfig type with email settings
type MailConfig struct {
	ImapServer        string
	StartTLS          bool
	TLSAcceptAllCerts bool
	Email             string
	EmailPass         string
}

// MatterMailConfig type with Mattermail settings
type MatterMailConfig struct {
	Name              string
	Server            string
	Team              string
	Channel           string
	MattermostUser    string
	MattermostPass    string
	MailTemplate      string
	Debug             bool
	Disabled          bool
	LinesToPreview    int
	NoRedirectChannel bool
	NoAttachment      bool
	Filter            *Filter
}

// Config type to parse config.json
type Config struct {
	MailConfig
	MatterMailConfig
}

const defLinesToPreview = 10

// IsValid set default value for config and check if valid return err
func (c *Config) IsValid() error {
	if c.Name == "" {
		return fmt.Errorf("Field 'Name' in config.json is empty")
	}

	if c.Server == "" {
		return fmt.Errorf("Field 'Server' in config.json is empty")
	}

	if !validateURL(c.Server) {
		return fmt.Errorf("Field 'Server' in config.json need to start with http:// or https:// and be a valid url: %v", c.Server)
	}

	if c.Team == "" {
		return fmt.Errorf("Field 'Team' in config.json is empty")
	}

	if !validateTeam(c.Team) {
		return fmt.Errorf("Field 'Team' in config.json contains invalid chars, make sure if you are using url team name: %v", c.Team)
	}

	if c.Channel != "" && !validateChannel(c.Channel) {
		return fmt.Errorf("Field 'Channel' in config.json contains invalid chars, make sure if you are using url channel name or username. This field need to start with # for channel or @ for username: %v", c.Channel)
	}

	if c.MattermostUser == "" {
		return fmt.Errorf("Field 'MattermostUser' in config.json is empty")
	}

	if c.MattermostPass == "" {
		return fmt.Errorf("Field 'MattermostPass' in config.json is empty")
	}

	if c.ImapServer == "" {
		return fmt.Errorf("Field 'ImapServer' in config.json is empty")
	}

	if !validateImap(c.ImapServer) {
		return fmt.Errorf("Field 'ImapServer' in config.json need to be a valid url: %v", c.ImapServer)
	}

	if !validateEmail(c.Email) {
		return fmt.Errorf("Field 'Email' in config.json need to be a valid email: %v", c.Email)
	}

	if c.EmailPass == "" {
		return fmt.Errorf("Field 'EmailPass' in config.json is empty")
	}

	if c.MailTemplate == "" {
		return fmt.Errorf("Field 'MailTemplate' in config.json is empty")
	}

	if c.LinesToPreview <= 0 {
		return fmt.Errorf("Field 'LinesToPreview' in config.json need to be greater than 0")
	}

	if c.Filter != nil {
		if err := c.Filter.Valid(); err != nil {
			return fmt.Errorf("Error in Filter:%v", err)
		}
	}

	return nil
}

// ParseConfigList read data and parse a Config list
func ParseConfigList(data []byte) ([]*Config, error) {

	var cfg []*Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("Could not parse data it does not to be a valid json file: %v", err.Error())
	}

	// Set default value
	for _, c := range cfg {
		// compatibility with old versions
		c.Channel = strings.TrimSpace(c.Channel)
		c.Channel = strings.ToLower(c.Channel)

		if len(c.Channel) > 0 {
			if !strings.HasPrefix(c.Channel, "#") && !strings.HasPrefix(c.Channel, "@") {
				c.Channel = "#" + c.Channel
			}
		}

		// default value
		if c.LinesToPreview <= 0 {
			c.LinesToPreview = defLinesToPreview
		}

		if err := c.IsValid(); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9\._%+\-]+@[a-z0-9\.\-_]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

func validateURL(url string) bool {
	Re := regexp.MustCompile(`^https?://[a-z0-9\.\-_]+:?([0-9]{1,5})?$`)
	return Re.MatchString(url)
}

func validateImap(url string) bool {
	Re := regexp.MustCompile(`^[a-z0-9\.\-_]+:?([0-9]{1,5})?$`)
	return Re.MatchString(url)
}

func validateChannel(channel string) bool {
	Re := regexp.MustCompile(`^(#|@)[a-z0-9\.\-_]+$`)
	return Re.MatchString(channel)
}

func validateTeam(team string) bool {
	Re := regexp.MustCompile(`^[a-z0-9\.\-_]+$`)
	return Re.MatchString(team)
}
