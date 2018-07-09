# ![mattermail icon](https://github.com/rodcorsi/mattermail/raw/master/img/icon.png) MatterMail

_MatterMail_ is an integration service for [Mattermost](http://www.mattermost.org/), _MatterMail_ listen an email box and publish all received emails in channels or users in Mattermost.

[![Build Status](https://travis-ci.org/rodcorsi/mattermail.svg?branch=master)](https://travis-ci.org/rodcorsi/mattermail)
[![Coverage Status](https://coveralls.io/repos/github/rodcorsi/mattermail/badge.svg?branch=master)](https://coveralls.io/github/rodcorsi/mattermail?branch=master)

![mattermail screenshot](https://github.com/rodcorsi/mattermail/raw/master/img/screenshot.png)

## Install

Download the [Latest Version](https://github.com/rodcorsi/mattermail/releases/latest)

## Usage

1.  You need to create an user in Mattermost server and you can use MatterMail icon as profile picture.

1.  Get the [Team and Channels](https://github.com/rodcorsi/mattermail#teamchannel) and check if the user has permission to post in these channels

1.  Edit the file config.json

1.  Execute the command to put in background

```bash
./mattermail > /var/log/mattermail.log 2>&1 &
```

## Migrate configuration

To upgrade the config.json to new version using this command:

```bash
./mattermail migrate -c ./config.json > ./new_config.json
```

## Configuration

Minimal configuration:

```javascript
{
    "Directory": "./data/",
    "Profiles":[
        {
            "Name":              "Orders",
            "Channels":          ["#orders"],

            "Email":{
                "ImapServer":        "imap.example.com:143",
                "Username":          "orders@example.com",
                "Password":          "password"
            },

            "Mattermost":{
                "Server":   "https://mattermost.example.com",
                "Team":     "team1",
                "User":     "mattermail@example.com",
                "Password": "password"
            }
        }
    ]
}
```

### Directory

Location where the cache is stored, default value is `./data/`

### Profiles

You can set multiple profiles using different names

<<<<<<< HEAD
| Field             |  Type   | Default |     Obrigatory     | Information                                                                                               |
| ----------------- | :-----: | ------- | :----------------: | --------------------------------------------------------------------------------------------------------- |
| Name              | string  |         | :white_check_mark: | Name of profile, used to log                                                                              |
| Channels          |  array  |         | :white_check_mark: | List of channels where the email will be posted. You can use `#channel` or `@username`                    |
| Email             | object  |         | :white_check_mark: | Configuration of Email [(details)](https://github.com/rodcorsi/mattermail#email)                          |
| Mattermost        | object  |         | :white_check_mark: | Configuration of Mattermost [(details)](https://github.com/rodcorsi/mattermail#mattermost)                |
| MailTemplate      | string  |         |                    | Template used to format message to post [(details)](https://github.com/rodcorsi/mattermail#mailtemplate)  |
| LinesToPreview    |   int   | 10      |                    | Number of email lines that will be posted                                                                 |
| Attachment        | boolean | true    |                    | Inform if attachments will be posted in Mattermost                                                        |
| Disabled          | boolean | false   |                    | Disable this profile                                                                                      |
| RedirectBySubject | boolean | true    |                    | Inform if redirect email by subject [(details)](https://github.com/rodcorsi/mattermail#redirectbysubject) |
| Filter            | object  |         |                    | Filter used to redirect email [(details)](https://github.com/rodcorsi/mattermail#filter)                  |
=======
| Field             | Type    | Default | Obrigatory         | Information                                                                                              |
|-------------------|:-------:|---------|:------------------:|----------------------------------------------------------------------------------------------------------|
| Name              | string  |         | :white_check_mark: | Name of profile, used to log                                                                             |
| Channels          | array   |         | :white_check_mark: | List of channels where the email will be posted. You can use `#channel` or `@username`                   |
| Email             | object  |         | :white_check_mark: | Configuration of Email [(details)](https://github.com/rodcorsi/mattermail#email)                         |
| Mattermost        | object  |         | :white_check_mark: | Configuration of Mattermost [(details)](https://github.com/rodcorsi/mattermail#mattermost)               |
| MailTemplate      | string  |         |                    | Template used to format message to post [(details)](https://github.com/rodcorsi/mattermail#mailtemplate) |
| LinesToPreview    | int     | 10      |                    | Number of email lines that will be posted                                                                |
| Attachment        | boolean | true    |                    | Inform if attachments will be posted in Mattermost                                                       |
| Disabled          | boolean | false   |                    | Disable this profile                                                                                     |
| RedirectBySubject | boolean | true    |                    | Inform if redirect email by subject [(details)](https://github.com/rodcorsi/mattermail#redirectbysubject)|
| Filter            | object  |         |                    | Filter used to redirect email [(details)](https://github.com/rodcorsi/mattermail#filter)                 |
>>>>>>> b766d77... fixed imports to original repo

#### Email

Email configuration, used to access IMAP server

| Field             |  Type   | Default |     Obrigatory     | Information                                                                |
| ----------------- | :-----: | ------- | :----------------: | -------------------------------------------------------------------------- |
| ImapServer        | string  |         | :white_check_mark: | Address of imap server with port number ex: _imap.example.com:143_         |
| Username          | string  |         | :white_check_mark: | Email address or username used authenticate on email server                |
| Password          | string  |         | :white_check_mark: | Password used authenticate on email server                                 |
| StartTLS          | boolean | false   |                    | Enable StartTLS connection if server supports                              |
| TLSAcceptAllCerts | boolean | false   |                    | Accept insecure certificates with TLS connection                           |
| DisableIdle       | boolean | false   |                    | Disable imap idle and check email after 1 minute. Used in case of problems |

#### Mattermost

Mattermost configuration

<<<<<<< HEAD
| Field    |  Type   | Default |     Obrigatory     | Information                                                                                                                |
| -------- | :-----: | ------- | :----------------: | -------------------------------------------------------------------------------------------------------------------------- |
| Server   | string  |         | :white_check_mark: | Address of mattermost server. Please inform protocol and port if its necessary ex: _<https://mattermost.example.com:8065>_ |
| Team     | string  |         | :white_check_mark: | Team name. You can find teams name by [(URL)](https://github.com/rodcorsi/mattermail#teamchannel)                          |
| User     | string  |         | :white_check_mark: | User used to authenticate on Mattermos server                                                                              |
| Password | string  |         | :white_check_mark: | Password used to authenticate on Mattermos server                                                                          |
| UseAPIv3 | boolean | false   |                    | Set to use Mattermost Api V3                                                                                               |
=======
| Field     | Type   | Default | Obrigatory         | Information                                                                                                                |
|-----------|:------:|---------|:------------------:|----------------------------------------------------------------------------------------------------------------------------|
| Server    | string |         | :white_check_mark: | Address of mattermost server. Please inform protocol and port if its necessary ex: _<https://mattermost.example.com:8065>_ |
| Team      | string |         | :white_check_mark: | Team name. You can find teams name by [(URL)](https://github.com/rodcorsi/mattermail#teamchannel)                          |
| User      | string |         | :white_check_mark: | User used to authenticate on Mattermos server                                                                              |
| Password  | string |         | :white_check_mark: | Password used to authenticate on Mattermos server                                                                          |
| UseAPIv3  | string | true    |                    | Set to use Mattermost Api V3                                                                                               |
>>>>>>> b766d77... fixed imports to original repo

#### MailTemplate

This configuration formats email message using markdown to post on Mattermost.
The default configuration is `:incoming_envelope: _From: **{{.From}}**_\n>_{{.Subject}}_\n\n{{.Message}}`, in this example when Mattermail receives a message from `john@example.com`, with subject `Hello world` and message body `Hi I'm John`. This email will be formated to:

:incoming\*envelope: \_From: **john@example.com\***

> _Hello world_

Hi I'm John

#### RedirectBySubject

If the option `RedirectBySubject` is `true` the Mattermail will try to redirect an email and post it using the subject, ex:

| Subject                   | Destination                      |
| ------------------------- | -------------------------------- |
| [#orders] blah            | channel `orders`                 |
| [#orders #info] blah      | channel `orders` and `info`      |
| Fwd [#orders][#info] blah | channel `orders` and `info`      |
| [1234#orders] foo         | channel `orders`                 |
| [@john] blah              | user `john`                      |
| [@john #orders] blah      | user `john` and channel `orders` |

#### Filter

This option is used to redirect email following the rules.

```javascript
"Filter":            [
    /* if subject contains 'Feature' redirect to #feature */
    {"Subject":"Feature", "Channels": ["#feature"]},

    /* if from contains 'test@gmail.com' and subject contains 'to me' redirect to @test2*/
    {"From":"test@gmail.com", "Subject":"To Me", "Channels": ["@test2"]},

    /* if from contains '@companyb.com' redirect to #companyb */
    {"From":"@companyb.com", "Channel":"#companyb"} /**/

    /* if email belongs to the specific folder 'somefolder' redirect to #somechannel
    {"Folder":"somefolder", "Channel":"#somechannel"} /**/

    /* if email belongs to the specific folder 'somefolder' and Subject contains 'Test' redirect to #somechannel
    {"Folder":"anotherfolder", "Subject": "Test" ,"Channel":"#somechannel"} /**/

    /* if email belongs to the specific folder 'somefolder' and from contains 'bla@blah.blah' redirect to #somechannel
    {"Folder":"anotherfolder", "From": "blah@blah.blah" ,"Channel":"#somechannel"} /**/

    /* if from contains '@companyb.com' redirect to #companyb and @john */
    {"From":"@companyb.com", "Channels": ["#companyb", "@john"]} /**/

]
```

#### Team/Channel

You can find team and channel name by URL ex:

![mattermail teamchannel](https://github.com/rodcorsi/mattermail/raw/master/img/team_channel.png)

## Sequence that the email will be redirected

Mattermail post the email using this rules:

1 - Try to post using the subject if the option `RedirectBySubject` is `true`

2 - Try to post following the [Filter](https://github.com/rodcorsi/mattermail#filter) configuration.

3 - Post on channels/users defined on field `Channels` in `config.json`

## Options

```bash
$ ./mattermail --help
Usage:
    mattermail server  Starts Mattermail server
    mattermail migrate Migrates config.json to new version

For more details execute:

    mattermail [command] --help
```

## Building

You need [Go](http://golang.org) to build this project

```bash
go get github.com/rodcorsi/mattermail
```

## Mattermail as a service

### Using systemd

Considering your installation under `/opt/mattermail`, add `/etc/systemd/system/mattermail.service` file with the following content:

```
# mattermail
[Unit]
Description=mattermail server

[Service]
Type=simple
WorkingDirectory=/opt/mattermail
ExecStart=/opt/mattermail/mattermail server -c config.json
Nice=5

[Install]
WantedBy=multi-user.target
```

Enable service:

`systemctl enable mattermail`

Start service:

`systemctl start mattermail`

View status:

`systemctl status mattermail`

View log:

`journalctl -f -u mattermail.service`
