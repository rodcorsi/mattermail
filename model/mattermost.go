package model

import "github.com/pkg/errors"

const defaultUseAPIv3 = true

// Mattermost type with Mattermost connection settings
type Mattermost struct {
	Server   string
	Team     string
	User     string
	Password string
	UseAPIv3 *bool `json:",omitempty"`
}

// NewMattermost creates new Mattermost with default values
func NewMattermost() *Mattermost {
	mm := &Mattermost{
		UseAPIv3: new(bool),
	}
	*mm.UseAPIv3 = defaultUseAPIv3
	return mm
}

// Validate valids Mattermost
func (c *Mattermost) Validate() error {
	if c.Server == "" {
		return errors.New("Field 'Server' is empty set mattermost address eg.: https://mattermost.example.com")
	}

	if !validateURL(c.Server) {
		return errors.Errorf("Field 'Server' need to start with http:// or https:// and be a valid url: %v", c.Server)
	}

	if c.Team == "" {
		return errors.New("Field 'Team' is empty")
	}

	if !validateTeam(c.Team) {
		return errors.Errorf("Field 'Team' contains invalid chars, make sure if you are using url team name: %v", c.Team)
	}

	if c.User == "" {
		return errors.New("Field 'User' is empty")
	}

	if c.Password == "" {
		return errors.New("Field 'Password' is empty")
	}

	return nil
}

// Fix fields and using default if is necessary
func (c *Mattermost) Fix() {
	if c.UseAPIv3 == nil {
		x := defaultUseAPIv3
		c.UseAPIv3 = &x
	}
}
