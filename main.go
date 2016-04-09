package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
)

type config struct {
	Name           string
	Server         string
	Team           string
	Channel        string
	MattermostUser string
	MattermostPass string
	ImapServer     string
	StartTLS       bool
	Email          string
	EmailPass      string
	MailTemplate   string
	Debug          bool
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
		c := cfg
		go func() {
			InitMatterMail(&c)
			wg.Done()
		}()
	}

	wg.Wait()
}
