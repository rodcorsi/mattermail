package idle

import (
	"github.com/emersion/go-imap/client"
)

// Client is an IDLE client.
type Client struct {
	c *client.Client
}

// NewClient creates a new client.
func NewClient(c *client.Client) *Client {
	return &Client{c}
}

// Idle indicates to the server that the client is ready to receive unsolicited
// mailbox update messages. When the client wants to send commands again, it
// must first close stop.
func (c *Client) Idle(stop <-chan struct{}) error {
	cmd := &Command{}

	done := make(chan error, 1)
	res := &Response{
		Stop:   stop,
		Done:   done,
		Writer: c.c.Writer(),
	}

	if status, err := c.c.Execute(cmd, res); err != nil {
		return err
	} else if err := status.Err(); err != nil {
		return err
	} else {
		return <-done
	}
}

// SupportIdle checks if the server supports the IDLE extension.
func (c *Client) SupportIdle() (bool, error) {
	return c.c.Support(Capability)
}
