package model

import "testing"

func Test_validateURL(t *testing.T) {
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

func Test_validateImap(t *testing.T) {
	assert := func(test string, expected bool) {
		if validateImap(test) != expected {
			t.Fatalf("test %v expected %v", test, expected)
		}
	}
	assert("", false)
	assert("X", false)
	assert("imap.com", true)
}

func Test_validateChannel(t *testing.T) {
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

func Test_validateTeam(t *testing.T) {
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
