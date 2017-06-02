package mmail

import "net/mail"

// MailListener function called when an email arrives
type MailListener func(msg *mail.Message) error

// MailProvider interface to abstract email connection
type MailProvider interface {
	Start()
	Terminate()

	// AddListenerOnReceived adds a listener for when an email arrives
	AddListenerOnReceived(listener MailListener)
}
