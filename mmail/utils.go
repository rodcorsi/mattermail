package mmail

import (
	"encoding/base64"
	"log"
	"mime"
	"regexp"
	"strings"

	"github.com/jhillyerd/go.enmime"
	"github.com/paulrosania/go-charset/charset"
	_ "github.com/paulrosania/go-charset/data" //initiate go-charset data
)

//Replace cid:**** by embedded base64 image
func replaceCID(html *string, part *enmime.MIMEPart) string {
	cid := strings.Replace((*part).Header().Get("Content-ID"), "<", "", -1)
	cid = strings.Replace(cid, ">", "", -1)

	if len(cid) == 0 {
		return *html
	}

	b64 := "data:" + (*part).ContentType() + ";base64," + base64.StdEncoding.EncodeToString((*part).Content())

	return strings.Replace(*html, "cid:"+cid, b64, -1)
}

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

var channelRegex = regexp.MustCompile(`^([a-zA-Z]*:)?\s*?\[\s*?([#@][A-Za-z0-9.\-_]*)\s*?\]`)

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
