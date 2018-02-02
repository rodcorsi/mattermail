package model

import (
	"os"
	"path/filepath"
	"regexp"
)

func validateURL(url string) bool {
	Re := regexp.MustCompile(`^https?://[a-z0-9\.\-_]+:?([0-9]{1,5})?$`)
	return Re.MatchString(url)
}

func validateImap(url string) bool {
	Re := regexp.MustCompile(`^[a-z0-9\.\-_]+:?([0-9]{1,5})?$`)
	return Re.MatchString(url)
}

func validateChannel(channel string) bool {
	Re := regexp.MustCompile(`^(#|@)[a-z0-9\.\-_]+$`)
	return Re.MatchString(channel)
}

func validateTeam(team string) bool {
	Re := regexp.MustCompile(`^[a-z0-9\.\-_]+$`)
	return Re.MatchString(team)
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

func dedupStrings(elements []string) []string {
	encountered := map[string]bool{}

	for v := range elements {
		encountered[elements[v]] = true
	}

	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}
