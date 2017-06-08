# go-imap-idle

[![GoDoc](https://godoc.org/github.com/emersion/go-imap-idle?status.svg)](https://godoc.org/github.com/emersion/go-imap-idle)

[IDLE extension](https://tools.ietf.org/html/rfc2177) for [go-imap](https://github.com/emersion/go-imap).

## Usage

### Client

```go
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
```

Note that this is a minimal example, you'll need to:
* Stop idling and re-send an `IDLE` command [at least every 29 minutes](https://tools.ietf.org/html/rfc2177#section-3)
  to avoid being logged off
* Properly handle servers that don't support the `IDLE` extension

### Server

```go
s.Enable(idle.NewExtension())
```

## License

MIT
