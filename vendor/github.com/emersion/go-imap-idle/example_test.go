package idle_test

import (
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap-idle"
)

func ExampleClient_Idle() {
	// Let's assume c is an IMAP client
	var c *client.Client

	// Select a mailbox
	if _, err := c.Select("INBOX", false); err != nil {
		log.Fatal(err)
	}

	idleClient := idle.NewClient(c)

	// Create a channel to receive mailbox updates
	statuses := make(chan *imap.MailboxStatus)
	c.MailboxUpdates = statuses

	// Check support for the IDLE extension
	if ok, err := idleClient.SupportIdle(); err == nil && ok {
		// Start idling
		stop := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			done <- idleClient.Idle(stop)
		}()

		// Listen for updates
		for {
			select {
			case status := <-statuses:
				log.Println("New mailbox status:", status)
				close(stop)
			case err := <-done:
				if err != nil {
					log.Fatal(err)
				}
				log.Println("Not idling anymore")
				return
			}
		}
	} else {
		// Fallback: call periodically c.Noop()
	}
}
