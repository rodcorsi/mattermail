package mmail

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	idle "github.com/emersion/go-imap-idle"
	"github.com/emersion/go-imap/client"
	"github.com/rodcorsi/mattermail/model"
)

// MailProviderImap implements MailProvider using imap
type MailProviderImap struct {
	imapClient *client.Client
	idleClient *idle.Client
	cfg        *model.Email
	log        Logger
	cache      UIDCache
	idle       bool
	debug      bool
}

// MailBox default mail box
const MailBox = "INBOX"

// NewMailProviderImap creates a new MailProviderImap implementing MailProvider
func NewMailProviderImap(cfg *model.Email, log Logger, cache UIDCache, debug bool) *MailProviderImap {
	return &MailProviderImap{
		cfg:   cfg,
		cache: cache,
		log:   log,
		debug: debug,
	}
}

// CheckNewMessage gets new email from server
func (m *MailProviderImap) CheckNewMessage(handler MailHandler) error {
	m.log.Debug("MailProviderImap.CheckNewMessage")

	if err := m.checkConnection(); err != nil {
		return err
	}

	mbox, err := m.selectMailBox()
	if err != nil {
		return err
	}

	validity, uidnext := mbox.UidValidity, mbox.UidNext

	seqset := &imap.SeqSet{}
	next, err := m.cache.GetNextUID(validity)
	if err == ErrEmptyUID {
		m.log.Debug("MailProviderImap.CheckNewMessage: ErrEmptyUID search unread messages")

		criteria := &imap.SearchCriteria{
			WithoutFlags: []string{imap.SeenFlag},
		}

		uid, err := m.imapClient.UidSearch(criteria)
		if err != nil {
			m.log.Debug("MailProviderImap.CheckNewMessage: Error UIDSearch")
			return err
		}

		if len(uid) == 0 {
			m.log.Debug("MailProviderImap.CheckNewMessage: No new messages")
			return nil
		}

		m.log.Debugf("MailProviderImap.CheckNewMessage: found %v uid", len(uid))

		seqset.AddNum(uid...)

	} else if err != nil {
		return err

	} else {
		if uidnext > next {
			seqset.AddNum(next, uidnext)
		} else if uidnext < next {
			// reset cache
			m.cache.SaveNextUID(0, 0)
			return fmt.Errorf("MailProviderImap.CheckNewMessage: Cache error mailbox.next < cache.next")
		} else if uidnext == next {
			m.log.Debug("MailProviderImap.CheckNewMessage: No new messages")
			return nil
		}
	}

	messages := make(chan *imap.Message)
	done := make(chan error, 1)
	go func() {
		done <- m.imapClient.UidFetch(seqset, []string{imap.EnvelopeMsgAttr, "BODY[]"}, messages)
	}()

	for imapMsg := range messages {
		m.log.Debug("MailProviderImap.CheckNewMessage: PostMail uid:", imapMsg.Uid)

		r := imapMsg.GetBody("BODY[]")
		if r == nil {
			m.log.Debug("MailProviderImap.CheckNewMessage: message.GetBody(BODY[]) returns nil")
			continue
		}

		msg, err := mail.ReadMessage(r)
		if err != nil {
			m.log.Error("MailProviderImap.CheckNewMessage: Error on parse imap/message to mail/message")
			return err
		}

		if err := handler(msg); err != nil {
			m.log.Error("MailProviderImap.CheckNewMessage: Error handler")
			return err
		}
	}

	// Check command completion status
	if err := <-done; err != nil {
		m.log.Error("MailProviderImap.CheckNewMessage: Error on terminate fetch command")
		return err
	}

	if err := m.cache.SaveNextUID(validity, uidnext); err != nil {
		m.log.Error("MailProviderImap.CheckNewMessage: Error on save next uid")
		return err
	}

	return nil
}

