package mmail

import "io"

// MailHandler function called to handle mail message
type MailHandler func(mailReader io.Reader, folder string) error

// MailProvider interface to abstract email connection
type MailProvider interface {
	// CheckNewMessage gets new email from server
	CheckNewMessage(handler MailHandler, folders []string) error

	// WaitNewMessage waits for a new message (idle or time.Sleep)
	WaitNewMessage(timeout int, folders []string) error

	// Terminate mail connection
	Terminate() error
}
