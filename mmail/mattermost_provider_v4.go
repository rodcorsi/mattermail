package mmail

import (
	"fmt"
	"strings"

	mmModel "github.com/mattermost/platform/model"
	"github.com/rodcorsi/mattermail/model"
)

// MattermostProviderV4 default implementation of MattermostProvider
type MattermostProviderV4 struct {
	cfg         *model.Mattermost
	log         Logger
	user        *mmModel.User
	client      *mmModel.Client4
	team        *mmModel.Team
	channelList []*mmModel.Channel
}

// NewMattermostProviderV4 creates a new instance of Mattermost api V4
func NewMattermostProviderV4(cfg *model.Mattermost, log Logger) *MattermostProviderV4 {
	return &MattermostProviderV4{
		cfg: cfg,
		log: log,
	}
}

// Login log in Mattermost
func (m *MattermostProviderV4) Login() error {
	var resp *mmModel.Response
	m.client = mmModel.NewAPIv4Client(m.cfg.Server)

	m.log.Debugf("Login user:%v team:%v url:%v\n", m.cfg.User, m.cfg.Team, m.cfg.Server)

	m.user, resp = m.client.Login(m.cfg.User, m.cfg.Password)
	m.log.Debug("Mattermost Api V4 version:", resp.ServerVersion)
	if resp.Error != nil {
		return resp.Error
	}

	// Get Team
	m.team, resp = m.client.GetTeamByName(m.cfg.Team, "")
	if resp.Error != nil {
		return fmt.Errorf("Did not find team with name '%v'. Check if the team exist or if you are not using display name instead team name error:%v", m.cfg.Team, resp.Error)
	}

	//Discover channel id by channel name
	m.channelList, resp = m.client.GetChannelsForTeamForUser(m.team.Id, m.user.Id, "")
	if resp.Error != nil {
		return fmt.Errorf("Error on get channel list error:%v", resp.Error)
	}

	return nil
}

// Logout terminate connection with Mattermost
func (m *MattermostProviderV4) Logout() (err error) {
	if m.client != nil {
		_, resp := m.client.Logout()
		err = resp.Error
	}
	return
}

// GetChannelID gets channel id by channel name return empty string if not exists
func (m *MattermostProviderV4) GetChannelID(channelName string) string {
	if strings.HasPrefix(channelName, "#") {
		return m.getChannelIDByName(strings.TrimPrefix(channelName, "#"))
	} else if strings.HasPrefix(channelName, "@") {
		return m.getDirectChannelIDByName(strings.TrimPrefix(channelName, "@"))
	}
	return ""
}

// PostMessage posts a message in Mattermost
func (m *MattermostProviderV4) PostMessage(message, channelID string, attachments []*Attachment) error {
	m.log.Debugf("Post in channel id %v", channelID)

	// Upload attachments
	var fileIds []string
	for _, a := range attachments {
		if len(a.Content) == 0 {
			continue
		}

		fileResp, resp := m.client.UploadFile(a.Content, channelID, a.Filename)
		if resp.Error != nil {
			return resp.Error
		}

		if len(fileResp.FileInfos) != 1 {
			return fmt.Errorf("error on upload file - fileinfos len different of one %v", fileResp.FileInfos)
		}

		fileIds = append(fileIds, fileResp.FileInfos[0].Id)
	}

	// Post message
	post := &mmModel.Post{ChannelId: channelID, Message: message}

	if len(fileIds) > 0 {
		post.FileIds = fileIds
	}

	post, resp := m.client.CreatePost(post)
	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

func (m *MattermostProviderV4) getChannelIDByName(channelName string) string {
	for _, c := range m.channelList {
		if c.Name == channelName {
			return c.Id
		}
	}
	return ""
}

func (m *MattermostProviderV4) getDirectChannelIDByName(userName string) string {

	if m.user.Username == userName {
		m.log.Errorf("Impossible create a Direct channel, Mattermail user (%v) equals destination user (%v)\n", m.user.Username, userName)
		return ""
	}

	users, resp := m.client.SearchUsers(&mmModel.UserSearch{
		AllowInactive: false,
		TeamId:        m.team.Id,
		Term:          userName,
	})

	if resp.Error != nil {
		m.log.Error("Error on SearchUsers: ", resp.Error)
		return ""
	}

	var userID string
	for _, u := range users {
		if u.Username == userName {
			userID = u.Id
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

	directChannel, resp := m.client.CreateDirectChannel(m.user.Id, userID)
	if resp.Error != nil {
		m.log.Error("Error on CreateDirectChannel: ", resp.Error)
		return ""
	}

	return directChannel.Id
}
