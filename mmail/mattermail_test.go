package mmail

import "testing"

func TestCreateMattermostPost(t *testing.T) {
	cfg := MatterMailConfig{
		Name:              "test",
		Channel:           "#channel1",
		MailTemplate:      "%v|%v|%v",
		LinesToPreview:    1,
		NoRedirectChannel: false,
		NoAttachment:      false,
	}

	log := NewLog("test", false)

	getChannelID := func(channelName string) string {
		return channelName
	}

	msg := &MailMessage{
		From:      "jdoe@example.com",
		Subject:   "Subject",
		EmailText: "line one\nline two",
		EmailType: EmailTypeText,
		Attachments: []*Attachment{{
			Filename: "file1.txt",
			Content:  []byte("text of file1"),
		}},
	}

	mP, err := createMattermostPost(msg, cfg, log, getChannelID)

	if err != nil {
		t.Fatalf("error on create mattermostPost %v", err)
	}

	if mP.channelName != "#channel1" {
		t.Fatalf("expected #channel1 result:'%v'", mP.channelName)
	}

	if mP.message != "jdoe@example.com|Subject|line one ..." {
		t.Fatalf("expected 'jdoe@example.com|Subject|line one ...' result:'%v'", mP.message)
	}

	if len(mP.attachments) != 2 {
		t.Fatalf("expected 2 attachments found %v", len(mP.attachments))
	}

	// Subject
	cfg.LinesToPreview = 10
	msg.Subject = "[@user2] subject 2"
	mP, err = createMattermostPost(msg, cfg, log, getChannelID)
	if err != nil {
		t.Fatalf("error on create mattermostPost %v", err)
	}

	if mP.channelName != "@user2" {
		t.Fatalf("expected @user2 result:'%v'", mP.channelName)
	}

	if mP.message != "jdoe@example.com|[@user2] subject 2|line one\nline two" {
		t.Fatalf("expected 'jdoe@example.com|[@user2] subject 2|line one\nline two' result:'%v'", mP.message)
	}

	if len(mP.attachments) != 1 {
		t.Fatalf("expected 1 attachment found %v", len(mP.attachments))
	}

	// Filter + HTML
	msg.Subject = "Subject"
	msg.EmailType = EmailTypeHTML

	cfg.Filter = &Filter{&Rule{From: "jdoe@example.com", Channel: "#channel2"}}
	mP, err = createMattermostPost(msg, cfg, log, getChannelID)
	if err != nil {
		t.Fatalf("error on create mattermostPost %v", err)
	}

	if mP.channelName != "#channel2" {
		t.Fatalf("expected #channel2 result:'%v'", mP.channelName)
	}

	if len(mP.attachments) != 2 {
		t.Fatalf("expected 2 attachments found %v", len(mP.attachments))
	}
}
