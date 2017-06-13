package model

import "regexp"

func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9\._%+\-]+@[a-z0-9\.\-_]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

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
