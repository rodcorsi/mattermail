package model

import (
	"fmt"
	"strings"
)

// Mattermost type with Mattermost connection settings
type Mattermost struct {
	Server   string
	Team     string
	Channels []string
	User     string
	Password string
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

	if len(c.Channels) == 0 {
		return fmt.Errorf("Field 'Channels' need to set at least one channel or user for destination")
	}

	for _, channel := range c.Channels {
		if channel != "" && !validateChannel(channel) {
			return fmt.Errorf("Field 'Channels' contains invalid chars, make sure if you are using url channel name or username. This field need to start with # for channel or @ for username: %v", channel)
		}
	}

	if c.User == "" {
		return fmt.Errorf("Field 'User' is empty")
	}

	if c.Password == "" {
		return fmt.Errorf("Field 'Password' is empty")
	}

	return nil
}

// Fix fields and using default if is necessary
func (c *Mattermost) Fix() {
	for i, channel := range c.Channels {
		channel = strings.TrimSpace(channel)
		channel = strings.ToLower(channel)

		if !strings.HasPrefix(channel, "#") && !strings.HasPrefix(channel, "@") {
			channel = "#" + channel
		}
		c.Channels[i] = channel
	}
}
