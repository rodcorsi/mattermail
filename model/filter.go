package model

import (
	"errors"
	"strings"
)

// Rule for filter
type Rule struct {
	From     string
	Subject  string
	Channels []string
  Folder  string

}

// Filter has an array of rules
type Filter []*Rule

// Fix remove spaces and convert to lower case the rules
func (r *Rule) Fix() {
	r.From = strings.TrimSpace(strings.ToLower(r.From))
	r.Subject = strings.TrimSpace(strings.ToLower(r.Subject))

	for i, channel := range r.Channels {
		channel = strings.TrimSpace(channel)
		channel = strings.ToLower(channel)

		if !strings.HasPrefix(channel, "#") && !strings.HasPrefix(channel, "@") {
			channel = "#" + channel
		}
		r.Channels[i] = channel
	}
}

// Validate check if this rule is valid
func (r *Rule) Validate() error {
	if len(r.From) == 0 && len(r.Subject) == 0 && len(r.Folder) == 0 {
		return errors.New("Need to set From, Subject or Folder")
	}

	if len(r.Channels) == 0 {
		return errors.New("Need to set at least one channel or user for destination")
	}

	for _, channel := range r.Channels {
		if channel != "" && !validateChannel(channel) {
			return errors.New("Need to set #channel or @user")
		}
	}

	return nil
}

func (r *Rule) hasNonEmptyFolder() bool {
	if len(r.Folder) != 0 {
		return true
	}
	return false
}

func (r *Rule) getFolder() string {
	return r.Folder
}

func (r *Rule) matchFrom(from string) bool {
	from = strings.ToLower(from)
	if len(r.From) == 0 {
		return true
	}
	return strings.Contains(from, r.From)
}

func (r *Rule) matchSubject(subject string) bool {
	subject = strings.ToLower(subject)
	if len(r.Subject) == 0 {
		return true
	}
	return strings.Contains(subject, r.Subject)
}

func (r *Rule) matchFolder(folder string) bool {
	if len(r.Folder) == 0 {
		return true
	}
	return strings.Contains(folder, r.Folder)
}

// Match check if from, subject and folder meets this rule
func (r *Rule) Match(from, subject, folder string) bool {
	return r.matchFrom(from) && r.matchSubject(subject) && r.matchFolder(folder)
}


// GetChannels return the first channels with attempt the rules
func (f *Filter) GetChannels(from, subject string) []string {
	for _, r := range *f {
		if r.Match(from, subject) {
			return r.Channels
		}
	}
	return []string{""}
}

// ListFolder return all folders defined in filter rules
func (f *Filter) ListFolder() []string {
	var list []string
	for _, r := range *f {
		if r.hasNonEmptyFolder() {
			list = append(list, r.getFolder())
		}
	}
	return dedupStrings(list)
}

// Validate check if all rules is valid
func (f *Filter) Validate() error {
	if len(*f) == 0 {
		return errors.New("Filter need to be at least one rule to be valid")
	}

	for _, r := range *f {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Fix all rules
func (f *Filter) Fix() {
	for _, r := range *f {
		r.Fix()
	}
}
