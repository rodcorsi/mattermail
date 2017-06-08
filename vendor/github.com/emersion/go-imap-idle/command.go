package idle

import (
	"github.com/emersion/go-imap"
)

// An IDLE command.
// Se RFC 2177 section 3.
type Command struct{}

func (cmd *Command) Command() *imap.Command {
	return &imap.Command{Name: commandName}
}

func (cmd *Command) Parse(fields []interface{}) error {
	return nil
}
