package mmail

import (
	"crypto/tls"
	"net/mail"
	"strings"
	"time"

	"github.com/rodcorsi/mattermail/model"
	"github.com/emersion/go-imap"
	idle "github.com/emersion/go-imap-idle"
	"github.com/emersion/go-imap/client"
	"github.com/pkg/errors"
)

// MailProviderImap implements MailProvider using imap
type MailProviderImap struct {
	imapClient *client.Client
	idleClient *idle.Client
	cfg        *model.Email
	log        Logger
	caches     []UIDCache
	idle       bool
	debug      bool
}

// MailBox default mail box
const MailBox = "INBOX"

// NewMailProviderImap creates a new MailProviderImap implementing MailProvider
func NewMailProviderImap(cfg *model.Email, log Logger, caches []UIDCache, debug bool) *MailProviderImap {
	return &MailProviderImap{
		cfg:    cfg,
		caches: caches,
		log:    log,
		debug:  debug,
	}
}

// CheckNewMessage gets new email from server
func (m *MailProviderImap) CheckNewMessage(handler MailHandler, folders []string) error {

	m.log.Debug("MailProviderImap.CheckNewMessage")

	if err := m.checkConnection(); err != nil {
		return errors.Wrap(err, "checkConnection with imap server")
	}

	// add INBOX to our list
	folders = append(folders, MailBox)

	for _, folder := range folders {

		mbox, err := m.selectMailBox(folder)
		m.log.Debug("MailProviderImap.CheckNewMessage: select MailBox: ", folder)

		if err != nil {
			return errors.Wrap(err, "select mailbox")
		}

		validity, uidnext := mbox.UidValidity, mbox.UidNext

		seqset := &imap.SeqSet{}

		cacheid := 0

		// get first matching cache file
		for id, cache := range m.caches {
			if cache.GetMailBox() == folder {
				cacheid = id
			}
		}
		next, err := m.caches[cacheid].GetNextUID(validity)

		if err == ErrEmptyUID {
			m.log.Debug("MailProviderImap.CheckNewMessage: ErrEmptyUID search unread messages")

			criteria := &imap.SearchCriteria{
				WithoutFlags: []string{imap.SeenFlag},
			}

			uid, err := m.imapClient.UidSearch(criteria)
			if err != nil {
				m.log.Debug("MailProviderImap.CheckNewMessage: Error UIDSearch")
				return errors.Wrapf(err, "imap UIDSearch %v", criteria)
			}

			if len(uid) == 0 {
				m.log.Debug("MailProviderImap.CheckNewMessage: No new messages for MailBox ", folder)
				return nil
			}

			m.log.Debugf("MailProviderImap.CheckNewMessage: found %v uid in Mailbox %s", len(uid), folder)

			seqset.AddNum(uid...)

		} else if err != nil {
			return errors.Wrap(err, "GetNextUID")
		} else {
			if uidnext > next {
				seqset.AddNum(next, uidnext)
			} else if uidnext < next {
				// reset cache
				m.caches[cacheid].SaveNextUID(0, 0)
				return errors.New("Cache error mailbox.next < cache.next")
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
				return errors.Wrap(err, "parse imap/message to mail/message")
			}

			if err := handler(msg, folder); err != nil {
				m.log.Error("MailProviderImap.CheckNewMessage: Error handler")
				return errors.Wrap(err, "execute MailHandler")
			}
		}

		// Check command completion status
		if err := <-done; err != nil {
			m.log.Error("MailProviderImap.CheckNewMessage: Error on terminate fetch command")
			return errors.Wrap(err, "terminate fetch command")
		}

		if err := m.caches[cacheid].SaveNextUID(validity, uidnext); err != nil {
			m.log.Error("MailProviderImap.CheckNewMessage: Error on save next uid")
			return errors.Wrap(err, "save next uid")
		}
	}
	return nil
}

// WaitNewMessage waits for a new message (idle or time.Sleep)
func (m *MailProviderImap) WaitNewMessage(timeout int, folders []string) error {
	m.log.Debug("MailProviderImap.WaitNewMessage")

	// Idle mode
	if err := m.checkConnection(); err != nil {
		return errors.Wrap(err, "WaitNewMessage checkConnection")
	}

	m.log.Debug("MailProviderImap.WaitNewMessage: idle mode:", m.idle)

	if !m.idle {
		time.Sleep(time.Second * time.Duration(timeout))
		return nil
	}

	// add INBOX
	folders = append(folders, MailBox)

	for _, folder := range folders {
		if _, err := m.selectMailBox(folder); err != nil {
			return errors.Wrap(err, "select mailbox")
		}
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
				return errors.Wrap(err, "terminate idle")
			}
			m.log.Debug("MailProviderImap.WaitNewMessage: Terminate idle")
			return nil
		case <-reset:
			m.log.Debug("MailProviderImap.WaitNewMessage: Timeout")
			closeChannel()
		}
	}
}

