package mmail

import (
	"fmt"
	"strings"

	mmModel "github.com/mattermost/platform/model"
	"github.com/rodcorsi/mattermail/model"
)

// MattermostProvider interface to abstract Mattermost functions
type MattermostProvider interface {
	// Login log in Mattermost
	Login() error

	// Logout terminate connection with Mattermost
	Logout() error

	// GetChannelID gets channel id by channel name return empty string if not exists
	GetChannelID(channelName string) string

	// PostMessage posts a message in Mattermost
	PostMessage(message, channelID string, attachments []*Attachment) error
}

// MattermostDefault default implementation of MattermostProvider
type MattermostDefault struct {
	cfg         *model.Mattermost
	log         Logger
	user        *mmModel.User
	client      *mmModel.Client
	channelList *mmModel.ChannelList
}

// NewMattermostDefault creates a new instance of MattermostDefault
func NewMattermostDefault(cfg *model.Mattermost, log Logger) *MattermostDefault {
	return &MattermostDefault{
		cfg: cfg,
		log: log,
	}
}

// Login log in Mattermost
func (m *MattermostDefault) Login() error {
	m.client = mmModel.NewClient(m.cfg.Server)

	m.client.GetClientProperties()
	m.log.Debug("Mattermost version:", m.client.ServerVersion)

	m.log.Debugf("Login user:%v team:%v url:%v\n", m.cfg.User, m.cfg.Team, m.cfg.Server)

	result, apperr := m.client.Login(m.cfg.User, m.cfg.Password)
	if apperr != nil {
		return apperr
	}

	m.user = result.Data.(*mmModel.User)

	// Get Team
	teams := m.client.Must(m.client.GetAllTeams()).Data.(map[string]*mmModel.Team)

	teamMatch := false
	for _, t := range teams {
		if t.Name == m.cfg.Team {
			m.client.SetTeamId(t.Id)
			teamMatch = true
			break
		}
	}

	if !teamMatch {
		return fmt.Errorf("Did not find team with name '%v'. Check if the team exist or if you are not using display name instead team name", m.cfg.Team)
	}

	//Discover channel id by channel name
	m.channelList = m.client.Must(m.client.GetChannels("")).Data.(*mmModel.ChannelList)

	return nil
}

// Logout terminate connection with Mattermost
func (m *MattermostDefault) Logout() (err error) {
	if m.client != nil {
		_, err = m.client.Logout()
	}
	return
}

// GetChannelID gets channel id by channel name return empty string if not exists
func (m *MattermostDefault) GetChannelID(channelName string) string {
	if strings.HasPrefix(channelName, "#") {
		return m.getChannelIDByName(strings.TrimPrefix(channelName, "#"))
	} else if strings.HasPrefix(channelName, "@") {
		return m.getDirectChannelIDByName(strings.TrimPrefix(channelName, "@"))
	}
	return ""
}

// PostMessage posts a message in Mattermost
func (m *MattermostDefault) PostMessage(message, channelID string, attachments []*Attachment) error {
	m.log.Debugf("Post in channel id %v", channelID)

	// Upload attachments
	var fileIds []string
	for _, a := range attachments {
		if len(a.Content) == 0 {
			continue
		}

		resp, err := m.client.UploadPostAttachment(a.Content, channelID, a.Filename)
		if resp == nil {
			return err
		}

		if len(resp.FileInfos) != 1 {
			return fmt.Errorf("error on upload file - fileinfos len different of one %v", resp.FileInfos)
		}

		fileIds = append(fileIds, resp.FileInfos[0].Id)
	}

	// Post message
	post := &mmModel.Post{ChannelId: channelID, Message: message}

	if len(fileIds) > 0 {
		post.FileIds = fileIds
	}

	res, err := m.client.CreatePost(post)
	if res == nil {
		return err
	}

	return nil
}

func (m *MattermostDefault) getChannelIDByName(channelName string) string {
	for _, c := range *m.channelList {
		if c.Name == channelName {
			return c.Id
		}
	}
	return ""
}

func (m *MattermostDefault) getDirectChannelIDByName(userName string) string {

	if m.user.Username == userName {
		m.log.Errorf("Impossible create a Direct channel, Mattermail user (%v) equals destination user (%v)\n", m.user.Username, userName)
		return ""
	}

	//result, err := client.GetProfilesForDirectMessageList(client.GetTeamId())
	result, err := m.client.SearchUsers(mmModel.UserSearch{
		AllowInactive: false,
		TeamId:        m.client.GetTeamId(),
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
	dmID := m.getChannelIDByName(dmName)

	if dmID != "" {
		return dmID
	}

	m.log.Debug("Create direct channel to user:", userName)

	result, err = m.client.CreateDirectChannel(userID)
	if err != nil {
		m.log.Error("Error on CreateDirectChannel: ", err.Error())
		return ""
	}

	directChannel := result.Data.(*mmModel.Channel)
	return directChannel.Id
}
