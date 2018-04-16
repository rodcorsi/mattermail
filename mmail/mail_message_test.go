package mmail

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestReadMailMessage(t *testing.T) {
	email := `From: John Doe <jdoe@machine.example>
To: Mary Smith <mary@example.net>
Subject: Saying Hello
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.machine.example>

This is a message just to say hello.
`
	expected := &MailMessage{
		From:      "John Doe <jdoe@machine.example>",
		Subject:   "Saying Hello",
		EmailText: "This is a message just to say hello.\n",
		EmailType: EmailTypeText,
	}

	if err := testMailMessage(bytes.NewBuffer([]byte(email)), expected); err != nil {
		t.Fatal("Error on plain/text:", err.Error())
	}

	// gmail
	gmailbuf, err := os.Open(findDir("emltest") + "gmail.eml")
	if err != nil {
		t.Fatal("Error on open gmail.eml:", err)
	}

	expected = &MailMessage{
		From:      "Rodrigo <test@gmail.com>",
		Subject:   "Orçamento Teste",
		EmailText: "This is a test\n",
		EmailType: EmailTypeHTML,
		Attachments: []*Attachment{{
			Filename: "attach.bak",
			Content:  []byte("attachment"),
		}},
	}

	if err := testMailMessage(gmailbuf, expected); err != nil {
		t.Fatal("Error on gmail:", err.Error())
	}

	// thunderbird
	thunderbuf, err := os.Open(findDir("emltest") + "thunderbird.eml")
	if err != nil {
		t.Fatal("Error on open thunderbird.eml:", err)
	}

	expected = &MailMessage{
		From:      "Rodrigo <test@gmail.com>",
		Subject:   "Orçamento Teste",
		EmailText: "This is a test",
		EmailType: EmailTypeHTML,
		Attachments: []*Attachment{{
			Filename: "attach.bak",
			Content:  []byte("attachment"),
		}},
	}

	if err := testMailMessage(thunderbuf, expected); err != nil {
		t.Fatal("Error on thunderbird:", err.Error())
	}

	// roundcube
	roundbuf, err := os.Open(findDir("emltest") + "roundcube.eml")
	if err != nil {
		t.Fatal("Error on open roundcube.eml:", err)
	}

	expected = &MailMessage{
		From:      "test@example.com",
		Subject:   "Orçamento Teste",
		EmailText: "This is a test\n",
		EmailType: EmailTypeHTML,
		Attachments: []*Attachment{{
			Filename: "attach.bak",
			Content:  []byte("attachment"),
		}},
	}

	if err := testMailMessage(roundbuf, expected); err != nil {
		t.Fatal("Error on roundcube:", err.Error())
	}
}

func TestReadMailMessageRFC2047(t *testing.T) {
	email := `From: =?ISO-2022-JP?B?GyRCOzNFREJATzobKEI=?= <taro@example.com>
To: Mary Smith <mary@example.net>
Subject: Fwd: =?UTF-8?B?Q290YcOnw6NvIGRlIEZlcnJhbWVudGFz?=
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.machine.example>

`
	expected := &MailMessage{
		From:      "山田太郎 <taro@example.com>",
		Subject:   "Fwd: Cotação de Ferramentas",
		EmailType: EmailTypeText,
	}

	if err := testMailMessage(bytes.NewBuffer([]byte(email)), expected); err != nil {
		t.Fatal("Error on plain/text:", err.Error())
	}

	email = `From: =?ISO-8859-1?Q?Keld_J=F8rn_Simonsen?= <keld@dkuug.dk>
To: Mary Smith <mary@example.net>
Subject: =?ISO-8859-2?B?dSB1bmRlcnN0YW5kIHRoZSBleGFtcGxlLg==?=
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.machine.example>

`
	expected = &MailMessage{
		From:      "Keld Jørn Simonsen <keld@dkuug.dk>",
		Subject:   "u understand the example.",
		EmailType: EmailTypeText,
	}

	if err := testMailMessage(bytes.NewBuffer([]byte(email)), expected); err != nil {
		t.Fatal("Error on plain/text:", err.Error())
	}

	email = `From: =?US-ASCII?Q?Keith_Moore?= <moore@cs.utk.edu>
To: Mary Smith <mary@example.net>
Subject: =?ISO-8859-1?Q?Patrik_F=E4ltstr=F6m?= <paf@nada.kth.se>
Date: Fri, 21 Nov 1997 09:55:06 -0600
Message-ID: <1234@local.machine.example>

`
	expected = &MailMessage{
		From:      "Keith Moore <moore@cs.utk.edu>",
		Subject:   "Patrik Fältström <paf@nada.kth.se>",
		EmailType: EmailTypeText,
	}

	if err := testMailMessage(bytes.NewBuffer([]byte(email)), expected); err != nil {
		t.Fatal("Error on plain/text:", err.Error())
	}
}

func testMailMessage(r io.Reader, expected *MailMessage) error {
	mm, err := ReadMailMessage(r)
	if err != nil {
		return fmt.Errorf("Failed to parsing msg:%v", err)
	}

	if mm.From != expected.From {
		return fmt.Errorf("field From expected: '%v' found:'%v'", expected.From, mm.From)
	}

	if mm.Subject != expected.Subject {
		return fmt.Errorf("field Subject expected: '%v' found:'%v'", expected.Subject, mm.Subject)
	}

	if mm.EmailText != expected.EmailText {
		return fmt.Errorf("field EmailText expected: '%v' found:'%v'", expected.EmailText, mm.EmailText)
	}

	if mm.EmailType != expected.EmailType {
		return fmt.Errorf("field EmailType expected: '%v' found:'%v'", expected.EmailType, mm.EmailType)
	}

	if len(mm.Attachments) != len(expected.Attachments) {
		return fmt.Errorf("number of attachments expected: '%v' found:'%v'", len(expected.Attachments), len(mm.Attachments))
	}

	for i := range mm.Attachments {
		if mm.Attachments[i].Filename != expected.Attachments[i].Filename {
			return fmt.Errorf("different filename on index:%v expected:'%v' found:'%v'", i, mm.Attachments[i].Filename, expected.Attachments[i].Filename)
		}

		if bytes.Compare(mm.Attachments[i].Content, expected.Attachments[i].Content) != 0 {
			return fmt.Errorf("different content on index:%v expected:'%v' found:'%v'", i, mm.Attachments[i].Content, expected.Attachments[i].Content)
		}
	}
	return nil
}

func Test_removeNonUTF8(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"1", "foo", "foo"},
		{"2", "a\xc5z", "az"},
		{"3", "b\xe7\xe3\x6fc", "boc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeNonUTF8(tt.value); got != tt.want {
				t.Errorf("removeNonUTF() = %v, want %v", got, tt.want)
			}
		})
	}
}
