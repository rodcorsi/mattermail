package mmail

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/jhillyerd/go.enmime"
	"github.com/mattermost/platform/model"
)

const maxMattermostAttachments = 5

// MatterMail struct with configurations, loggers and Mattemost user
type MatterMail struct {
	cfg  MatterMailConfig
	user *model.User
	log  Logger
}

// postMessage Create a post in Mattermost
func (m *MatterMail) postMessage(client *model.Client, channelID string, message string, fileIds []string) error {
	post := &model.Post{ChannelId: channelID, Message: message}

	if len(fileIds) > 0 {
		post.FileIds = fileIds
	}

	res, err := client.CreatePost(post)
	if res == nil {
		return err
	}

	return nil
}

// PostFile Post files and message in Mattermost server
func (m *MatterMail) PostFile(from, subject, message, emailname string, emailbody string, attach []enmime.MIMEPart) error {

	client := model.NewClient(m.cfg.Server)

	m.log.Debug(client)

	m.log.Debugf("Login user:%v team:%v url:%v\n", m.cfg.MattermostUser, m.cfg.Team, m.cfg.Server)

	result, apperr := client.Login(m.cfg.MattermostUser, m.cfg.MattermostPass)
	if apperr != nil {
		return apperr
	}

	m.user = result.Data.(*model.User)

	m.log.Info("Post new message")

	defer client.Logout()

	// Get Team
	teams := client.Must(client.GetAllTeams()).Data.(map[string]*model.Team)

	teamMatch := false
	for _, t := range teams {
		if t.Name == m.cfg.Team {
			client.SetTeamId(t.Id)
			teamMatch = true
			break
		}
	}

	if !teamMatch {
		return fmt.Errorf("Did not find team with name '%v'. Check if the team exist or if you are not using display name instead team name", m.cfg.Team)
	}

	//Discover channel id by channel name
	var channelID, channelName string
	channelList := client.Must(client.GetChannels("")).Data.(*model.ChannelList)

	// redirect email by the subject
	if !m.cfg.NoRedirectChannel {
		m.log.Debug("Try to find channel/user by subject")
		channelName = getChannelFromSubject(subject)
		channelID = m.getChannelID(client, channelList, channelName)
	}

	// check filters
	if channelID == "" && m.cfg.Filter != nil {
		m.log.Debug("Did not find channel/user from Email Subject. Look for filter")
		channelName = m.cfg.Filter.GetChannel(from, subject)
		channelID = m.getChannelID(client, channelList, channelName)
	}

	// get default Channel config
	if channelID == "" {
		m.log.Debugf("Did not find channel/user in filters. Look for channel '%v'\n", m.cfg.Channel)
		channelName = m.cfg.Channel
		channelID = m.getChannelID(client, channelList, channelName)
	}

	if channelID == "" && !m.cfg.NoRedirectChannel {
		m.log.Debugf("Did not find channel/user with name '%v'. Trying channel town-square\n", m.cfg.Channel)
		channelName = "town-square"
		channelID = m.getChannelID(client, channelList, channelName)
	}

	if channelID == "" {
		return fmt.Errorf("Did not find any channel to post")
	}

	m.log.Debugf("Post email in %v", channelName)

	if m.cfg.NoAttachment || (len(attach) == 0 && len(emailname) == 0) {
		return m.postMessage(client, channelID, message, nil)
	}

	var fileIds []string
	filesUploaded := 0

	uploadFile := func(filename string, data []byte) error {
		if filesUploaded >= maxMattermostAttachments {
			m.log.Infof("File '%v' was not uploaded due the Mattermost limit of %v files", filename, maxMattermostAttachments)
			return nil
		}

		if len(data) == 0 {
			return nil
		}

		resp, err := client.UploadPostAttachment(data, channelID, filename)
		if resp == nil {
			return err
		}

		if len(resp.FileInfos) != 1 {
			return fmt.Errorf("error on upload file - fileinfos len different of one %v", resp.FileInfos)
		}

		filesUploaded++
		fileIds = append(fileIds, resp.FileInfos[0].Id)
		return nil
	}

	if len(emailname) > 0 {
		if err := uploadFile(emailname, []byte(emailbody)); err != nil {
			return err
		}
	}

	for _, a := range attach {
		if err := uploadFile(a.FileName(), a.Content()); err != nil {
			return err
		}
	}

	return m.postMessage(client, channelID, message, fileIds)
}

