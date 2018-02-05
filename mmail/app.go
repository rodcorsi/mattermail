package mmail

import (
	"sync"

	"github.com/cseeger-epages/mattermail/model"
	"github.com/pkg/errors"
)

// Start server
func Start(config *model.Config) error {
	var wg sync.WaitGroup

	if err := config.Validate(); err != nil {
		return errors.Wrap(err, "Config is invalid")
	}

	hasconfig := false
	var folders []string
	folders = append(folders, MailBox)

	for _, profile := range config.Profiles {
		if *profile.Disabled {
			continue
		}
		hasconfig = true

		if profile.Filter != nil {
			folders = append(folders, profile.Filter.ListFolder()...)
			folders = dedupStrings(folders)
		}

		wg.Add(1)
		debug := *config.Debug
		p := profile
		go func() {
			createMatterMail(p, config.Directory, folders, debug).Listen()
			wg.Done()
		}()
	}

	if !hasconfig {
		return errors.New(`There is no enabled profile. Check "Disabled" field in config.json`)
	}

	wg.Wait()

	return nil
}

func createMatterMail(profile *model.Profile, directory string, folders []string, debug bool) *MatterMail {
	logger := NewLog(profile.Name, debug)

	// mailbox caches for each mailbox defined in filter rules
	var caches []UIDCache
	caches = append(caches, NewUIDCacheFile(directory, profile.Email.Username, MailBox))
	for _, folder := range folders {
		caches = append(caches, NewUIDCacheFile(directory, profile.Email.Username, folder))
	}

	mailProvider := NewMailProviderImap(profile.Email, logger, caches, debug)
	mattermost := NewMattermostProvider(profile.Mattermost, logger)
	return NewMatterMail(profile, logger, mailProvider, mattermost)
}
