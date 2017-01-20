package mmail

import "testing"

func TestConfigIsValid(t *testing.T) {
	config := &Config{}
	valid := func() {
		if err := config.IsValid(); err == nil {
			t.Fatal("this config need to be invalid")
		}
	}

	valid()

	config.Name = "x"
	valid()

	config.Server = "z"
	valid()

	config.Server = "https://mattermost.example.com"
	valid()

	config.Team = "Team 2"
	valid()

	config.Team = "team2"
	valid()

	config.Channel = "Channel 1"
	valid()

	config.Channel = "#channel1"
	valid()

	config.MattermostUser = "user"
	valid()

	config.MattermostPass = "1234"
	valid()

	config.ImapServer = "Z"
	valid()

	config.ImapServer = "imap.ssj.com"
	valid()

	config.Email = "sdhsfk"
	valid()

	config.Email = "hahah@jdjd.xx"
	valid()

	config.EmailPass = "1234"
	valid()

	config.MailTemplate = "%v jss %v"
	valid()

	config.LinesToPreview = 10

	if err := config.IsValid(); err != nil {
		t.Fatal(err)
	}
}

func TestParseConfigList(t *testing.T) {
	if _, err := ParseConfigList([]byte("")); err == nil {
		t.Fatal("allow to load empty data")
	}

	if _, err := ParseConfigList([]byte("balalala")); err == nil {
		t.Fatal("allow to load invalid data")
	}

	config := `
[
	{
		"Name":              "Orders",
	}
]`
	if _, err := ParseConfigList([]byte(config)); err == nil {
		t.Fatal("allow to load incomplete data")
	}

	config = `
[
	{
		"Name":              "Orders",
		"Server":            "https://mattermost.example.com",
		"Team":              "team1",
		"Channel":           "#orders",
		"MattermostUser":    "mattermail@example.com",
		"MattermostPass":    "password",
		"ImapServer":        "imap.example.com:143",
		"Email":             "orders@example.com",
		"EmailPass":         "password",
		"MailTemplate":      ":incoming_envelope: _From: **%v**_\n>_%v_\n\n%v",
		"StartTLS":          false,
		"Disabled":          false,
		"Debug":             true,
		"LinesToPreview":    10,
		"NoRedirectChannel": false,
		"NoAttachment":      false,
		"Filter":            []
	}
]`
	if _, err := ParseConfigList([]byte(config)); err != nil {
		t.Fatalf("not allow to load valid data: %v", err.Error())
	}
}

func TestValidateEmail(t *testing.T) {
	assert := func(test string, expected bool) {
		if validateEmail(test) != expected {
			t.Fatalf("test %v expected %v", test, expected)
		}
	}
	assert("", false)
	assert("x", false)
	assert("as@xjsj@", false)
	assert("@", false)
	assert("@hdjdh", false)
	assert("ajjs@", false)
	assert("jjsj@kdkd", false)
	assert("jjsj@kdkd.ccj", true)
}

func TestValidateURL(t *testing.T) {
	assert := func(test string, expected bool) {
		if validateURL(test) != expected {
			t.Fatalf("test %v expected %v", test, expected)
		}
	}
	assert("", false)
	assert("x", false)
	assert("asdjfkla.com", false)
	assert("http.com", false)
	assert("http://skjks", true)
	assert("https://mattermost.example.com", true)
}

func TestValidateImap(t *testing.T) {
	assert := func(test string, expected bool) {
		if validateImap(test) != expected {
			t.Fatalf("test %v expected %v", test, expected)
		}
	}
	assert("", false)
	assert("X", false)
	assert("imap.com", true)
}

func TestValidateChannel(t *testing.T) {
	assert := func(test string, expected bool) {
		if validateChannel(test) != expected {
			t.Fatalf("test %v expected %v", test, expected)
		}
	}
	assert("", false)
	assert("SDFgfs", false)
	assert("D D", false)
	assert("chane", false)
	assert(" cha ne ", false)
	assert("#chabn", true)
	assert("@djdj", true)
}

func TestValidateTeam(t *testing.T) {
	assert := func(test string, expected bool) {
		if validateTeam(test) != expected {
			t.Fatalf("test %v expected %v", test, expected)
		}
	}
	assert("", false)
	assert("SDFgfs", false)
	assert("D D", false)
	assert("team", true)
}
