package model

import (
	"testing"
)

func TestEmail_Validate(t *testing.T) {
	config := &Email{}
	valid := func(n int) {
		if err := config.Validate(); err == nil {
			t.Fatal("Test:", n, "this config need to be invalid")
		}
	}

	valid(0)

	config.ImapServer = "Z"
	valid(1)

	config.ImapServer = "imap.ssj.com"
	valid(2)

	config.Username = "foo@blah.hh"
	valid(4)

	config.Password = "1234"

	if err := config.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestEmail_Fix(t *testing.T) {
	e := Email{}
	e.Fix()

	if *e.StartTLS != defaultStartTLS {
		t.Fatal("Expected StartTLS:", defaultStartTLS, " result:", *e.StartTLS)
	}

	if *e.TLSAcceptAllCerts != defaultTLSAcceptAllCerts {
		t.Fatal("Expected TLSAcceptAllCerts:", defaultTLSAcceptAllCerts, " result:", *e.TLSAcceptAllCerts)
	}
}
