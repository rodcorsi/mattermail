package model

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// ConfigV1 depracated
type ConfigV1 []struct {
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
	ImapServer        string
	StartTLS          bool
	TLSAcceptAllCerts bool
	Email             string
	EmailPass         string
}

// ParseConfigV1 read data and parse json ConfigV1
func ParseConfigV1(data []byte) (*ConfigV1, error) {
	cfg := &ConfigV1{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, errors.Wrap(err, "Could not parse data it does not to be a valid json file")
	}

	return cfg, nil
}
