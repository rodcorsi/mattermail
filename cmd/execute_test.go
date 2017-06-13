package cmd

import "testing"

func TestParseCommand(t *testing.T) {
	assertServer := func(n int, args []string) {
		cmd, err := parseCommand(args)
		if _, ok := cmd.(*serverCommand); !ok {
			t.Fatalf("Test %v Expected serverCommand result:%T args:%v", n, cmd, args)
		}

		if err != nil {
			t.Fatalf("Test %v Error on parse a valid command args:%v error:%v", n, args, err.Error())
		}
	}

	assertServer(0, []string{"mattermail"})
	assertServer(1, []string{"mattermail", "-c", "./config.json"})
	assertServer(2, []string{"mattermail", "server"})
	assertServer(3, []string{"mattermail", "server", "nothing123"})

	assertMigrate := func(n int, args []string) {
		cmd, err := parseCommand(args)
		if _, ok := cmd.(*migrateCommand); !ok {
			t.Fatalf("Test %v Expected migrateCommand result:%T args:%v", n, cmd, args)
		}

		if err != nil {
			t.Fatalf("Test %v Error on parse a valid command args:%v error:%v", n, args, err.Error())
		}
	}
	assertMigrate(4, []string{"mattermail", "migrate"})
	assertMigrate(5, []string{"mattermail", "migrate", "-c", "./config.json"})

	assertString := func(n int, args []string) {
		cmd, err := parseCommand(args)
		if _, ok := cmd.(*stringCommand); !ok {
			t.Fatalf("Test %v Expected stringCommand result:%T args:%v", n, cmd, args)
		}

		if err != nil {
			t.Fatalf("Test %v Error on parse a valid command args:%v error:%v", n, args, err.Error())
		}
	}

	assertString(6, []string{"mattermail", "--help"})
	assertString(7, []string{"mattermail", "-h"})
	assertString(8, []string{"mattermail", "--version"})
	assertString(9, []string{"mattermail", "-v"})
}