func (m *MailProviderImap) selectMailBox(mailbox string) (*imap.MailboxStatus, error) {

	if m.imapClient.Mailbox() != nil && m.imapClient.Mailbox().Name == mailbox {
		if err := m.imapClient.Close(); err != nil {
			m.log.Debug("MailProviderImap.selectMailBox: Error on close mailbox:", err.Error())
		}
	}

	m.log.Debug("MailProviderImap.selectMailBox: Select mailbox:", mailbox)

	mbox, err := m.imapClient.Select(mailbox, true)
	if err != nil {
		m.log.Error("MailProviderImap.selectMailBox: Error on select", mailbox)
		return nil, errors.Wrapf(err, "select mailbox '%v'", mailbox)
	}
	return mbox, nil
}

// checkConnection if is connected return nil or try to connect
func (m *MailProviderImap) checkConnection() error {
	var err error
	if m.imapClient != nil {
		// ConnectingState 0
		// NotAuthenticatedState 1
		// AuthenticatedState 2
		// SelectedState 6
		// LogoutState 8
		// ConnectedState 7
		cliState := m.imapClient.State()
		if cliState == imap.AuthenticatedState || cliState == imap.SelectedState {
			m.log.Debug("MailProviderImap.CheckConnection: Client state", cliState)
			m.log.Debug("MailProviderImap.CheckConnection: Connection state", m.imapClient.State())
			m.log.Debug("MailProviderImap.CheckConnection: IsTLS", m.imapClient.IsTLS())
			if err = m.imapClient.Check(); err != nil {
				m.log.Debug("MailProviderImap.CheckConnection: Check", err)

				// on error try to recconnect to resolv the problem
				if err = m.Terminate(); err != nil {
					m.log.Debug("MailProviderImap.CheckConnection: Terminate", err)
				}

				if err = m.Connect(); err != nil {
					m.log.Error("MailProviderImap.CheckConnection: Unable to login:", m.cfg.Username)
				}
			}
			return nil
		}
	}

	if err = m.Connect(); err != nil {
		m.log.Error("MailProviderImap.CheckConnection: Unable to login:", m.cfg.Username)
	}

	if _, err = m.selectMailBox(MailBox); err != nil {
		return errors.Wrap(err, "select mailbox on checkConnection")
	}

	return nil
}

func (m *MailProviderImap) Connect() error {

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
		return errors.Wrapf(err, "unable to connect '%v'", m.cfg.ImapServer)
	}

	if m.debug {
		m.imapClient.SetDebug(m.log)
	}

	// Max timeout awaiting a command
	m.imapClient.Timeout = time.Minute * 3

	if *m.cfg.StartTLS {
		starttls, err := m.imapClient.SupportStartTLS()
		if err != nil {
			return errors.Wrap(err, "check support STARTTLS")
		}

		if starttls {
			m.log.Debug("MailProviderImap.CheckConnection:StartTLS")
			var tconfig tls.Config
			if *m.cfg.TLSAcceptAllCerts {
				tconfig.InsecureSkipVerify = true
			}
			err = m.imapClient.StartTLS(&tconfig)
			if err != nil {
				return errors.Wrap(err, "enable StartTLS")
			}
		}
	}

	m.log.Infof("Connected with %q\n", m.cfg.ImapServer)

	err = m.imapClient.Login(m.cfg.Username, m.cfg.Password)
	if err != nil {
		m.log.Error("MailProviderImap.CheckConnection: Unable to login:", m.cfg.Username)
		return errors.Wrapf(err, "unable to login username:'%v'", m.cfg.Username)
	}

	idleClient := idle.NewClient(m.imapClient)
	m.idle, err = idleClient.SupportIdle()

	if err != nil {
		m.idle = false
		m.log.Error("MailProviderImap.CheckConnection: Error on check idle support")
		return errors.Wrap(err, "on check idle support")
	}
	return nil
}

// Terminate imap connection
func (m *MailProviderImap) Terminate() error {
	if m.imapClient != nil {
		m.log.Info("MailProviderImap.Terminate")
		if err := m.imapClient.Logout(); err != nil {
			m.log.Debug("MailProviderImap.Terminate: imap.Logout", err.Error())
		}
		if err := m.imapClient.Close(); err != nil {
			m.log.Debug("MailProviderImap.Terminate: imap.Close", err.Error())
		}
		if err := m.imapClient.Terminate(); err != nil {
			m.log.Debug("MailProviderImap.Terminate: imap.Terminate", err.Error())
			return err
		}
		// clean up clients
		m.imapClient = nil
		m.idleClient = nil
	}

	return nil
}
