package mmail

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
	"crypto/tls"

	"github.com/jhillyerd/go.enmime"
	"github.com/mattermost/platform/model"
	"github.com/mxk/go-imap/imap"
	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/data"
)

// MatterMail struct with configurations, loggers and Mattemost user
type MatterMail struct {
	cfg        *Config
	imapClient *imap.Client
	info       *log.Logger
	eror       *log.Logger
	debg       *log.Logger
	user       *model.User
}

func (m *MatterMail) tryTime(message string, fn func() error) {
	if err := fn(); err != nil {
		m.info.Println(message, err, "\n", "Try again in 30s")
		time.Sleep(30 * time.Second)
		fn()
	}
}

// LogoutImapClient logout imap connection after 5 seconds
func (m *MatterMail) LogoutImapClient() {
	if m.imapClient != nil {
		m.imapClient.Logout(time.Second * 5)
	}
}

// CheckImapConnection if is connected return nil or try to connect
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
		var tconfig tls.Config
		if m.cfg.TLSAcceptAllCerts {
			tconfig.InsecureSkipVerify = true
		}
		_, err = m.imapClient.StartTLS(&tconfig)
		if err != nil {
			return err
		}
	}

	//Check if server support IDLE mode
	/*
		if !m.imapClient.Caps["IDLE"] {
			return fmt.Errorf("The server %q does not support IDLE\n", m.cfg.ImapServer)
		}
	*/
	m.info.Printf("Connected with %q\n", m.cfg.ImapServer)

	_, err = m.imapClient.Login(m.cfg.Email, m.cfg.EmailPass)
	if err != nil {
		m.eror.Println("Unable to login:", m.cfg.Email)
		return err
	}

	return nil
}

// CheckNewMails Check if exist a new mail and post it
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
		m.debg.Println("Error UIDSearch UTF-8:")
		m.debg.Println(err)
		m.debg.Println("Try with US-ASCII")

		// try again with US-ASCII
		cmd, err = imap.Wait(m.imapClient.Send("UID SEARCH", append([]imap.Field{"CHARSET", "US-ASCII"}, specs...)...))
		if err != nil {
			m.eror.Println("UID SEARCH US-ASCII")
			return err
		}
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
		}
		m.eror.Println("Fetch error:", rsp.Info)
		return err
	}

	cmd.Data = nil

	if postmail {
		m.debg.Println("CheckNewMails: Mark all messages with flag \\Seen")

		//Mark all messages seen
		_, err = imap.Wait(m.imapClient.UIDStore(seq, "+FLAGS.SILENT", `\Seen`))
		if err != nil {
			m.eror.Println("Error UIDStore \\Seen")
			return err
		}
	}

	return nil
}

// IdleMailBox Change to state idle in imap server
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
	timeout := 0
	for {
		err := m.imapClient.Recv(time.Second)
		timeout++
		if err == nil || timeout > 180 {
			break
		}
	}
	return nil
}

// postMessage Create a post in Mattermost
func (m *MatterMail) postMessage(client *model.Client, channelID string, message string, filenames *[]string) error {
	post := &model.Post{ChannelId: channelID, Message: message}

	if filenames != nil && len(*filenames) > 0 && m.cfg.NoAttachments != true {
		post.Filenames = *filenames
	}

	res, err := client.CreatePost(post)
	if res == nil {
		return err
	}

	return nil
}

// PostFile Post files and message in Mattermost server
func (m *MatterMail) PostFile(from, subject, message, emailname string, emailbody *string, attach *[]enmime.MIMEPart) error {

	client := model.NewClient(m.cfg.Server)

	m.debg.Println(client)

	m.debg.Printf("Login user:%v team:%v url:%v\n", m.cfg.MattermostUser, m.cfg.Team, m.cfg.Server)

	result, apperr := client.Login(m.cfg.MattermostUser, m.cfg.MattermostPass)
	if apperr != nil {
		return apperr
	}

	m.user = result.Data.(*model.User)

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
	var channelID, channelName string
	channelList := client.Must(client.GetChannels("")).Data.(*model.ChannelList)

	// redirect email by the subject
	if !m.cfg.NoRedirectChannel {
		m.debg.Println("Try to find channel/user by subject")
		channelName = getChannelFromSubject(subject)
		channelID = m.getChannelID(client, channelList, channelName)
	}

	// check filters
	if channelID == "" && m.cfg.Filter != nil {
		m.debg.Println("Did not find channel/user from Email Subject. Look for filter")
		channelName = m.cfg.Filter.GetChannel(from, subject)
		channelID = m.getChannelID(client, channelList, channelName)
	}

	// get default Channel config
	if channelID == "" {
		m.debg.Printf("Did not find channel/user in filters. Look for channel '%v'\n", m.cfg.Channel)
		channelName = m.cfg.Channel
		channelID = m.getChannelID(client, channelList, channelName)
	}

	if channelID == "" && !m.cfg.NoRedirectChannel {
		m.debg.Printf("Did not find channel/user with name '%v'. Trying channel town-square\n", m.cfg.Channel)
		channelName = "town-square"
		channelID = m.getChannelID(client, channelList, channelName)
	}

	if channelID == "" {
		return fmt.Errorf("Did not find any channel to post")
	}

	m.debg.Printf("Post email in %v", channelName)

	if len(*attach) == 0 && len(emailname) == 0 {
		return m.postMessage(client, channelID, message, nil)
	}

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	var email []byte

	addPart := func(filename string) error {
		part, err := writer.CreateFormFile("files", filename)
		if err != nil {
			return err
		}

		if _, err = part.Write(email); err != nil {
			return err
		}
		return nil
	}

	if len(emailname) > 0 {
		email = []byte(*emailbody)
		if err := addPart(emailname); err != nil {
			return err
		}
	}

	for _, a := range *attach {
		email = a.Content()
		if err := addPart(a.FileName()); err != nil {
			return err
		}
	}

	field, err := writer.CreateFormField("channel_id")
	if err != nil {
		return err
	}

	_, err = field.Write([]byte(channelID))
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

	return m.postMessage(client, channelID, message, &resp.Data.(*model.FileUploadResponse).Filenames)
}

