package cmd

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/rodcorsi/mattermail/mmail"
)

type serverCommand struct {
	configFile string
}

func (sc *serverCommand) execute() error {
	file, err := ioutil.ReadFile(sc.configFile)
	if err != nil {
		return fmt.Errorf("Could not load: %v\n%v", sc.configFile, err.Error())
	}

	cfgs, err := mmail.ParseConfigList(file)
	if err != nil {
		return err
	}

	hasconfig := false

	var wg sync.WaitGroup
	for _, cfg := range cfgs {
		if cfg.Disabled {
			continue
		}
		hasconfig = true

		wg.Add(1)
		c := cfg
		go func() {
			logger := mmail.NewLog(c.Name, c.Debug)
			mailProvider := mailProvider(c.MailConfig, logger, c.Debug)
			mmail.InitMatterMail(c.MatterMailConfig, logger, mailProvider)
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

func mailProvider(cfg mmail.MailConfig, logger mmail.Logger, debug bool) mmail.MailProvider {
	return mmail.NewMailProviderImap(cfg, logger, debug)
}
