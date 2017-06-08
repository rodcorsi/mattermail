package sasl_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/emersion/go-sasl"
)

func TestNewPlainClient(t *testing.T) {
	c := sasl.NewPlainClient("identity", "username", "password")

	mech, ir, err := c.Start()
	if err != nil {
		t.Fatal("Error while starting client:", err)
	}
	if mech != "PLAIN" {
		t.Error("Invalid mechanism name:", mech)
	}

	expected := []byte{105, 100, 101, 110, 116, 105, 116, 121, 0, 117, 115, 101, 114, 110, 97, 109, 101, 0, 112, 97, 115, 115, 119, 111, 114, 100}
	if bytes.Compare(ir, expected) != 0 {
		t.Error("Invalid initial response:", ir)
	}
}

func TestNewPlainServer(t *testing.T) {
	var authenticated = false
	s := sasl.NewPlainServer(func (identity, username, password string) error {
		if username != "username" {
			return errors.New("Invalid username: " + username)
		}
		if password != "password" {
			return errors.New("Invalid password: " + password)
		}
		if identity != "identity" {
			return errors.New("Invalid identity: " + identity)
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
	if len(challenge) > 0 {
		t.Error("Invalid non-empty initial challenge:", challenge)
	}

	response := []byte{105, 100, 101, 110, 116, 105, 116, 121, 0, 117, 115, 101, 114, 110, 97, 109, 101, 0, 112, 97, 115, 115, 119, 111, 114, 100}
	challenge, done, err = s.Next(response)
	if err != nil {
		t.Fatal("Error while finishing authentication:", err)
	}
	if !done {
		t.Fatal("Authentication not finished after sending PLAIN credentials")
	}
	if len(challenge) > 0 {
		t.Error("Invalid non-empty final challenge:", challenge)
	}

	if !authenticated {
		t.Error("Not authenticated")
	}
}
