package mmail

import (
	"encoding/base64"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/jhillyerd/enmime"
	"github.com/pkg/errors"
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

// ReadMailMessage convert net/mail in MailMessage
func ReadMailMessage(r io.Reader) (*MailMessage, error) {
	mm := &MailMessage{}
	env, err := enmime.ReadEnvelope(r) // read message body with enmime
	if err != nil {
		return nil, errors.Wrap(err, "read message body")
	}

	mm.From = env.GetHeader("From")
	mm.Subject = env.GetHeader("Subject")
	mm.EmailText = env.Text

	var emailbody string
	if len(env.HTML) > 0 {
		mm.EmailType = EmailTypeHTML
		emailbody = env.HTML
		for _, p := range env.Inlines {
			emailbody = replaceCID(emailbody, p)
		}

		for _, p := range env.OtherParts {
			emailbody = replaceCID(emailbody, p)
		}

	} else {
		mm.EmailType = EmailTypeText
		emailbody = env.Text
	}

	mm.EmailBody = emailbody

	mm.Attachments = make([]*Attachment, len(env.Attachments))

	for i, a := range env.Attachments {
		mm.Attachments[i] = &Attachment{
			Filename: removeNonUTF8(a.FileName),
			Content:  a.Content,
		}
	}
	return mm, nil
}

//Replace cid:**** by embedded base64 image
func replaceCID(html string, part *enmime.Part) string {
	cid := strings.Replace(part.Header.Get("Content-ID"), "<", "", -1)
	cid = strings.Replace(cid, ">", "", -1)

	if len(cid) == 0 {
		return html
	}

	b64 := "data:" + part.ContentType + ";base64," + base64.StdEncoding.EncodeToString(part.Content)

	return strings.Replace(html, "cid:"+cid, b64, -1)
}

func removeNonUTF8(s string) string {
	if !utf8.ValidString(s) {
		v := make([]rune, 0, len(s))
		for i, r := range s {
			if r == utf8.RuneError {
				_, size := utf8.DecodeRuneInString(s[i:])
				if size == 1 {
					continue
				}
			}
			v = append(v, r)
		}
		s = string(v)
	}
	return s
}
