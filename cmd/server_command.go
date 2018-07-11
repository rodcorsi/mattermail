package cmd

import (
	"flag"
	"fmt"

	"github.com/rodcorsi/mattermail/mmail"
	"github.com/rodcorsi/mattermail/model"
)

type serverCommand struct {
	configFile string
}

func (sc *serverCommand) execute() error {
	config, err := model.NewConfigFromFile(sc.configFile)
	if err != nil {
		return fmt.Errorf("Error on read '%v' file, make sure if this file is has a valid configuration.\nExecute 'mattermail migrate -c %v' to migrate this file to new version if it is necessary, learn more at https://github.com/rodcorsi/mattermail/#migrate-configuration.\n\nerr:%v", sc.configFile, sc.configFile, err.Error())
	}

	fmt.Printf("Mattermail Server Version: %v\n", Version)
	return mmail.Start(config)
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
