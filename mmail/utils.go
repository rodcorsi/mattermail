package mmail

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/paulrosania/go-charset/data" //initiate go-charset data
)

var channelRegex = regexp.MustCompile(`[#@][A-Za-z0-9.\-_]+`)
var bracketsRegex = regexp.MustCompile(`\[[^\]]*\]`)

// getChannelsFromSubject extract channel from subject ex:
// getChannelsFromSubject([#mychannel] blablanla) => #mychannel
func getChannelsFromSubject(subject string) []string {
	ret := bracketsRegex.FindAllString(subject, -1)

	if len(ret) == 0 {
		return nil
	}

	var channels []string

	for _, e := range ret {
		chs := channelRegex.FindAllString(e, -1)
		for _, c := range chs {
			channels = append(channels, strings.ToLower(c))
		}
	}

	if len(channels) == 0 {
		return nil
	}

	return channels
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

func findDir(dir string) string {
	fileName := "."
	if _, err := os.Stat("./" + dir + "/"); err == nil {
		fileName, _ = filepath.Abs("./" + dir + "/")
	} else if _, err := os.Stat("../" + dir + "/"); err == nil {
		fileName, _ = filepath.Abs("../" + dir + "/")
	} else if _, err := os.Stat("/tmp/" + dir); err == nil {
		fileName, _ = filepath.Abs("/tmp/" + dir)
	}

	return fileName + "/"
}
