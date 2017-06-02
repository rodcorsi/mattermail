package mmail

import (
	"bytes"
	"crypto/tls"
	"net/mail"
	"strings"
	"time"

	"github.com/mxk/go-imap/imap"
)

// MailProviderImap implements MailProvider using imap
type MailProviderImap struct {
	imapClient *imap.Client
	cfg        MailConfig
	log        Logger
	listener   MailListener
}

// NewMailProviderImap creates a new MailProviderImap implementing MailProvider
func NewMailProviderImap(cfg MailConfig, log Logger) *MailProviderImap {
	return &MailProviderImap{
		cfg: cfg,
		log: log,
	}
}

// checkConnection if is connected return nil or try to connect
func (m *MailProviderImap) checkConnection() error {
	if m.imapClient != nil && (m.imapClient.State() == imap.Auth || m.imapClient.State() == imap.Selected) {
		m.log.Debug("CheckConnection: Connection alive")
		return nil
	}

	var err error

	//Start connection with server
	if strings.HasSuffix(m.cfg.ImapServer, ":993") {
		m.log.Debug("CheckConnection: DialTLS")
		m.imapClient, err = imap.DialTLS(m.cfg.ImapServer, nil)
	} else {
		m.log.Debug("CheckConnection: Dial")
		m.imapClient, err = imap.Dial(m.cfg.ImapServer)
	}

	if err != nil {
		m.log.Error("Unable to connect:", err)
		return err
	}

	if m.cfg.StartTLS && m.imapClient.Caps["STARTTLS"] {
		m.log.Debug("CheckConnection:StartTLS")
		var tconfig tls.Config
		if m.cfg.TLSAcceptAllCerts {
			tconfig.InsecureSkipVerify = true
		}
		_, err = m.imapClient.StartTLS(&tconfig)
		if err != nil {
			return err
		}
	}

	m.log.Infof("Connected with %q\n", m.cfg.ImapServer)

	_, err = m.imapClient.Login(m.cfg.Email, m.cfg.EmailPass)
	if err != nil {
		m.log.Error("Unable to login:", m.cfg.Email)
		return err
	}

	return nil
}

// checkNewMails Check if exist a new mail
func (m *MailProviderImap) checkNewMails() error {
	m.log.Debug("checkNewMails")

	if err := m.checkConnection(); err != nil {
		return err
	}

	var (
		cmd *imap.Command
		rsp *imap.Response
	)

	// Open a mailbox (synchronous command - no need for imap.Wait)
	m.imapClient.Select("INBOX", false)

	var specs []imap.Field
	specs = append(specs, "UNSEEN")
	seq := &imap.SeqSet{}

	// get headers and UID for UnSeen message in src inbox...
	cmd, err := imap.Wait(m.imapClient.UIDSearch(specs...))
	if err != nil {
		m.log.Debug("Error UIDSearch UTF-8:")
		m.log.Debug(err)
		m.log.Debug("Try with US-ASCII")

		// try again with US-ASCII
		cmd, err = imap.Wait(m.imapClient.Send("UID SEARCH", append([]imap.Field{"CHARSET", "US-ASCII"}, specs...)...))
		if err != nil {
			m.log.Error("UID SEARCH US-ASCII")
			return err
		}
	}

	for _, rsp := range cmd.Data {
		for _, uid := range rsp.SearchResults() {
			m.log.Debug("checkNewMails:AddNum ", uid)
			seq.AddNum(uid)
		}
	}

	// no new messages
	if seq.Empty() {
		m.log.Debug("checkNewMails: No new messages")
		return nil
	}

	cmd, _ = m.imapClient.UIDFetch(seq, "BODY[]")
	postmail := false

	for cmd.InProgress() {
		m.log.Debug("checkNewMails: cmd in Progress")
		// Wait for the next response (no timeout)
		m.imapClient.Recv(-1)

		// Process command data
		for _, rsp = range cmd.Data {
			msgFields := rsp.MessageInfo().Attrs
			header := imap.AsBytes(msgFields["BODY[]"])
			if msg, _ := mail.ReadMessage(bytes.NewReader(header)); msg != nil {
				m.log.Debug("checkNewMails:PostMail")
				// Call listener
				if err := m.listener(msg); err != nil {
					return err
				}
				postmail = true
			}
		}
		cmd.Data = nil
	}

	// Check command completion status
	if rsp, err := cmd.Result(imap.OK); err != nil {
		if err == imap.ErrAborted {
			m.log.Error("Fetch command aborted")
			return err
		}
		m.log.Error("Fetch error:", rsp.Info)
		return err
	}

	cmd.Data = nil

	if postmail {
		m.log.Debug("checkNewMails: Mark all messages with flag \\Seen")

		//Mark all messages seen
		_, err = imap.Wait(m.imapClient.UIDStore(seq, "+FLAGS.SILENT", `\Seen`))
		if err != nil {
			m.log.Error("Error UIDStore \\Seen")
			return err
		}
	}

	return nil
}

// idleMailBox Change to state idle in imap server
func (m *MailProviderImap) idleMailBox() error {
	m.log.Debug("idleMailBox")

	if err := m.checkConnection(); err != nil {
		return err
	}

	// Open a mailbox (synchronous command - no need for imap.Wait)
	m.imapClient.Select("INBOX", false)

	_, err := m.imapClient.Idle()
	if err != nil {
		return err
	}

	defer m.imapClient.IdleTerm()
	timeout := 0
	for {
		err := m.imapClient.Recv(time.Second)
		timeout++
		if err == nil || timeout > 180 {
			break
		}
	}
	return nil
}

// Terminate imap connection
func (m *MailProviderImap) Terminate() {
	if m.imapClient != nil {
		m.imapClient.Logout(time.Second * 5)
	}
}

func (m *MailProviderImap) tryTime(message string, fn func() error) {
	if err := fn(); err != nil {
		m.log.Info(message, err, "\n", "Try again in 30s")
		time.Sleep(30 * time.Second)
		fn()
	}
}

// AddListenerOnReceived adds a listener for when an email arrives
func (m *MailProviderImap) AddListenerOnReceived(listener MailListener) {
	m.listener = listener
}

// Start mail monitor to dispach when an email arrives
func (m *MailProviderImap) Start() {
	m.tryTime("Error on check new email:", m.checkNewMails)

	for {
		m.tryTime("Error Idle:", m.idleMailBox)
		m.tryTime("Error on check new email:", m.checkNewMails)
	}
}
