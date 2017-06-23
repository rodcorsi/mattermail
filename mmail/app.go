package mmail

import (
	"errors"
	"fmt"
	"sync"

	"github.com/rodcorsi/mattermail/model"
)

// Start server
func Start(config *model.Config) error {
	var wg sync.WaitGroup

	if err := config.Validate(); err != nil {
		return fmt.Errorf("Config is invalid err:%v", err.Error())
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
	cache := NewUIDCacheFile(directory, profile.Email.Address, MailBox)
	mailProvider := NewMailProviderImap(profile.Email, logger, cache, debug)
	mattermost := NewMattermostDefault(profile.Mattermost, logger)
	return NewMatterMail(profile, logger, mailProvider, mattermost)
}
