package model

import (
	"fmt"
)

// Mattermost type with Mattermost connection settings
type Mattermost struct {
	Server   string
	Team     string
	User     string
	Password string
}

// NewMattermost creates new Mattermost with default values
func NewMattermost() *Mattermost {
	return &Mattermost{}
}

// Validate valids Mattermost
func (c *Mattermost) Validate() error {
	if c.Server == "" {
		return fmt.Errorf("Field 'Server' is empty set mattermost address eg.: https://mattermost.example.com")
	}

	if !validateURL(c.Server) {
		return fmt.Errorf("Field 'Server' need to start with http:// or https:// and be a valid url: %v", c.Server)
	}

	if c.Team == "" {
		return fmt.Errorf("Field 'Team' is empty")
	}

	if !validateTeam(c.Team) {
		return fmt.Errorf("Field 'Team' contains invalid chars, make sure if you are using url team name: %v", c.Team)
	}

	if c.User == "" {
		return fmt.Errorf("Field 'User' is empty")
	}

	if c.Password == "" {
		return fmt.Errorf("Field 'Password' is empty")
	}

	return nil
}
