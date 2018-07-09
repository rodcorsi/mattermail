package cmd

// Version show the current version, changed during the make build
var Version = "4.0-dev"

type command interface {
	execute() error
	parse(arguments []string) error
}

// Execute parses command line and execute it
func Execute(args []string) error {
	cmd, err := parseCommand(args)
	if err != nil {
		return err
	}

	return cmd.execute()
}

func parseCommand(args []string) (cmd command, err error) {
	if len(args) <= 1 {
		cmd = &serverCommand{}
		cmd.parse([]string{})
	} else {
		switch args[1] {
		case "server":
			cmd = &serverCommand{}
			err = cmd.parse(args[2:])
		case "migrate":
			cmd = &migrateCommand{}
			err = cmd.parse(args[2:])
		case "-h", "--help":
			cmd = &stringCommand{usage}
		case "-v", "--version":
			cmd = &stringCommand{version}
		default:
			cmd = &serverCommand{}
			err = cmd.parse(args[1:])
		}
	}
	return
}

var usage = `MatterMail is an integration service for Mattermost, MatterMail listen
an email box and publish all received emails in a channel, private group
or user in Mattermost

Version: ` + Version + `

Usage:
	mattermail server  Starts Mattermail server
	mattermail migrate Migrates config.json to new version

For more details execute:

	mattermail [command] --help

Options:
    -h, --help  Show this help

`

var version = "Version: " + Version + "\n"
