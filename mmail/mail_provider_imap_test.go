package mmail

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"net/mail"
	"sync/atomic"
	"testing"
	"time"

	imap "github.com/emersion/go-imap"
	idle "github.com/emersion/go-imap-idle"
	"github.com/emersion/go-imap/backend"
	"github.com/emersion/go-imap/backend/memory"
	"github.com/emersion/go-imap/server"
	"github.com/rodcorsi/mattermail/model"
)

type testServer struct {
	*server.Server
	addr string
	be   backend.Backend
}

var ts *testServer

const debugImap = false

func init() {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("Cannot listen:" + err.Error())
	}

	ts = &testServer{
		be: memory.New(),
	}

	ts.Server = server.New(ts.be)
	ts.Enable(idle.NewExtension())

	go ts.Serve(l)

	ts.addr = l.Addr().String()

	ts.AllowInsecureAuth = true

	_, err = net.Dial("tcp", ts.addr)
	if err != nil {
		panic("Cannot connect to server:" + err.Error())
	}
}

func TestCheckNewMessage(t *testing.T) {
	// Create a memory backend
	user, _ := ts.be.Login("username", "password")
	inbox, _ := user.GetMailbox("INBOX")

	email, _ := ioutil.ReadFile(findDir("emltest") + "gmail.eml")
	literal := bytes.NewBuffer(email)
	inbox.CreateMessage([]string{}, time.Now(), literal)

	email, _ = ioutil.ReadFile(findDir("emltest") + "thunderbird.eml")
	literal = bytes.NewBuffer(email)
	inbox.CreateMessage([]string{}, time.Now(), literal)

	config := model.NewEmail()
	config.Address = "username"
	config.Password = "password"
	config.ImapServer = ts.addr

	mP := NewMailProviderImap(config, NewLog("", debugImap), debugImap)

	defer mP.Terminate()

	var count uint32

	err := mP.CheckNewMessage(func(msg *mail.Message) error {
		if msg == nil {
			return errors.New("Messsage nil")
		}
		atomic.AddUint32(&count, 1)
		return nil
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	if count != 2 {
		t.Fatal("Expected 2 messages, received", count)
	}
}

func TestWaitNewMessage(t *testing.T) {
	config := model.NewEmail()
	config.Address = "username"
	config.Password = "password"
	config.ImapServer = ts.addr
	*config.StartTLS = true

	mP := NewMailProviderImap(config, NewLog("", debugImap), debugImap)

	done := make(chan error, 1)
	go func() {
		done <- mP.WaitNewMessage(60)
	}()

	defer mP.Terminate()

	time.Sleep(time.Second * 2)

	mP.imapClient.MailboxUpdates <- imap.NewMailboxStatus("INBOX", []string{"MESSAGES", "UNSEEN"})

	if err := <-done; err != nil {
		t.Fatal("Error WaitNewMessage:", err.Error())
	}
}
