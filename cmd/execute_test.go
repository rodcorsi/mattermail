package cmd

import "testing"

func TestParseCommand(t *testing.T) {
	assertServer := func(args []string) {
		cmd, err := parseCommand(args)
		if _, ok := cmd.(*serverCommand); !ok {
			t.Fatalf("Expected serverCommand result:%T args:%v", cmd, args)
		}

		if err != nil {
			t.Fatalf("Error on parse a valid command args:%v error:%v", args, err.Error())
		}
	}

	assertServer([]string{"mattermail"})
	assertServer([]string{"mattermail", "-c", "./config.json"})
	assertServer([]string{"mattermail", "server"})
	assertServer([]string{"mattermail", "server", "nothing123"})

	assertString := func(args []string) {
		cmd, err := parseCommand(args)
		if _, ok := cmd.(*stringCommand); !ok {
			t.Fatalf("Expected stringCommand result:%T args:%v", cmd, args)
		}

		if err != nil {
			t.Fatalf("Error on parse a valid command args:%v error:%v", args, err.Error())
		}
	}

	assertString([]string{"mattermail", "--help"})
	assertString([]string{"mattermail", "-h"})
	assertString([]string{"mattermail", "--version"})
	assertString([]string{"mattermail", "-v"})
}
