package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/rodcorsi/mattermail/mmail"
)

const defLinesToPreview = 10

// Version show the current version, changed during the make build
var Version = "3.0-dev"
var configFile string
var help bool
var version bool

func init() {
	flag.StringVar(&configFile, "config", "./config.json", "Sets the file location for config.json")
	flag.StringVar(&configFile, "c", "./config.json", "Sets the file location for config.json")

	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")

	flag.BoolVar(&version, "v", false, "Print version")
	flag.BoolVar(&version, "version", false, "Print version")
}

const usage = `mattermail [Options]

MatterMail is an integration service for Mattermost, MatterMail listen an email
box and publish all received emails in a channel or private group in Mattermost

Options:
    -c, --config  Sets the file location for config.json
                  Default: ./config.json
    -h, --help    Show this help
    -v, --version Print current version
`

func main() {

	flag.Parse()

	if help {
		fmt.Println(usage)
		fmt.Println(Version)
		return
	}

	if version {
		fmt.Println(Version)
		return
	}

	log.Println("Loading ", configFile)

	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Could not load: %v\n%v", configFile, err.Error())
	}

	cfgs, err := mmail.ParseConfigList(file)
	if err != nil {
		log.Fatal(err)
	}

	hasconfig := false
	log.Println("MatterMail version:", Version)

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
			mailProvider := mailProvider(c.MailConfig, logger)
			mmail.InitMatterMail(c.MatterMailConfig, logger, mailProvider)
			wg.Done()
		}()
	}

	wg.Wait()

	if !hasconfig {
		log.Println(`There is no enabled profile. Check "Disabled" field in config.json`)
	}
}

func mailProvider(cfg mmail.MailConfig, logger mmail.Logger) mmail.MailProvider {
	return mmail.NewMailProviderImap(cfg, logger)
}
