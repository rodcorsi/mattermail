package model

import (
	"errors"
	"strings"
)

// Rule for filter
type Rule struct {
	From    string
	Subject string
	Channel string
	Folder  string
}

// Filter has an array of rules
type Filter []*Rule

// Fix remove spaces and convert to lower case the rules
func (r *Rule) Fix() {
	r.From = strings.TrimSpace(strings.ToLower(r.From))
	r.Subject = strings.TrimSpace(strings.ToLower(r.Subject))
	r.Channel = strings.TrimSpace(strings.ToLower(r.Channel))

	if !strings.HasPrefix(r.Channel, "#") && !strings.HasPrefix(r.Channel, "@") {
		r.Channel = "#" + r.Channel
	}
}

// Validate check if this rule is valid
func (r *Rule) Validate() error {
	if len(r.From) == 0 && len(r.Subject) == 0 && len(r.Folder) == 0 {
		return errors.New("Need to set From, Subject or Folder")
	}

	if len(r.Channel) == 0 {
		return errors.New("Need to set a Channel")
	}

	if !strings.HasPrefix(r.Channel, "#") && !strings.HasPrefix(r.Channel, "@") {
		return errors.New("Need to set a #channel or @user")
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

// GetChannel return the first channel with attempt the rules
func (f *Filter) GetChannel(from, subject, folder string) string {
	for _, r := range *f {
		if r.Match(from, subject, folder) {
			return r.Channel
		}
	}
	return ""
}

// list folder to select MailBox
func (f *Filter) ListFolder() []string {
	var list []string
	for _, r := range *f {
		if r.hasNonEmptyFolder() {
			list = append(list, r.getFolder())
		}
	}
	return list
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
