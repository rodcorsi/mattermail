package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jhillyerd/go.enmime"
	"github.com/mattermost/platform/model"
	"github.com/mxk/go-imap/imap"
	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/data"
)

//Number of lines of email to show in post
const linestopreview = 10

type MatterMail struct {
	cfg        *config
	imapClient *imap.Client
	info       *log.Logger
	eror       *log.Logger
	debg       *log.Logger
}

func (m *MatterMail) tryTime(message string, fn func() error) {
	if err := fn(); err != nil {
		m.info.Println(message, err, "\n", "Try again in 30s")
		time.Sleep(30 * time.Second)
		fn()
	}
}

func (m *MatterMail) LogoutImapClient() {
	if m.imapClient != nil {
		m.imapClient.Logout(time.Second * 5)
	}
}

func (m *MatterMail) CheckImapConnection() error {
	if m.imapClient != nil && (m.imapClient.State() == imap.Auth || m.imapClient.State() == imap.Selected) {
		m.debg.Println("CheckImapConnection: Connection alive")
		return nil
	}

	var err error

	//Start connection with server
	if strings.HasSuffix(m.cfg.ImapServer, ":993") {
		m.debg.Println("CheckImapConnection: DialTLS")
		m.imapClient, err = imap.DialTLS(m.cfg.ImapServer, nil)
	} else {
		m.debg.Println("CheckImapConnection: Dial")
		m.imapClient, err = imap.Dial(m.cfg.ImapServer)
	}

	if err != nil {
		m.eror.Println("Unable to connect:", err)
		return err
	}

	if m.cfg.StartTLS && m.imapClient.Caps["STARTTLS"] {
		m.debg.Println("CheckImapConnection:StartTLS")
		_, err = m.imapClient.StartTLS(nil)
		if err != nil {
			return err
		}
	}

	//Check if server support IDLE mode
	if !m.imapClient.Caps["IDLE"] {
		return fmt.Errorf("The server %q does not support IDLE\n", m.cfg.ImapServer)
	}

	m.info.Printf("Connected with %q\n", m.cfg.ImapServer)

	_, err = m.imapClient.Login(m.cfg.Email, m.cfg.EmailPass)
	if err != nil {
		m.eror.Println("Unable to login:", m.cfg.Email)
		return err
	}

	return nil
}

//Check if exist a new mail and post it
func (m *MatterMail) CheckNewMails() error {
	m.debg.Println("CheckNewMails")

	if err := m.CheckImapConnection(); err != nil {
		return err
	}

	var (
		cmd *imap.Command
		rsp *imap.Response
	)

	// Open a mailbox (synchronous command - no need for imap.Wait)
	m.imapClient.Select("INBOX", false)

	var specs []imap.Field
	specs = append(specs, "UNSEEN")
	seq := &imap.SeqSet{}

	// get headers and UID for UnSeen message in src inbox...
	cmd, err := imap.Wait(m.imapClient.UIDSearch(specs...))
	if err != nil {
		m.eror.Println("UIDSearch:")
		return err
	}

	for _, rsp := range cmd.Data {
		for _, uid := range rsp.SearchResults() {
			m.debg.Println("CheckNewMails:AddNum ", uid)
			seq.AddNum(uid)
		}
	}

	// no new messages
	if seq.Empty() {
		m.debg.Println("CheckNewMails: No new messages")
		return nil
	}

	cmd, _ = m.imapClient.UIDFetch(seq, "BODY[]")
	postmail := false

	for cmd.InProgress() {
		m.debg.Println("CheckNewMails: cmd in Progress")
		// Wait for the next response (no timeout)
		m.imapClient.Recv(-1)

		// Process command data
		for _, rsp = range cmd.Data {
			msgFields := rsp.MessageInfo().Attrs
			header := imap.AsBytes(msgFields["BODY[]"])
			if msg, _ := mail.ReadMessage(bytes.NewReader(header)); msg != nil {
				m.debg.Println("CheckNewMails:PostMail")
				if err := m.PostMail(msg); err != nil {
					return err
				}
				postmail = true
			}
		}
		cmd.Data = nil
	}

	// Check command completion status
	if rsp, err := cmd.Result(imap.OK); err != nil {
		if err == imap.ErrAborted {
			m.eror.Println("Fetch command aborted")
			return err
		} else {
			m.eror.Println("Fetch error:", rsp.Info)
			return err
		}
	}

	cmd.Data = nil

	if postmail {
		m.debg.Println("CheckNewMails: Mark all messages with flag \\Seen")

		//Mark all messages seen
		_, err = imap.Wait(m.imapClient.UIDStore(seq, "+FLAGS.SILENT", `\Seen`))
		if err != nil {
			m.eror.Printf("Error UIDStore \\Seen")
			return err
		}
	}

	return nil
}

