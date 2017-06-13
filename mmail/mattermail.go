package mmail

import (
	"fmt"
	"net/mail"
	"time"
	"unicode/utf8"

	"github.com/rodcorsi/mattermail/model"
)

const maxMattermostAttachments = 5
const maxMattermostPostSize = 4000

// MatterMail struct with configurations, loggers and Mattemost user
type MatterMail struct {
	cfg          *model.Profile
	log          Logger
	mmProvider   MattermostProvider
	mailProvider MailProvider
}

// PostNetMail parse net/mail.Message and post in Mattermost
func (m *MatterMail) PostNetMail(msg *mail.Message) error {
	mMsg, err := ParseMailMessage(msg)
	if err != nil {
		return err
	}

	return m.PostMailMessage(mMsg)
}

// PostMailMessage MailMessage in Mattermost
func (m *MatterMail) PostMailMessage(msg *MailMessage) error {
	if err := m.mmProvider.Login(); err != nil {
		return err
	}

	defer func() {
		if err := m.mmProvider.Logout(); err != nil {
			m.log.Error("Logout error err:", err.Error())
		}
	}()

	m.log.Info("Post new message")

	mP, err := createMattermostPost(msg, m.cfg, m.log, m.mmProvider.GetChannelID)

	if err != nil {
		return err
	}

	m.log.Debugf("Post email in %v", mP.channelName)

	return m.mmProvider.PostMessage(mP.message, mP.channelID, mP.attachments)
}

// Listen starts MatterMail server
func (m *MatterMail) Listen() {
	m.log.Debug("Debug mode on")
	m.log.Info("Checking new emails")

	defer m.mailProvider.Terminate()

	for {
		if err := m.mailProvider.CheckNewMessage(m.PostNetMail); err != nil {
			m.log.Error("MatterMail.InitMatterMail Error on check new messsage:", err.Error())
			m.log.Info("Try again in 30s")
			time.Sleep(time.Second * 30)
		}

		if err := m.mailProvider.WaitNewMessage(60); err != nil {
			m.log.Error("MatterMail.InitMatterMail Error on wait new message:", err.Error())
			m.log.Info("Try again in 30s")
			time.Sleep(time.Second * 30)
		}
	}
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

type mattermostPost struct {
	channelName string
	channelID   string
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
		return nil, fmt.Errorf("Error on format Mail Template err:%v", err.Error())
	}

	// Mattermost post limit
	if utf8.RuneCountInString(mP.message) > maxMattermostPostSize {
		mP.message = string([]rune(mP.message)[:(maxMattermostPostSize-5)]) + " ..."
		postedfullmessage = false
		log.Info("Email has been cut because is larger than 4000 characters")
	}

	// Try to discovery the channel
	// redirect email by the subject
	if *cfg.RedirectChannel {
		log.Debug("Try to find channel/user by subject")
		mP.channelName = getChannelFromSubject(msg.Subject)
		mP.channelID = getChannelID(mP.channelName)
	}

	// check filters
	if mP.channelID == "" && cfg.Filter != nil {
		log.Debug("Did not find channel/user from Email Subject. Look for filter")
		mP.channelName = cfg.Filter.GetChannel(msg.From, msg.Subject)
		mP.channelID = getChannelID(mP.channelName)
	}

	// get default Channel config
	if mP.channelID == "" {
		log.Debugf("Did not find channel/user in filters. Look for channel '%v'\n", cfg.Channels)
		mP.channelName = cfg.Channels[0]
		mP.channelID = getChannelID(mP.channelName)
	}

	if mP.channelID == "" && *cfg.RedirectChannel {
		log.Debugf("Did not find channel/user with name '%v'. Trying channel town-square\n", cfg.Channels)
		mP.channelName = "town-square"
		mP.channelID = getChannelID(mP.channelName)
	}

	if mP.channelID == "" {
		return nil, fmt.Errorf("Did not find any channel to post")
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
