package mmail

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/cseeger-epages/mattermail/model"
)

// Start server
func Start(config *model.Config) error {
	var wg sync.WaitGroup

	if err := config.Validate(); err != nil {
		return errors.Wrap(err, "Config is invalid")
	}

	hasconfig := false

	for _, profile := range config.Profiles {
		if *profile.Disabled {
			continue
		}
		hasconfig = true

		wg.Add(1)
		debug := *config.Debug
		p := profile
		go func() {
			createMatterMail(p, config.Directory, debug).Listen()
			wg.Done()
		}()
	}

	if !hasconfig {
		return errors.New(`There is no enabled profile. Check "Disabled" field in config.json`)
	}

	wg.Wait()

	return nil
}

func createMatterMail(profile *model.Profile, directory string, debug bool) *MatterMail {
	logger := NewLog(profile.Name, debug)
	cache := NewUIDCacheFile(directory, profile.Email.Username, MailBox)
	mailProvider := NewMailProviderImap(profile.Email, logger, cache, debug)
	mattermost := NewMattermostProvider(profile.Mattermost, logger)
	return NewMatterMail(profile, logger, mailProvider, mattermost)
}
