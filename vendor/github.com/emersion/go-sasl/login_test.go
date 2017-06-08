package sasl_test

import (
	"errors"
	"testing"

	"github.com/emersion/go-sasl"
)

func TestNewLoginServer(t *testing.T) {
	var authenticated = false
	s := sasl.NewLoginServer(func (username, password string) error {
		if username != "tim" {
			return errors.New("Invalid username: " + username)
		}
		if password != "tanstaaftanstaaf" {
			return errors.New("Invalid password: " + password)
		}

		authenticated = true
		return nil
	})

	challenge, done, err := s.Next(nil)
	if err != nil {
		t.Fatal("Error while starting server:", err)
	}
	if done {
		t.Fatal("Done after starting server")
	}
	if string(challenge) != "Username:" {
		t.Error("Invalid first challenge:", challenge)
	}

	challenge, done, err = s.Next([]byte("tim"))
	if err != nil {
		t.Fatal("Error while sending username:", err)
	}
	if done {
		t.Fatal("Done after sending username")
	}
	if string(challenge) != "Password:" {
		t.Error("Invalid challenge after sending username:", challenge)
	}

	challenge, done, err = s.Next([]byte("tanstaaftanstaaf"))
	if err != nil {
		t.Fatal("Error while sending password:", err)
	}
	if !done {
		t.Fatal("Authentication not finished after sending password")
	}
	if len(challenge) > 0 {
		t.Error("Invalid non-empty final challenge:", challenge)
	}

	if !authenticated {
		t.Error("Not authenticated")
	}
}
