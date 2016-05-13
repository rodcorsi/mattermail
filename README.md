# ![mattermail icon](https://github.com/rodrigocorsi2/mattermail/raw/master/img/icon.png) MatterMail #

*MatterMail* is an integration service for [Mattermost](http://www.mattermost.org/), *MatterMail* listen an email box and publish all received emails in a channel or private group in Mattermost.

![mattermail screenshot](https://github.com/rodrigocorsi2/mattermail/raw/master/img/screenshot.png)

## Building
You need [Go](http://golang.org) to build this project

```bash
$ go get github.com/tools/godep
$ go get github.com/rodrigocorsi2/mattermail
```

## Usage
1. You need to create a user in Mattermost server, you can use MatterMail icon as profile picture.

2. Get the *Channel Handle* of the channel and check if the user has permission to post in this channel
![mattermail channel_handle](https://github.com/rodrigocorsi2/mattermail/raw/master/img/channel_handle.png)

3. Edit the file conf.json, e.g.:

```javascript
[
	{
		"Name":          "Orders",
		"Server":        "https://mattermost.example.com",
		"Team":          "team1",
		"Channel":       "orders",
		"MattermostUser":"mattermail@example.com",
		"MattermostPass":"password",
		"ImapServer":    "imap.example.com:143",
		"Email":         "orders@example.com",
		"EmailPass":     "password",
		"MailTemplate":  ":incoming_envelope: _From: **%v**_\n>_%v_\n\n%v",
		"StartTLS":      true,   /*Optional default false*/
		"Disabled":      false,  /*Optional default false*/
		"Debug":         false   /*Optional default false*/
	},
	{
		"Name":          "Bugs",
		"Server":        "https://mattermost.example.com",
		"Team":          "team1",
		"Channel":       "bugs",
		"MattermostUser":"mattermail@example.com",
		"MattermostPass":"password",
		"ImapServer":    "imap.gmail.com:993",
		"Email":         "bugs@gmail.com",
		"EmailPass":     "password",
		"MailTemplate":  ":incoming_envelope: _From: **%v**_\n>_%v_\n\n%v",
		"StartTLS":      false,  /*Optional default false*/
		"Disabled":      false,  /*Optional default false*/
		"Debug":         true    /*Optional default false*/
	},
	{
		/*.... other if you want ....*/
	}
]
```

4. Execute the command to put in background

```
$ ./mattermail > /var/log/mattermail.log 2>&1 &
```
