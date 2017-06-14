package mmail

import (
	"net/mail"
	"os"
	"testing"

	"github.com/rodcorsi/mattermail/model"
)

func TestCreateMattermostPost(t *testing.T) {
	cfg := model.NewProfile()
	cfg.Name = "test"
	cfg.Channels = []string{"#channel1"}
	*cfg.MailTemplate = "{{.From}}|{{.Subject}}|{{.Message}}"
	*cfg.LinesToPreview = 1

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

	if _, ok := mP.channelMap["#channel1"]; !ok {
		t.Fatalf("expected #channel1 result:'%v'", mP.channelMap)
	}

	if mP.message != "jdoe@example.com|Subject|line one ..." {
		t.Fatalf("expected 'jdoe@example.com|Subject|line one ...' result:'%v'", mP.message)
	}

	if len(mP.attachments) != 2 {
		t.Fatalf("expected 2 attachments found %v", len(mP.attachments))
	}

	// Subject
	*cfg.LinesToPreview = 10
	msg.Subject = "[@user2] subject 2"
	mP, err = createMattermostPost(msg, cfg, log, getChannelID)
	if err != nil {
		t.Fatalf("error on create mattermostPost %v", err)
	}

	if _, ok := mP.channelMap["@user2"]; !ok {
		t.Fatalf("expected @user2 result:'%v'", mP.channelMap)
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

	cfg.Filter = &model.Filter{&model.Rule{From: "jdoe@example.com", Channel: "#channel2"}}
	mP, err = createMattermostPost(msg, cfg, log, getChannelID)
	if err != nil {
		t.Fatalf("error on create mattermostPost %v", err)
	}

	if _, ok := mP.channelMap["#channel2"]; !ok {
		t.Fatalf("expected #channel2 result:'%v'", mP.channelMap)
	}

	if len(mP.attachments) != 2 {
		t.Fatalf("expected 2 attachments found %v", len(mP.attachments))
	}
}

type mattermostMock struct{}

func (m *mattermostMock) Login() error                           { return nil }
func (m *mattermostMock) Logout() error                          { return nil }
func (m *mattermostMock) GetChannelID(channelName string) string { return "id1234" }

func (m *mattermostMock) PostMessage(message, channelID string, attachments []*Attachment) error {
	return nil
}

func TestMatterMail_PostNetMail(t *testing.T) {
	// gmail
	gmailbuf, err := os.Open(findDir("emltest") + "gmail.eml")
	if err != nil {
		t.Fatal("Error on open gmail.eml:", err)
	}

	msg, err := mail.ReadMessage(gmailbuf)
	if err != nil {
		t.Fatalf("Failed parsing email:%v", err)
	}

	profile := model.NewProfile()
	profile.Channels = []string{"#town-square"}

	mm := NewMatterMail(profile, NewLog("", false), nil, &mattermostMock{})

	if err := mm.PostNetMail(msg); err != nil {
		t.Fatal("Error on PostNetMail err:", err.Error())
	}
}
