package model

import "fmt"

const (
	defaultStartTLS          = false
	defaultTLSAcceptAllCerts = false
)

// Email type with email settings
type Email struct {
	ImapServer        string
	Address           string
	Password          string
	StartTLS          *bool `json:",omitempty"`
	TLSAcceptAllCerts *bool `json:",omitempty"`
}

// Validate set default value for email and check if valid return err
func (c *Email) Validate() error {
	if c.ImapServer == "" {
		return fmt.Errorf("Field 'ImapServer' is empty set imap server address eg.: imap.example.com:143")
	}

	if !validateImap(c.ImapServer) {
		return fmt.Errorf("Field 'ImapServer' need to be a valid url: %v", c.ImapServer)
	}

	if !validateEmail(c.Address) {
		return fmt.Errorf("Field 'Address' need to be a valid email: %v", c.Address)
	}

	if c.Password == "" {
		return fmt.Errorf("Field 'Password' is empty")
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
}