//Change to state idle in imap server
func (m *MatterMail) IdleMailBox() error {
	m.debg.Println("IdleMailBox")

	if err := m.CheckImapConnection(); err != nil {
		return err
	}

	// Open a mailbox (synchronous command - no need for imap.Wait)
	m.imapClient.Select("INBOX", false)

	_, err := m.imapClient.Idle()
	if err != nil {
		return err
	}

	defer m.imapClient.IdleTerm()

	for {
		err := m.imapClient.Recv(time.Second)
		if err == nil {
			break
		}
	}
	return nil
}

func addPart(client *model.Client, filename string, content *[]byte, writer *multipart.Writer) error {
	part, err := writer.CreateFormFile("files", filename)
	if err != nil {
		return err
	}

	_, err = part.Write(*content)
	if err != nil {
		return err
	}
	return nil
}

//Create a post in Mattermost
func (m *MatterMail) postMessage(client *model.Client, channel_id string, message string, filenames *[]string) error {
	post := &model.Post{ChannelId: channel_id, Message: message}

	if len(*filenames) > 0 {
		post.Filenames = *filenames
	}

	res, err := client.CreatePost(post)
	if res == nil {
		return err
	}

	return nil
}

//Post files and message in Mattermost server
func (m *MatterMail) PostFile(message string, emailname string, emailbody *string, attach *[]enmime.MIMEPart) error {

	client := model.NewClient(m.cfg.Server)

	m.debg.Println("Login %v on %v", m.cfg.MattermostUser, m.cfg.Team)

	if _, err := client.Login(m.cfg.MattermostUser, m.cfg.MattermostPass); err != nil {
		return err
	}

	m.info.Println("Post new message")

	defer client.Logout()

	// Get Team
	teams := client.Must(client.GetAllTeams()).Data.(map[string]*model.Team)

	teamMatch := false
	for _, t := range teams {
		if t.Name == m.cfg.Team {
			client.SetTeamId(t.Id)
			teamMatch = true
			break
		}
	}

	if !teamMatch {
		return fmt.Errorf("Did not find team with name %v", m.cfg.Team)
	}

	//Discover channel id by channel name
	var channel_id string

	rget := client.Must(client.GetChannels("")).Data.(*model.ChannelList)

	nameMatch := false
	for _, c := range rget.Channels {
		if c.Name == m.cfg.Channel {
			channel_id = c.Id
			nameMatch = true
			break
		}
	}

	if !nameMatch {
		return fmt.Errorf("Did not find channel with name %v", m.cfg.Channel)
	}

	if len(*attach) == 0 && len(emailname) == 0 {
		return m.postMessage(client, channel_id, message, nil)
	}

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	var email []byte
	if len(emailname) > 0 {
		email = []byte(*emailbody)
		if err := addPart(client, emailname, &email, writer); err != nil {
			return err
		}
	}

	for _, a := range *attach {
		email = a.Content()
		if err := addPart(client, a.FileName(), &email, writer); err != nil {
			return err
		}
	}

	field, err := writer.CreateFormField("channel_id")
	if err != nil {
		return err
	}

	_, err = field.Write([]byte(channel_id))
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	resp, err := client.UploadPostAttachment(buf.Bytes(), writer.FormDataContentType())
	if resp == nil {
		return err
	}

	return m.postMessage(client, channel_id, message, &resp.Data.(*model.FileUploadResponse).Filenames)
}

//Read number of lines of string
func readLines(s string, nmax int) string {
	lines := regexp.MustCompile("\r\n|\n").Split(s, nmax+1)

	if nmax < len(lines) {
		return strings.Join(lines[:nmax], "\n")
	}
	return strings.Join(lines[:], "\n")
}

