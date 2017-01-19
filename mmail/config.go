package mmail

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

// Config type to parse config.json
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
	NoAttachment      bool
	Filter            *Filter
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
		return fmt.Errorf("Field 'Server' in config.json need to start with http:// or https:// and be a valid url")
	}

	if c.Team == "" {
		return fmt.Errorf("Field 'Team' in config.json is empty")
	}

	if !validateTeam(c.Team) {
		return fmt.Errorf("Field 'Team' in config.json contains invalid chars, make sure if you are using url team name")
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
		return fmt.Errorf("Field 'ImapServer' in config.json need to be a valid url")
	}

	if !validateEmail(c.Email) {
		return fmt.Errorf("Field 'Email' in config.json is empty")
	}

	if c.EmailPass == "" {
		return fmt.Errorf("Field 'EmailPass' in config.json is empty")
	}

	if c.MailTemplate == "" {
		return fmt.Errorf("Field 'MailTemplate' in config.json is empty")
	}

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

// LoadConfigArray load json file and returns an array of config
func LoadConfigArray(configFile string) ([]*Config, error) {
	log.Println("Loading ", configFile)

	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Could not load: %v", err.Error())
	}

	var cfg []*Config
	err = json.Unmarshal(file, &cfg)

	if err != nil {
		return nil, fmt.Errorf("Could not parse: %v", err.Error())
	}

	// Set default value
	for _, c := range cfg {
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
	Re := regexp.MustCompile(`^(#|@)?[a-z0-9\.\-_]+$`)
	return Re.MatchString(channel)
}

func validateTeam(team string) bool {
	Re := regexp.MustCompile(`^[a-z0-9\.\-_]+$`)
	return Re.MatchString(team)
}