// WaitNewMessage waits for a new message (idle or time.Sleep)
func (m *MailProviderImap) WaitNewMessage(timeout int) error {
	m.log.Debug("MailProviderImap.WaitNewMessage")

	// Idle mode
	if err := m.checkConnection(); err != nil {
		return err
	}

	m.log.Debug("MailProviderImap.WaitNewMessage: idle mode:", m.idle)

	if !m.idle {
		time.Sleep(time.Second * time.Duration(timeout))
		return nil
	}

	if _, err := m.selectMailBox(); err != nil {
		return err
	}

	if m.idleClient == nil {
		m.idleClient = idle.NewClient(m.imapClient)
	}

	// Create a channel to receive mailbox updates
	statuses := make(chan *imap.MailboxStatus)
	m.imapClient.MailboxUpdates = statuses

	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- m.idleClient.Idle(stop)
	}()

	reset := time.After(time.Second * time.Duration(timeout))

	closed := false
	closeChannel := func() {
		if !closed {
			close(stop)
			closed = true
		}
	}

	for {
		select {
		case status := <-statuses:
			m.log.Debug("MailProviderImap.WaitNewMessage: New mailbox status:", status)
			closeChannel()

		case err := <-done:
			if err != nil {
				m.log.Error("MailProviderImap.WaitNewMessage: Error on terminate idle", err.Error())
				return err
			}
			m.log.Debug("MailProviderImap.WaitNewMessage: Terminate idle")
			return nil
		case <-reset:
			m.log.Debug("MailProviderImap.WaitNewMessage: Timeout")
			closeChannel()
		}
	}
}

func (m *MailProviderImap) selectMailBox() (*imap.MailboxStatus, error) {

	if m.imapClient.Mailbox() != nil && m.imapClient.Mailbox().Name == MailBox {
		if err := m.imapClient.Close(); err != nil {
			m.log.Debug("MailProviderImap.selectMailBox: Error on close mailbox:", err.Error())
		}
	}

	m.log.Debug("MailProviderImap.selectMailBox: Select mailbox:", MailBox)

	mbox, err := m.imapClient.Select(MailBox, true)
	if err != nil {
		m.log.Error("MailProviderImap.selectMailBox: Error on select", MailBox)
		return nil, err
	}
	return mbox, nil
}

// checkConnection if is connected return nil or try to connect
func (m *MailProviderImap) checkConnection() error {
	if m.imapClient != nil && (m.imapClient.State() == imap.AuthenticatedState || m.imapClient.State() == imap.SelectedState) {
		m.log.Debug("MailProviderImap.CheckConnection: Connection state", m.imapClient.State)
		return nil
	}

	var err error

	//Start connection with server
	if strings.HasSuffix(m.cfg.ImapServer, ":993") {
		m.log.Debug("MailProviderImap.CheckConnection: DialTLS")
		m.imapClient, err = client.DialTLS(m.cfg.ImapServer, nil)
	} else {
		m.log.Debug("MailProviderImap.CheckConnection: Dial")
		m.imapClient, err = client.Dial(m.cfg.ImapServer)
	}

	if err != nil {
		m.log.Error("MailProviderImap.CheckConnection: Unable to connect:", m.cfg.ImapServer)
		return err
	}

	if m.debug {
		m.imapClient.SetDebug(m.log)
	}

	// Max timeout awaiting a command
	m.imapClient.Timeout = time.Minute * 3

	if *m.cfg.StartTLS {
		starttls, err := m.imapClient.SupportStartTLS()
		if err != nil {
			return err
		}

		if starttls {
			m.log.Debug("MailProviderImap.CheckConnection:StartTLS")
			var tconfig tls.Config
			if *m.cfg.TLSAcceptAllCerts {
				tconfig.InsecureSkipVerify = true
			}
			err = m.imapClient.StartTLS(&tconfig)
			if err != nil {
				return err
			}
		}
	}

	m.log.Infof("Connected with %q\n", m.cfg.ImapServer)

	err = m.imapClient.Login(m.cfg.Username, m.cfg.Password)
	if err != nil {
		m.log.Error("MailProviderImap.CheckConnection: Unable to login:", m.cfg.Username)
		return err
	}

	if _, err = m.selectMailBox(); err != nil {
		return err
	}

	idleClient := idle.NewClient(m.imapClient)
	m.idle, err = idleClient.SupportIdle()
	if err != nil {
		m.idle = false
		m.log.Error("MailProviderImap.CheckConnection: Error on check idle support")
		return err
	}

	return nil
}

// Terminate imap connection
func (m *MailProviderImap) Terminate() error {
	if m.imapClient != nil {
		m.log.Info("MailProviderImap.Terminate Logout")
		if err := m.imapClient.Logout(); err != nil {
			m.log.Error("MailProviderImap.Terminate Error:", err.Error())
			return err
		}
	}

	return nil
}
