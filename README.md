# MatterMail #

*MatterMail* is a integration service for [Mattermost](http://www.mattermost.org/), *MatterMail* listen a email box and publish all received emails in a channel or private group in Mattermost.

## Building
You need [Go](http://golang.org) to build this project
```
$ cd $GOPATH
$ go get https://github.com/rodrigocorsi2/mattermail
$ cd mattermail
$ go build
```

## Usage
1. You need to create a user in Mattermost server
2. Get channel id click on channel name > View Info 
3. Edit the file conf.json, e.g.:

>
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