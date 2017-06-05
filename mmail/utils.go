package mmail

import (
	"log"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/data" //initiate go-charset data
)

// NonASCII Decode non ASCII header string RFC 1342
func NonASCII(encoded string) string {

	regexRFC1342, _ := regexp.Compile(`=\?.*?\?=`)
	dec := new(mime.WordDecoder)
	dec.CharsetReader = charset.NewReader

	result := regexRFC1342.ReplaceAllStringFunc(encoded, func(encoded string) string {
		decoded, err := dec.Decode(encoded)
		if err != nil {
			log.Println("Error decode NonASCII", encoded, err)
			return encoded
		}
		return decoded
	})

	return result
}

var channelRegex = regexp.MustCompile(`\[[^\]]*?([#@][A-Za-z0-9.\-_]+)[^\]]*?\]`)

// getChannelFromSubject extract channel from subject ex:
// getChannelFromSubject([#mychannel] blablanla) => #mychannel
func getChannelFromSubject(subject string) string {
	ret := channelRegex.FindStringSubmatch(subject)
	if len(ret) < 2 {
		return ""
	}
	return strings.ToLower(ret[len(ret)-1])
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
