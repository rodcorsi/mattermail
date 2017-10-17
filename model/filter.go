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
	if len(r.From) == 0 && len(r.Subject) == 0 {
		return errors.New("Need to set From or Subject")
	}

	if len(r.Channel) == 0 {
		return errors.New("Need to set a Channel")
	}

	if !strings.HasPrefix(r.Channel, "#") && !strings.HasPrefix(r.Channel, "@") {
		return errors.New("Need to set a #channel or @user")
	}
	return nil
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

// Match check if from and subject meets this rule
func (r *Rule) Match(from, subject string) bool {
	return r.matchFrom(from) && r.matchSubject(subject)
}

// GetChannel return the first channel with attempt the rules
func (f *Filter) GetChannel(from, subject string) string {
	for _, r := range *f {
		if r.Match(from, subject) {
			return r.Channel
		}
	}
	return ""
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
