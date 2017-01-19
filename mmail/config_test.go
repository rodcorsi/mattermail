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

	config.Channel = "channel1"
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

	if err := config.IsValid(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfigArray(t *testing.T) {
	if _, err := LoadConfigArray(""); err == nil {
		t.Fatal("allow to load empty filename")
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
	assert("chane", true)
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
