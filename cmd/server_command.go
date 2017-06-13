package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/rodcorsi/mattermail/mmail"
	"github.com/rodcorsi/mattermail/model"
)

type serverCommand struct {
	configFile string
}

func (sc *serverCommand) execute() error {
	data, err := ioutil.ReadFile(sc.configFile)
	if err != nil {
		return fmt.Errorf("Could not load: %v\n%v", sc.configFile, err.Error())
	}

	cfgs := &model.Config{}
	if err = json.Unmarshal(data, cfgs); err != nil {
		return fmt.Errorf("Error on read '%v' file, make sure if this file is has a valid configuration.\nExecute 'mattermail migrate -c %v' to migrate this file to new version if is necessary.\nerr:%v", sc.configFile, sc.configFile, err.Error())
	}

	cfgs.Fix()

	if err = cfgs.Validate(); err != nil {
		return fmt.Errorf("File '%v' is invalid err:%v", sc.configFile, err.Error())
	}

	hasconfig := false

	var wg sync.WaitGroup
	for _, profile := range cfgs.Profiles {
		if *profile.Disabled {
			continue
		}
		hasconfig = true

		wg.Add(1)
		c := profile
		debug := *cfgs.Debug

		go func() {
			logger := mmail.NewLog(c.Name, debug)
			mailProvider := mailProvider(c.Email, logger, debug)
			mmProvider := mmail.NewMattermostDefault(c.Mattermost, logger)
			mm := mmail.NewMatterMail(c, logger, mailProvider, mmProvider)
			mm.Listen()
			wg.Done()
		}()
	}

	wg.Wait()

	if !hasconfig {
		return errors.New(`There is no enabled profile. Check "Disabled" field in config.json`)
	}

	return nil
}

func (sc *serverCommand) parse(arguments []string) error {
	flags := flag.NewFlagSet("server", flag.ExitOnError)
	flags.Usage = serverUsage

	flags.StringVar(&sc.configFile, "config", "./config.json", "Sets the file location for config.json")
	flags.StringVar(&sc.configFile, "c", "./config.json", "Sets the file location for config.json")

	return flags.Parse(arguments)
}

func serverUsage() {
	fmt.Printf(`Start Mattermail server using configuration file

Usage:
	mattermail server [options]

Options:
    -c, --config  Sets the file location for config.json
                  Default: ./config.json
    -h, --help    Show this help
`)
}

func mailProvider(cfg *model.Email, logger mmail.Logger, debug bool) mmail.MailProvider {
	return mmail.NewMailProviderImap(cfg, logger, debug)
}
