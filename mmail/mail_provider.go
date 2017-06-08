package mmail

import "net/mail"

// MailHandler function called to handle mail message
type MailHandler func(msg *mail.Message) error

// MailProvider interface to abstract email connection
type MailProvider interface {
	// CheckNewMessage gets new email from server
	CheckNewMessage(handler MailHandler) error

	// WaitNewMessage waits for a new message (idle or time.Sleep)
	WaitNewMessage(timeout int) error

	// Terminate mail connection
	Terminate() error
}
