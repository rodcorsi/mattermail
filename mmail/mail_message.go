package mmail

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"

	"github.com/jhillyerd/go.enmime"
)

// Emails content type
const (
	EmailTypeHTML = iota
	EmailTypeText
)

// Attachment filename and content
type Attachment struct {
	Filename string
	Content  []byte
}

// MailMessage mail message with fields used in mattermail
type MailMessage struct {
	From        string
	Subject     string
	EmailText   string
	EmailBody   string
	EmailType   int
	Attachments []*Attachment
}

// ParseMailMessage convert net/mail in MailMessage
func ParseMailMessage(msg *mail.Message) (*MailMessage, error) {
	mm := &MailMessage{}

	mime, err := enmime.ParseMIMEBody(msg) // Parse message body with enmime

	if err != nil {
		return nil, fmt.Errorf("Error on parse mail.Message:%v", err.Error())
	}

	mm.From = NonASCII(msg.Header.Get("From"))
	mm.Subject = mime.GetHeader("Subject")
	mm.EmailText = mime.Text

	var emailbody string
	if len(mime.HTML) > 0 {
		mm.EmailType = EmailTypeHTML
		emailbody = mime.HTML
		for _, p := range mime.Inlines {
			emailbody = replaceCID(emailbody, p)
		}

		for _, p := range mime.OtherParts {
			emailbody = replaceCID(emailbody, p)
		}

	} else {
		mm.EmailType = EmailTypeText
		emailbody = mime.Text
	}

	mm.EmailBody = emailbody

	mm.Attachments = make([]*Attachment, len(mime.Attachments))

	for i, a := range mime.Attachments {
		mm.Attachments[i] = &Attachment{
			Filename: a.FileName(),
			Content:  a.Content(),
		}
	}
	return mm, nil
}

//Replace cid:**** by embedded base64 image
func replaceCID(html string, part enmime.MIMEPart) string {
	cid := strings.Replace(part.Header().Get("Content-ID"), "<", "", -1)
	cid = strings.Replace(cid, ">", "", -1)

	if len(cid) == 0 {
		return html
	}

	b64 := "data:" + part.ContentType() + ";base64," + base64.StdEncoding.EncodeToString(part.Content())

	return strings.Replace(html, "cid:"+cid, b64, -1)
}
