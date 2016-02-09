# ![mattermail icon](https://github.com/rodrigocorsi2/mattermail/raw/master/img/icon.png) MatterMail #

*MatterMail* is an integration service for [Mattermost](http://www.mattermost.org/), *MatterMail* listen an email box and publish all received emails in a channel or private group in Mattermost.

![mattermail screenshot](https://github.com/rodrigocorsi2/mattermail/raw/master/img/screenshot.png)

## Building
You need [Go](http://golang.org) to build this project

```
$ go get github.com/jhillyerd/go.enmime
$ go get github.com/mattermost/platform/model
$ go get github.com/mxk/go-imap/imap
$ go get github.com/paulrosania/go-charset/charset
$ go get github.com/paulrosania/go-charset/data
$ go get github.com/rodrigocorsi2/mattermail
```
	
## Usage
1. You need to create a user in Mattermost server
2. Get channel id click on channel name > View Info 
3. Edit the file conf.json, e.g.:
```
[
	{
		"Name":          "Orders",
		"Server":        "https://mattermost.example.com",
		"Team":          "team1",
		"ChannelId":     "euw6fsafdasybyaixinrqxjme3enw",
		"MattermostUser":"orders@example.com",
		"MattermostPass":"password",
		"ImapServer":    "imap.example.com:143",
		"Email":         "orders@example.com",
		"EmailPass":     "password",
		"MailTemplate":  ">:incoming_envelope: _From:_ **%v**\n_%v_\n\n%v"
	},
	{
		.... other if you want ....
	}
]
```

4. Execute the command to put in background
```
$ ./mattermail > /var/log/mattermail.log 2>&1 &
```