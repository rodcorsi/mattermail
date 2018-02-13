package mmail

import (
	"strings"

	mmModel "github.com/mattermost/mattermost-server/model"
	"github.com/pkg/errors"
	"github.com/rodcorsi/mattermail/model"
)

// MattermostProviderV3 default implementation of MattermostProvider
type MattermostProviderV3 struct {
	cfg         *model.Mattermost
	log         Logger
	user        *mmModel.User
	client      *mmModel.Client
	channelList *mmModel.ChannelList
}

// NewMattermostProviderV3 creates a new instance of Mattermost api V3
func NewMattermostProviderV3(cfg *model.Mattermost, log Logger) *MattermostProviderV3 {
	return &MattermostProviderV3{
		cfg: cfg,
		log: log,
	}
}

// Login log in Mattermost
func (m *MattermostProviderV3) Login() error {
	m.client = mmModel.NewClient(m.cfg.Server)

	m.log.Debug("Mattermost Api V3 version:", m.client.ServerVersion)
	m.log.Debugf("Login user:%v team:%v url:%v\n", m.cfg.User, m.cfg.Team, m.cfg.Server)

	result, apperr := m.client.Login(m.cfg.User, m.cfg.Password)
	if apperr != nil {
		return errors.Wrap(apperr, "login on Mattermost V3")
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
		return errors.Errorf("Did not find team with name '%v'. Check if the team exist or if you are not using display name instead team name", m.cfg.Team)
	}

	//Discover channel id by channel name
	m.channelList = m.client.Must(m.client.GetChannels("")).Data.(*mmModel.ChannelList)

	return nil
}

// Logout terminate connection with Mattermost
func (m *MattermostProviderV3) Logout() (err error) {
	if m.client != nil {
		_, err = m.client.Logout()
	}
	return
}

// GetChannelID gets channel id by channel name return empty string if not exists
func (m *MattermostProviderV3) GetChannelID(channelName string) string {
	if strings.HasPrefix(channelName, "#") {
		return m.getChannelIDByName(strings.TrimPrefix(channelName, "#"))
	} else if strings.HasPrefix(channelName, "@") {
		return m.getDirectChannelIDByName(strings.TrimPrefix(channelName, "@"))
	}
	return ""
}

// PostMessage posts a message in Mattermost
func (m *MattermostProviderV3) PostMessage(message, channelID string, attachments []*Attachment) error {
	m.log.Debugf("Post in channel id %v", channelID)

	// Upload attachments
	var fileIds []string
	for _, a := range attachments {
		if len(a.Content) == 0 {
			continue
		}

		resp, err := m.client.UploadPostAttachment(a.Content, channelID, a.Filename)
		if resp == nil {
			return errors.Wrapf(err, "Upload Attachment on mattermost channel id:'%v' filename:'%v'", channelID, a.Filename)
		}

		if len(resp.FileInfos) != 1 {
			return errors.Errorf("error on upload file - fileinfos len different of one %v", resp.FileInfos)
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
		return errors.Wrapf(err, "create mattermost post %v", post)
	}

	return nil
}

func (m *MattermostProviderV3) getChannelIDByName(channelName string) string {
	for _, c := range *m.channelList {
		if c.Name == channelName {
			return c.Id
		}
	}
	return ""
}

func (m *MattermostProviderV3) getDirectChannelIDByName(userName string) string {

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