//Replace cid:**** by embedded base64 image
func replaceCID(html *string, part *enmime.MIMEPart) string {
	cid := strings.Replace((*part).Header().Get("Content-ID"), "<", "", -1)
	cid = strings.Replace(cid, ">", "", -1)

	if len(cid) == 0 {
		return *html
	}

	b64 := "data:" + (*part).ContentType() + ";base64," + base64.StdEncoding.EncodeToString((*part).Content())

	return strings.Replace(*html, "cid:"+cid, b64, -1)
}

//Decode non ASCII header string RFC 1342
//encoded-word = "=?" charset "?" encoding "?" encoded-text "?="
func NonASCII(encoded string) string {

	regex_rfc1342, _ := regexp.Compile(`=\?[^\?]*\?.\?[^\?]*\?=`)

	result := regex_rfc1342.ReplaceAllStringFunc(encoded, func(encoded string) string {
		//0 utf 1 B/Q 2 code
		v := strings.Split(encoded, "?")[1:4]
		var decoded string
		switch strings.ToLower(v[1]) {
		case "b": //Base64
			data, err := base64.StdEncoding.DecodeString(v[2])
			if err != nil {
				log.Println("Error decode Base64", err)
				return encoded
			}

			decoded = string(data)

		case "q": //Quoted-Printable
			data, err := ioutil.ReadAll(quotedprintable.NewReader(strings.NewReader(v[2])))
			if err != nil {
				log.Println("Error decode Quoted-Printable", err)
				return encoded
			}
			decoded = string(data)

		default:
			log.Println("Unknow encoding " + v[1])
			return encoded
		}

		//Decode charset
		r, err := charset.NewReader(strings.ToLower(v[0]), strings.NewReader(decoded))
		if err != nil {
			log.Println("Error decode charset", err)
			return encoded
		}

		result, _ := ioutil.ReadAll(r)

		return string(result)
	})

	return result
}

//Post an email in Mattermost
func (m *MatterMail) PostMail(msg *mail.Message) error {
	mime, _ := enmime.ParseMIMEBody(msg) // Parse message body with enmime

	var emailname, emailbody string
	if len(mime.HTML) > 0 {
		emailname = "email.html"
		emailbody = mime.HTML
		for _, p := range mime.Inlines {
			emailbody = replaceCID(&emailbody, &p)
		}

		for _, p := range mime.OtherParts {
			emailbody = replaceCID(&emailbody, &p)
		}

	} else if len(mime.Text) > 0 {
		emailname = "email.txt"
		emailbody = mime.Text
	}

	// read only some lines of text
	partmessage := readLines(mime.Text, linestopreview)

	if partmessage != mime.Text && len(partmessage) > 0 {
		partmessage += " ..."
	}

	message := fmt.Sprintf(m.cfg.MailTemplate, NonASCII(msg.Header.Get("From")), mime.GetHeader("Subject"), partmessage)

	return m.PostFile(message, emailname, &emailbody, &mime.Attachments)
}

type devNull int

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

func InitMatterMail(cfg *config) {
	m := &MatterMail{
		cfg:  cfg,
		info: log.New(os.Stdout, "INFO "+cfg.Name+"\t", log.Ltime),
		eror: log.New(os.Stderr, "EROR "+cfg.Name+"\t", log.Ltime),
	}

	if cfg.Debug {
		m.debg = log.New(os.Stdout, "DEBG "+cfg.Name+"\t", log.Ltime)
		imap.DefaultLogger = log.New(os.Stdout, "IMAP "+cfg.Name+"\t", log.Ltime)
		imap.DefaultLogMask = imap.LogConn | imap.LogRaw
	} else {
		m.debg = log.New(devNull(0), "", 0)
	}

	defer m.LogoutImapClient()

	m.info.Println("MatterMail version:", Version)
	m.debg.Println("Debug mode on")
	m.info.Println("Checking new emails")
	m.tryTime("Error on check new email:", m.CheckNewMails)
	m.info.Println("Waiting new messages")

	for {
		m.tryTime("Error Idle:", m.IdleMailBox)
		m.tryTime("Error on check new email:", m.CheckNewMails)
	}
}
