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

	config.Channels = []string{""}
	valid(5)

	config.Channels = []string{"Channel 1"}
	valid(6)

	config.Channels = []string{"#channel1"}
	valid(7)

	config.User = "user"
	valid(8)

	config.Password = "1234"

	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestMattermost_Fix(t *testing.T) {
	m := Mattermost{
		Channels: []string{"  test  "},
	}
	m.Fix()

	if m.Channels[0] != "#test" {
		t.Fatal("Expected #test result:", m.Channels)
	}
}
