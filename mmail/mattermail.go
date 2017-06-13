package mmail

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	mmModel "github.com/mattermost/platform/model"
	"github.com/rodcorsi/mattermail/model"
)

const maxMattermostAttachments = 5
const maxMattermostPostSize = 4000

// MatterMail struct with configurations, loggers and Mattemost user
type MatterMail struct {
	cfg  *model.Profile
	user *mmModel.User
	log  Logger
}

func getChannelIDByName(channelList *mmModel.ChannelList, channelName string) string {
	for _, c := range *channelList {
		if c.Name == channelName {
			return c.Id
		}
	}
	return ""
}

func (m *MatterMail) getDirectChannelIDByName(client *mmModel.Client, channelList *mmModel.ChannelList, userName string) string {

	if m.user.Username == userName {
		m.log.Errorf("Impossible create a Direct channel, Mattermail user (%v) equals destination user (%v)\n", m.user.Username, userName)
		return ""
	}

	//result, err := client.GetProfilesForDirectMessageList(client.GetTeamId())
	result, err := client.SearchUsers(mmModel.UserSearch{
		AllowInactive: false,
		TeamId:        client.GetTeamId(),
		Term:          userName,
	})

	if err != nil {
		m.log.Error("Error on SearchUsers: ", err.Error())
		return ""
	}

	profiles := result.Data.([]*mmModel.User)
	var userID string

	for _, p := range profiles {
		if p.Username == userName {
			userID = p.Id
			break
		}
	}

	if userID == "" {
		m.log.Debug("Did not find the username:", userName)
		return ""
	}

	dmName := mmModel.GetDMNameFromIds(m.user.Id, userID)
	dmID := getChannelIDByName(channelList, dmName)

	if dmID != "" {
		return dmID
	}

	m.log.Debug("Create direct channel to user:", userName)

	result, err = client.CreateDirectChannel(userID)
	if err != nil {
		m.log.Error("Error on CreateDirectChannel: ", err.Error())
		return ""
	}

	directChannel := result.Data.(*mmModel.Channel)
	return directChannel.Id
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

	client := mmModel.NewClient(m.cfg.Mattermost.Server)

	m.log.Debugf("Login user:%v team:%v url:%v\n", m.cfg.Mattermost.User, m.cfg.Mattermost.Team, m.cfg.Mattermost.Server)

	result, apperr := client.Login(m.cfg.Mattermost.User, m.cfg.Mattermost.Password)
	if apperr != nil {
		return apperr
	}

	m.user = result.Data.(*mmModel.User)

	m.log.Info("Post new message")

	defer client.Logout()

	// Get Team
	teams := client.Must(client.GetAllTeams()).Data.(map[string]*mmModel.Team)

	teamMatch := false
	for _, t := range teams {
		if t.Name == m.cfg.Mattermost.Team {
			client.SetTeamId(t.Id)
			teamMatch = true
			break
		}
	}

	if !teamMatch {
		return fmt.Errorf("Did not find team with name '%v'. Check if the team exist or if you are not using display name instead team name", m.cfg.Mattermost.Team)
	}

	//Discover channel id by channel name
	channelList := client.Must(client.GetChannels("")).Data.(*mmModel.ChannelList)

	mP, err := createMattermostPost(msg, m.cfg, m.log, func(channelName string) string {
		if strings.HasPrefix(channelName, "#") {
			return getChannelIDByName(channelList, strings.TrimPrefix(channelName, "#"))
		} else if strings.HasPrefix(channelName, "@") {
			return m.getDirectChannelIDByName(client, channelList, strings.TrimPrefix(channelName, "@"))
		}
		return ""
	})

	if err != nil {
		return err
	}

	m.log.Debugf("Post email in %v", mP.channelName)

	// Upload attachments
	var fileIds []string
	for _, a := range mP.attachments {
		if len(a.Content) == 0 {
			continue
		}

		resp, err := client.UploadPostAttachment(a.Content, mP.channelID, a.Filename)
		if resp == nil {
			return err
		}

		if len(resp.FileInfos) != 1 {
			return fmt.Errorf("error on upload file - fileinfos len different of one %v", resp.FileInfos)
		}

		fileIds = append(fileIds, resp.FileInfos[0].Id)
	}

	// Post message
	post := &mmModel.Post{ChannelId: mP.channelID, Message: mP.message}

	if len(fileIds) > 0 {
		post.FileIds = fileIds
	}

	res, err := client.CreatePost(post)
	if res == nil {
		return err
	}

	return nil
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
		log.Debugf("Did not find channel/user in filters. Look for channel '%v'\n", cfg.Mattermost.Channels)
		mP.channelName = cfg.Mattermost.Channels[0]
		mP.channelID = getChannelID(mP.channelName)
	}

	if mP.channelID == "" && *cfg.RedirectChannel {
		log.Debugf("Did not find channel/user with name '%v'. Trying channel town-square\n", cfg.Mattermost.Channels)
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

// InitMatterMail init MatterMail server
func InitMatterMail(cfg *model.Profile, log Logger, mailprovider MailProvider) {
	m := &MatterMail{
		cfg: cfg,
		log: log,
	}

	m.log.Debug("Debug mode on")
	m.log.Info("Checking new emails")

	defer mailprovider.Terminate()

	for {
		if err := mailprovider.CheckNewMessage(m.PostNetMail); err != nil {
			m.log.Error("MatterMail.InitMatterMail Error on check new messsage:", err.Error())
			m.log.Info("Try again in 30s")
			time.Sleep(time.Second * 30)
		}

		if err := mailprovider.WaitNewMessage(60); err != nil {
			m.log.Error("MatterMail.InitMatterMail Error on wait new message:", err.Error())
			m.log.Info("Try again in 30s")
			time.Sleep(time.Second * 30)
		}
	}
}