func (m *MatterMail) getChannelID(client *model.Client, channelList *model.ChannelList, channelName string) string {
	if strings.HasPrefix(channelName, "#") {
		return getChannelIDByName(channelList, strings.TrimPrefix(channelName, "#"))
	} else if strings.HasPrefix(channelName, "@") {
		return m.getDirectChannelIDByName(client, channelList, strings.TrimPrefix(channelName, "@"))
	}
	return ""
}

func getChannelIDByName(channelList *model.ChannelList, channelName string) string {
	for _, c := range *channelList {
		if c.Name == channelName {
			return c.Id
		}
	}
	return ""
}

func (m *MatterMail) getDirectChannelIDByName(client *model.Client, channelList *model.ChannelList, userName string) string {

	if m.user.Username == userName {
		m.log.Errorf("Impossible create a Direct channel, Mattermail user (%v) equals destination user (%v)\n", m.user.Username, userName)
		return ""
	}

	//result, err := client.GetProfilesForDirectMessageList(client.GetTeamId())
	result, err := client.SearchUsers(model.UserSearch{
		AllowInactive: false,
		TeamId:        client.GetTeamId(),
		Term:          userName,
	})

	if err != nil {
		m.log.Error("Error on SearchUsers: ", err.Error())
		return ""
	}

	profiles := result.Data.([]*model.User)
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

	dmName := model.GetDMNameFromIds(m.user.Id, userID)
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

	directChannel := result.Data.(*model.Channel)
	return directChannel.Id
}

// ParseMailMessage parse mail message to post in Mattermost
func (m *MatterMail) ParseMailMessage(msg *mail.Message) error {
	mime, _ := enmime.ParseMIMEBody(msg) // Parse message body with enmime

	// read only some lines of text
	partmessage := readLines(mime.Text, m.cfg.LinesToPreview)

	postedfullmessage := false

	if partmessage != mime.Text && len(partmessage) > 0 {
		partmessage += " ..."
	} else if partmessage == mime.Text {
		postedfullmessage = true
	}

	var emailname, emailbody string
	if len(mime.HTML) > 0 {
		emailname = "email.html"
		emailbody = mime.HTML
		for _, p := range mime.Inlines {
			emailbody = replaceCID(&emailbody, &p)
		}

		for _, p := range mime.OtherParts {
			emailbody = replaceCID(&emailbody, &p)
		}

	} else if len(mime.Text) > 0 && !postedfullmessage {
		emailname = "email.txt"
		emailbody = mime.Text
	}

	subject := mime.GetHeader("Subject")
	from := NonASCII(msg.Header.Get("From"))
	message := fmt.Sprintf(m.cfg.MailTemplate, from, subject, partmessage)

	// Mattermost post limit
	if utf8.RuneCountInString(message) > 4000 {
		message = string([]rune(message)[:3995]) + " ..."
		m.log.Info("Email has been cut because is larger than 4000 characters")
	}

	return m.PostFile(from, subject, message, emailname, emailbody, mime.Attachments)
}

// InitMatterMail init MatterMail server
func InitMatterMail(cfg MatterMailConfig, log Logger, mailprovider MailProvider) {
	m := &MatterMail{
		cfg: cfg,
		log: log,
	}

	m.log.Debug("Debug mode on")
	m.log.Info("Checking new emails")

	mailprovider.AddListenerOnReceived(m.ParseMailMessage)

	defer mailprovider.Terminate()
	mailprovider.Start()
}
