package model

import (
	"github.com/pkg/errors"
)

const (
	defaultStartTLS          = false
	defaultTLSAcceptAllCerts = false
	defaultDisableIdle       = false
)

// Email type with email settings
type Email struct {
	ImapServer        string
	Username          string
	Password          string
	StartTLS          *bool `json:",omitempty"`
	TLSAcceptAllCerts *bool `json:",omitempty"`
	DisableIdle       *bool `json:",omitempty"`
}

// NewEmail creates new Email with default values
func NewEmail() *Email {
	email := &Email{
		StartTLS:          new(bool),
		TLSAcceptAllCerts: new(bool),
		DisableIdle:       new(bool),
	}
	*email.StartTLS = defaultStartTLS
	*email.TLSAcceptAllCerts = defaultTLSAcceptAllCerts
	*email.DisableIdle = defaultDisableIdle
	return email
}

// Validate set default value for email and check if valid return err
func (c *Email) Validate() error {
	if c.ImapServer == "" {
		return errors.New("Field 'ImapServer' is empty set imap server address eg.: imap.example.com:143")
	}

	if !validateImap(c.ImapServer) {
		return errors.Errorf("Field 'ImapServer' need to be a valid url: %v", c.ImapServer)
	}

	if c.Username == "" {
		return errors.New("Field 'Username' is empty, set email address or username eg.: test@example.com")
	}

	if c.Password == "" {
		return errors.New("Field 'Password' is empty")
	}

	return nil
}

// Fix fields and using default if is necessary
func (c *Email) Fix() {
	if c.StartTLS == nil {
		x := defaultStartTLS
		c.StartTLS = &x
	}
	if c.TLSAcceptAllCerts == nil {
		x := defaultTLSAcceptAllCerts
		c.TLSAcceptAllCerts = &x
	}
	if c.DisableIdle == nil {
		x := defaultDisableIdle
		c.DisableIdle = &x
	}
}
