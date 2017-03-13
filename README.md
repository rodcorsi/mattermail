# ![mattermail icon](https://github.com/rodcorsi/mattermail/raw/master/img/icon.png) MatterMail

*MatterMail* is an integration service for [Mattermost](http://www.mattermost.org/), *MatterMail* listen an email box and publish all received emails in a channel or private group in Mattermost.

[![Build Status](https://travis-ci.org/rodrigocorsi2/mattermail.svg?branch=master)](https://travis-ci.org/rodrigocorsi2/mattermail)
[![Coverage Status](https://coveralls.io/repos/github/rodrigocorsi2/mattermail/badge.svg?branch=master)](https://coveralls.io/github/rodrigocorsi2/mattermail?branch=master)

![mattermail screenshot](https://github.com/rodcorsi/mattermail/raw/master/img/screenshot.png)

## Redirect to the channel by subject (Version 3.0 or later)

Mattermail post the email using this rules (if "`NoRedirectChannel:false`"):

1 - If the email subject contains "`[#anychannelname] blablabla`" or "`[@usertosend] xxxxxx`", Mattermail will try to post to the channel or to the username

2 - If the email subject doesn't contain channel or username, Mattermail will try to post the channel defined in `config.json`

3 - If Mattermail can not post the email will try to post in "Town Square"


## Install
  * For Mattermost 3.0 or later: [Latest Version](https://github.com/rodcorsi/mattermail/releases/latest)

  * For Mattermost 2.2: [Version 2.2](https://github.com/rodcorsi/mattermail/releases/tag/v2.2)

## Usage
1. You need to create a user in Mattermost server, you can use MatterMail icon as profile picture.

2. Get the *Channel Handle* of the channel and check if the user has permission to post in this channel
![mattermail channel_handle](https://github.com/rodcorsi/mattermail/raw/master/img/channel_handle.png)

3. Edit the file config.json, e.g.:

```javascript
[
	{
		"Name":          "Orders",
		"Server":        "https://mattermost.example.com",
		"Team":          "team1",
		"Channel":       "#orders",
		"MattermostUser":"mattermail@example.com",
		"MattermostPass":"password",
		"ImapServer":    "imap.example.com:143",
		"Email":         "orders@example.com",
		"EmailPass":     "password",
		"MailTemplate":  ":incoming_envelope: _From: **%v**_\n>_%v_\n\n%v"
	},
	{
		"Name":              "Bugs",
		"Server":            "https://mattermost.example.com",
		"Team":              "team1",
		"Channel":           "@user123",
		"MattermostUser":    "mattermail@example.com",
		"MattermostPass":    "password",
		"ImapServer":        "imap.gmail.com:993",
		"Email":             "bugs@gmail.com",
		"EmailPass":         "password",
		"MailTemplate":      ":incoming_envelope: _From: **%v**_\n>_%v_\n\n%v",
		"StartTLS":          false,  /*Optional default false*/
		"TLSAcceptAllCerts": false,  /*Optional default false*/
		"Disabled":          false,  /*Optional default false*/
		"Debug":             true    /*Optional default false*/
        "LinesToPreview":    20,     /*Optional default 10*/
		"NoRedirectChannel": true,   /*Optional default false*/
        "NoAttachment":      true,   /*Optional leave out attachments*/

        /*Filter works only (Version 3.0 or later)*/
        "Filter":            [
            /* if subject contains 'Feature' redirect to #feature */
            {"Subject":"Feature", "Channel":"#feature"},
            
            /* if from contains 'test@gmail.com' and subject contains 'to me' redirect to @test2*/
            {"From":"test@gmail.com", "Subject":"To Me", "Channel":"@test2"},
            
            /* if from contains '@companyb.com' redirect to #companyb */
			{"From":"@companyb.com", "Channel":"#companyb"} /**/
        ]
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
## Options

```bash
$ ./mattermail --help
Options:
    -c, --config  Sets the file location for config.json
                  Default: ./config.json 
    -h, --help    Show this help
    -v, --version Print current version
```

## Building
You need [Go](http://golang.org) to build this project

```bash
$ go get github.com/rodcorsi/mattermail
```

### If you want to build MatterMail to Mattermost 2.2 you need to use `release-2.2` branch

