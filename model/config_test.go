package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	config := &Config{}
	valid := func(n int) {
		if err := config.Validate(); err == nil {
			t.Fatal("Test:", n, "this config need to be invalid")
		}
	}

	valid(0)

	config = NewConfig()
	config.Directory = os.TempDir()
	valid(1)

	config.Profiles = []*Profile{{}}
	valid(2)

	config.Profiles = []*Profile{NewProfile()}
	valid(3)
}

func TestConfig_Fix(t *testing.T) {
	c := &Config{}
	c.Profiles = append(c.Profiles, NewProfile())
	c.Fix()

	if *c.Debug != defaultDebug {
		t.Fatal("Expected Debug:", defaultDebug, " result:", *c.Debug)
	}
}

func TestMigrateFromV1(t *testing.T) {
	configV1 := `
[
	{
		"Name":              "Orders",
		"Server":            "https://mattermost.example.com",
		"Team":              "team1",
		"Channel":           "#orders",
		"MattermostUser":    "mattermail@example.com",
		"MattermostPass":    "matterpassword",
		"ImapServer":        "imap.example.com:143",
		"Email":             "orders@example.com",
		"EmailPass":         "emailpassword",
		"MailTemplate":      ":incoming_envelope: _From: **%v**_\n>_%v_\n\n%v",
		"StartTLS":          true,
		"Disabled":          false,
		"Debug":             true,
		"LinesToPreview":    10,
		"NoRedirectChannel": true,
		"NoAttachment":      false,
		"Filter":            [
			{"Subject":"Feature", "Channel":"#feature"},
			{"From":"test@gmail.com", "Subject":"To Me", "Channel":"@test2"}
		]
	}
]`

	v1, err := ParseConfigV1([]byte(configV1))
	if err != nil {
		t.Fatal("Error on parse Config V1", err.Error())
	}

	result := MigrateFromV1(*v1)
	redirect := false
	startTLS := true

	expected := &Config{
		Directory: defaultDirectory,
		Profiles: []*Profile{{
			Name:              "Orders",
			Channels:          []string{"#orders"},
			RedirectBySubject: &redirect,
			Email: &Email{
				ImapServer: "imap.example.com:143",
				Username:   "orders@example.com",
				Password:   "emailpassword",
				StartTLS:   &startTLS,
			},
			Mattermost: &Mattermost{
				Server:   "https://mattermost.example.com",
				Team:     "team1",
				User:     "mattermail@example.com",
				Password: "matterpassword",
			},
			Filter: &Filter{
				&Rule{
					Subject:  "Feature",
					Channels: []string{"#feature"},
				},
				&Rule{
					From:     "test@gmail.com",
					Subject:  "To Me",
					Channels: []string{"@test2"},
				},
			},
		}},
	}

	exp, _ := json.MarshalIndent(expected, "", "\t")
	res, _ := json.MarshalIndent(result, "", "\t")
	if !reflect.DeepEqual(exp, res) {
		t.Fatalf("Different Config expected:\n%v\n result:\n%v", string(exp), string(res))
	}
}

func TestNewConfigFromFile(t *testing.T) {
	if _, err := NewConfigFromFile(":invalid:"); err == nil {
		t.Fatal("Expected error to load invalid file path")
	}

	rootDir := filepath.Join(findDir("model"), "../")
	if _, err := NewConfigFromFile(filepath.Join(rootDir, "README.md")); err == nil {
		t.Fatal("Expected error to load invalid json file")
	}

	if _, err := NewConfigFromFile(filepath.Join(rootDir, "config.json")); err != nil {
		t.Fatal("Error on load valid config.json", err.Error())
	}
}
