package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/cseeger-epages/mattermail/model"
)

type migrateCommand struct {
	configFile string
}

func (mc *migrateCommand) execute() error {
	data, err := ioutil.ReadFile(mc.configFile)
	if err != nil {
		return fmt.Errorf("Could not load: %v\n%v", mc.configFile, err.Error())
	}

	v1, err := model.ParseConfigV1(data)
	if err != nil {
		return err
	}

	config := model.MigrateFromV1(*v1)

	configData, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return fmt.Errorf("Could not marshal config: %v\n%v", mc.configFile, err.Error())
	}

	fmt.Println(string(configData))
	return nil
}

func (mc *migrateCommand) parse(arguments []string) error {
	flags := flag.NewFlagSet("migrate", flag.ExitOnError)
	flags.Usage = migrateUsage

	flags.StringVar(&mc.configFile, "config", "./config.json", "Sets the file location for config.json")
	flags.StringVar(&mc.configFile, "c", "./config.json", "Sets the file location for config.json")
	return flags.Parse(arguments)
}

func migrateUsage() {
	fmt.Printf(`Migrate Mattermail config.json to new version

Usage:
	mattermail migrate [options]

Options:
    -c, --config  Sets the file location for config.json
                  Default: ./config.json
    -h, --help    Show this help
`)
}
