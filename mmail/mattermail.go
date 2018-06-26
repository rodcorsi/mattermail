package mmail

import (
	"io"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/rodcorsi/mattermail/model"
)

const (
	maxMattermostAttachments = 5
	maxMattermostPostSize    = 4000
	tryAgainTime             = 30
	waitMessageTimeout       = 60
)

// MatterMail struct with configurations, loggers and Mattemost user
type MatterMail struct {
	cfg          *model.Profile
	log          Logger
	mmProvider   MattermostProvider
	mailProvider MailProvider
}

// PostNetMail read net/mail.Message and post in Mattermost
func (m *MatterMail) PostNetMail(mailReader io.Reader) error {
	mMsg, err := ReadMailMessage(mailReader)
	if err != nil {
		return errors.Wrap(err, "parse mail message")
	}

	return m.PostMailMessage(mMsg)
}

// PostMailMessage MailMessage in Mattermost
func (m *MatterMail) PostMailMessage(msg *MailMessage) error {
	if err := m.mmProvider.Login(); err != nil {
		return errors.Wrap(err, "login on Mattermost to post mail message")
	}

	defer func() {
		if err := m.mmProvider.Logout(); err != nil {
			m.log.Error("Logout error err:", err)
		}
	}()

	m.log.Info("Post new message")

	mP, err := createMattermostPost(msg, m.cfg, m.log, m.mmProvider.GetChannelID)

	if err != nil {
		return errors.Wrap(err, "create mattermost post")
	}

	for name, id := range mP.channelMap {
		m.log.Debugf("Post email in %v", name)
		if err := m.mmProvider.PostMessage(mP.message, id, mP.attachments); err != nil {
			return errors.Wrap(err, "post message on mattermost")
		}
	}

	return nil
}

// Listen starts MatterMail server
func (m *MatterMail) Listen() {
	m.log.Debug("Debug mode on")
	m.log.Info("Checking new emails")

	defer m.mailProvider.Terminate()

	for {
		if err := m.checkAndWait(); err != nil {
			m.log.Debug(err.Error())
			m.log.Info("Terminate Mail Provider after error")
			m.mailProvider.Terminate()
			m.log.Infof("Try again in %vs", tryAgainTime)
			time.Sleep(time.Second * tryAgainTime)
		} else {
			time.Sleep(time.Second * 2)
		}
	}
}

func (m *MatterMail) checkAndWait() error {
	if err := m.mailProvider.CheckNewMessage(m.PostNetMail); err != nil {
		m.log.Error("MatterMail.InitMatterMail Error on check new messsage:", err.Error())
		m.mailProvider.Terminate()
		return errors.Wrap(err, "check new message")
	}

	time.Sleep(time.Second * 2)

	if err := m.mailProvider.WaitNewMessage(waitMessageTimeout); err != nil {
		m.log.Error("MatterMail.InitMatterMail Error on wait new message:", err.Error())
		m.mailProvider.Terminate()
		return errors.Wrap(err, "wait new message")
	}
	return nil
}

// NewMatterMail creates a new MatterMail instance
func NewMatterMail(cfg *model.Profile, log Logger, mailProvider MailProvider, mmProvider MattermostProvider) *MatterMail {
	return &MatterMail{
		cfg:          cfg,
		log:          log,
		mailProvider: mailProvider,
		mmProvider:   mmProvider,
	}
}

// map[channel name] = channel id
type channelMap map[string]string

type mattermostPost struct {
	channelMap  channelMap
	message     string
	attachments []*Attachment
}

func createMattermostPost(msg *MailMessage, cfg *model.Profile, log Logger, getChannelID func(string) string) (*mattermostPost, error) {
	mP := &mattermostPost{}

	// read only some lines of text
	partmessage := readLines(msg.EmailText, *cfg.LinesToPreview)

	postedfullmessage := false

	if partmessage != msg.EmailText && len(partmessage) > 0 {
		partmessage += " ..."
	} else if partmessage == msg.EmailText {
		postedfullmessage = true
	}

	// Apply MailTemplate to format message
	var err error
	mP.message, err = cfg.FormatMailTemplate(msg.From, msg.Subject, partmessage)
	if err != nil {
		return nil, errors.Wrap(err, "format Mail Template")
	}

	// Mattermost post limit
	if utf8.RuneCountInString(mP.message) > maxMattermostPostSize {
		mP.message = string([]rune(mP.message)[:(maxMattermostPostSize-5)]) + " ..."
		postedfullmessage = false
		log.Info("Email has been cut because is larger than 4000 characters")
	}

	mP.channelMap = chooseChannel(cfg, msg, log, getChannelID)

	if mP.channelMap == nil {
		return nil, errors.New("Did not find any channel to post")
	}

	// Attachments
	if !*cfg.Attachment {
		return mP, nil
	}

	// Post original email
	if msg.EmailType == EmailTypeHTML {
		mP.attachments = append(mP.attachments, &Attachment{
			Filename: "email.html",
			Content:  []byte(msg.EmailBody),
		})
	} else if !postedfullmessage {
		mP.attachments = append(mP.attachments, &Attachment{
			Filename: "email.txt",
			Content:  []byte(msg.EmailBody),
		})
	}

	// Attachments
	for _, a := range msg.Attachments {
		if len(mP.attachments) >= maxMattermostAttachments {
			log.Debugf("Max number of attachments '%v'\n", maxMattermostAttachments)
			break
		}
		mP.attachments = append(mP.attachments, a)
	}

	return mP, nil
}

func validateChannelNames(channelNames []string, getChannelID func(string) string) channelMap {
	channels := make(channelMap)
	gotOne := false
	for _, v := range channelNames {
		if id := getChannelID(v); id != "" {
			gotOne = true
			channels[v] = id
		}
	}

	if !gotOne {
		return nil
	}

	return channels
}

func chooseChannel(cfg *model.Profile, msg *MailMessage, log Logger, getChannelID func(string) string) channelMap {
	var chMap channelMap

	// Try to discovery the channel
	// redirect email by the subject
	if *cfg.RedirectBySubject {
		log.Debug("Try to find channel/user by subject")
		if chMap = validateChannelNames(getChannelsFromSubject(msg.Subject), getChannelID); chMap != nil {
			return chMap
		}
	}

	// check filters
	if cfg.Filter != nil {
		log.Debug("Did not find channel/user from Email Subject. Look for filter")
		if chMap = validateChannelNames(cfg.Filter.GetChannels(msg.From, msg.Subject), getChannelID); chMap != nil {
			return chMap
		}
	}

	// get default Channel config
	log.Debugf("Did not find channel/user in filters. Look for channel '%v'\n", cfg.Channels)
	if chMap = validateChannelNames(cfg.Channels, getChannelID); chMap != nil {
		return chMap
	}

	return nil
}
