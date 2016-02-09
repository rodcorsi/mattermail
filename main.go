package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
)

type config struct {
	Name, Server, Team, ChannelId, MattermostUser, MattermostPass, ImapServer, Email, EmailPass, MailTemplate string
}

func loadconfig() []config {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		log.Fatal("Could not load config.json", err)
	}

	var cfg []config
	err = json.Unmarshal(file, &cfg)

	if err != nil {
		log.Fatal("Could not parse config.json", err)
	}

	return cfg
}

func main() {

	cfgs := loadconfig()

	var wg sync.WaitGroup
	for _, cfg := range cfgs {
		wg.Add(1)
		go func() {
			InitMatterMail(&cfg)
			wg.Done()
		}()
	}

	wg.Wait()
}
