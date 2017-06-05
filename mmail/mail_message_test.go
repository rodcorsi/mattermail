package mmail

import (
	"bytes"
	"net/mail"
	"testing"
)

func TestParseMailMessage(t *testing.T) {
	email := `From: John Doe <jdoe@machine.example>
To: Mary Smith <mary@example.net>
Subject: Saying Hello
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.machine.example>

This is a message just to say hello.
`
	msg, err := mail.ReadMessage(bytes.NewBuffer([]byte(email)))
	if err != nil {
		t.Fatal("Failed parsing email:", err)
	}

	mm, err := ParseMailMessage(msg)
	if err != nil {
		t.Fatal("Failed to parsing msg:", err)
	}

	if mm.From != "John Doe <jdoe@machine.example>" {
		t.Error("Error on field From:", mm.From)
	}

	if mm.Subject != "Saying Hello" {
		t.Error("Error on field Subject:", mm.Subject)
	}

	if mm.EmailText != "This is a message just to say hello.\n" {
		t.Error("Error on field EmailText:", mm.EmailText)
	}

	if mm.EmailType != EmailTypeText {
		t.Error("Expected EmailTypeText result:", mm.EmailType)
	}
}
