package mmail

import "github.com/rodcorsi/mattermail/model"

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

// NewMattermostProvider creates a new instance of Mattermost
func NewMattermostProvider(cfg *model.Mattermost, log Logger) MattermostProvider {
	if *cfg.UseAPIv3 {
		log.Info("Using Mattermost Api Version 3")
		return NewMattermostProviderV3(cfg, log)
	}
	log.Info("Using Mattermost Api Version 4")
	return NewMattermostProviderV4(cfg, log)
}
