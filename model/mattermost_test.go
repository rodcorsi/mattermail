package model

import (
	"testing"
)

func TestMattermost_Validate(t *testing.T) {
	config := &Mattermost{}
	valid := func(n int) {
		if err := config.Validate(); err == nil {
			t.Fatal("Test:", n, "this config need to be invalid")
		}
	}

	valid(0)

	config.Server = "z"
	valid(1)

	config.Server = "https://mattermost.example.com"
	valid(2)

	config.Team = "Team 2"
	valid(3)

	config.Team = "team2"
	valid(4)

	config.User = "user"
	valid(5)

	config.Password = "1234"

	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}
}
