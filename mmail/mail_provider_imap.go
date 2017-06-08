package mmail

import (
	"crypto/tls"
	"net/mail"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	idle "github.com/emersion/go-imap-idle"
	"github.com/emersion/go-imap/client"
)

// MailProviderImap implements MailProvider using imap
type MailProviderImap struct {
	imapClient *client.Client
	cfg        MailConfig
	log        Logger
	idle       bool
	debug      bool
}

const mailBox = "INBOX"

// NewMailProviderImap creates a new MailProviderImap implementing MailProvider
func NewMailProviderImap(cfg MailConfig, log Logger, debug bool) *MailProviderImap {
	return &MailProviderImap{
		cfg:   cfg,
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

	if err := m.selectMailBox(); err != nil {
		return err
	}

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

	seqset := &imap.SeqSet{}
	seqset.AddNum(uid...)

	messages := make(chan *imap.Message, len(uid))
	done := make(chan error, 1)
	go func() {
		done <- m.imapClient.UidFetch(seqset, []string{imap.EnvelopeMsgAttr, "BODY[]"}, messages)
	}()

	emailPosted := make(map[uint32]bool)
	for _, v := range uid {
		emailPosted[v] = false
	}

	for imapMsg := range messages {
		m.log.Debug("MailProviderImap.CheckNewMessage: PostMail uid:", imapMsg.Uid)
		if emailPosted[imapMsg.Uid] {
			m.log.Debug("MailProviderImap.CheckNewMessage: Email was posted uid:", imapMsg.Uid)
			continue
		}

		r := imapMsg.GetBody("BODY[]")
		if r == nil {
			m.log.Error("MailProviderImap.CheckNewMessage: message.GetBody(BODY[]) returns nil")
			continue
		}

		msg, err := mail.ReadMessage(r)
		if err != nil {
			m.log.Error("MailProviderImap.CheckNewMessage: Error on parse imap/message to mail/message")
			continue
		}

		if err := handler(msg); err != nil {
			m.log.Debug("MailProviderImap.CheckNewMessage: Error handler:", err.Error())
			continue
		} else {
			emailPosted[imapMsg.Uid] = true
		}
	}

	// Check command completion status
	if err := <-done; err != nil {
		m.log.Error("MailProviderImap.CheckNewMessage: Error on terminate fetch command")
		return err
	}

	errorset := &imap.SeqSet{}
	for k, posted := range emailPosted {
		if !posted {
			errorset.AddNum(k)
		}
	}

	if errorset.Empty() {
		return nil
	}

	// Mark all valid messages as read
	err = m.imapClient.UidStore(errorset, imap.RemoveFlags, []interface{}{imap.SeenFlag}, nil)
	if err != nil {
		m.log.Error("MailProviderImap.CheckNewMessage: Error UIDStore UNSEEN")
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

	if err := m.selectMailBox(); err != nil {
		return err
	}

	idleClient := idle.NewClient(m.imapClient)

	// Create a channel to receive mailbox updates
	statuses := make(chan *imap.MailboxStatus)
	m.imapClient.MailboxUpdates = statuses

	stop := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- idleClient.Idle(stop)
	}()

	reset := time.After(time.Second * time.Duration(timeout))

	var lock sync.Mutex
	closed := false
	closeChannel := func() {
		lock.Lock()
		if !closed {
			close(stop)
			closed = true
		}
		lock.Unlock()
	}

	for {
		select {
		case status := <-statuses:
			m.log.Debug("MailProviderImap.WaitNewMessage: New mailbox status:", status.Format())
			closeChannel()

		case err := <-done:
			if err != nil {
				m.log.Error("MailProviderImap.WaitNewMessage: Error on terminate idle", err.Error())
				return err
			}
			return nil
		case <-reset:
			closeChannel()
		}
	}
}

func (m *MailProviderImap) selectMailBox() error {
	if m.imapClient.Mailbox != nil && m.imapClient.Mailbox.Name == mailBox {
		return nil
	}

	_, err := m.imapClient.Select(mailBox, false)
	if err != nil {
		m.log.Error("MailProviderImap.selectMailBox: Error on select", mailBox)
		return err
	}
	return nil
}

// checkConnection if is connected return nil or try to connect
func (m *MailProviderImap) checkConnection() error {
	if m.imapClient != nil && m.imapClient.State != imap.LogoutState {
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

	if m.cfg.StartTLS {
		starttls, err := m.imapClient.SupportStartTLS()
		if err != nil {
			return err
		}

		if starttls {
			m.log.Debug("MailProviderImap.CheckConnection:StartTLS")
			var tconfig tls.Config
			if m.cfg.TLSAcceptAllCerts {
				tconfig.InsecureSkipVerify = true
			}
			err = m.imapClient.StartTLS(&tconfig)
			if err != nil {
				return err
			}
		}
	}

	m.log.Infof("Connected with %q\n", m.cfg.ImapServer)

	err = m.imapClient.Login(m.cfg.Email, m.cfg.EmailPass)
	if err != nil {
		m.log.Error("MailProviderImap.CheckConnection: Unable to login:", m.cfg.Email)
		return err
	}

	if err = m.selectMailBox(); err != nil {
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