func (m *MatterMail) getChannelID(client *model.Client, channelList *model.ChannelList, channelName string) string {
	if strings.HasPrefix(channelName, "#") {
		return getChannelIDByName(channelList, strings.TrimPrefix(channelName, "#"))
	} else if strings.HasPrefix(channelName, "@") {
		return m.getDirectChannelIDByName(client, channelList, strings.TrimPrefix(channelName, "@"))
	}
	return ""
}

func getChannelIDByName(channelList *model.ChannelList, channelName string) string {
	for _, c := range channelList.Channels {
		if c.Name == channelName {
			return c.Id
		}
	}
	return ""
}

func (m *MatterMail) getDirectChannelIDByName(client *model.Client, channelList *model.ChannelList, userName string) string {

	if m.user.Username == userName {
		m.eror.Printf("Impossible create a Direct channel, Mattermail user (%v) equals destination user (%v)\n", m.user.Username, userName)
		return ""
	}

	result, err := client.GetProfilesForDirectMessageList(client.GetTeamId())

	if err != nil {
		m.eror.Println("Error on GetProfilesForDirectMessageList: ", err.Error())
		return ""
	}

	profiles := result.Data.(map[string]*model.User)
	var userID string

	for k, p := range profiles {
		if p.Username == userName {
			userID = k
			break
		}
	}

	if userID == "" {
		m.debg.Println("Did not find the username:", userName)
		return ""
	}

	dmName := model.GetDMNameFromIds(m.user.Id, userID)
	dmID := getChannelIDByName(channelList, dmName)

	if dmID != "" {
		return dmID
	}

	m.debg.Println("Create direct channel to user:", userName)

	result, err = client.CreateDirectChannel(userID)
	if err != nil {
		m.eror.Println("Error on CreateDirectChannel: ", err.Error())
		return ""
	}

	directChannel := result.Data.(*model.Channel)
	return directChannel.Id
}

var channelRegex = regexp.MustCompile(`^\s*?\[\s*?([#@][A-Za-z0-9\-_]*)\s*?\]`)

// getChannelFromSubject extract channel from subject ex:
// getChannelFromSubject([#mychannel] blablanla) => #mychannel
func getChannelFromSubject(subject string) string {
	ret := channelRegex.FindStringSubmatch(subject)
	if len(ret) < 2 {
		return ""
	}
	return strings.ToLower(ret[1])
}

//Read number of lines of string
func readLines(s string, nmax int) string {
	if nmax <= 0 {
		return ""
	}

	var rxlines string
	if strings.Contains(s, "\r\n") {
		rxlines = "\r\n"
	} else {
		rxlines = "\n"
	}

	lines := regexp.MustCompile(rxlines).Split(s, nmax+1)

	ret := ""
	for i, l := range lines {
		if i >= nmax {
			break
		}
		if i > 0 {
			ret += rxlines
		}
		ret += l
	}
	if nmax+1 == len(lines) && strings.HasSuffix(s, rxlines) {
		ret += rxlines
	}
	return ret
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

// NonASCII Decode non ASCII header string RFC 1342
// encoded-word = "=?" charset "?" encoding "?" encoded-text "?="
func NonASCII(encoded string) string {

	regexRFC1342, _ := regexp.Compile(`=\?[^\?]*\?.\?[^\?]*\?=`)

	result := regexRFC1342.ReplaceAllStringFunc(encoded, func(encoded string) string {
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

// PostMail Post an email in Mattermost
func (m *MatterMail) PostMail(msg *mail.Message) error {
	mime, _ := enmime.ParseMIMEBody(msg) // Parse message body with enmime

	// read only some lines of text
	partmessage := readLines(mime.Text, m.cfg.LinesToPreview)

	postedfullmessage := false

	if partmessage != mime.Text && len(partmessage) > 0 {
		partmessage += " ..."
	} else if partmessage == mime.Text {
		postedfullmessage = true
	}

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

	} else if len(mime.Text) > 0 && !postedfullmessage {
		emailname = "email.txt"
		emailbody = mime.Text
	}
	
	re := regexp.MustCompile(m.cfg.ReplyDelimiter)
	locs := re.FindStringIndex(partmessage)
	if  locs != nil {
		rIndex := locs[0]
		partmessage = partmessage[:rIndex]
	}
	

	subject := mime.GetHeader("Subject")
	from := NonASCII(msg.Header.Get("From"))
	message := fmt.Sprintf(m.cfg.MailTemplate, from, subject, partmessage)

	return m.PostFile(from, subject, message, emailname, &emailbody, &mime.Attachments)
}

type devNull int

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

// InitMatterMail init MatterMail server
func InitMatterMail(cfg *Config) {
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

	m.debg.Println("Debug mode on")
	m.info.Println("Checking new emails")
	m.tryTime("Error on check new email:", m.CheckNewMails)
	m.info.Println("Waiting new messages")

	for {
		m.tryTime("Error Idle:", m.IdleMailBox)
		m.tryTime("Error on check new email:", m.CheckNewMails)
	}
}
