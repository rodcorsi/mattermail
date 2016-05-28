package mmail

import (
	"fmt"
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
}

// IsValid check if this rule is valid
func (r *Rule) IsValid() error {
	if len(r.From) == 0 && len(r.Subject) == 0 {
		return fmt.Errorf("Need to set From or Subject")
	}

	if len(r.Channel) == 0 {
		return fmt.Errorf("Need to set a Channel")
	}

	if !strings.HasPrefix(r.Channel, "#") && !strings.HasPrefix(r.Channel, "@") {
		return fmt.Errorf("Need to set a #channel or @user")
	}
	return nil
}

func (r *Rule) meetsFrom(from string) bool {
	from = strings.ToLower(from)
	if len(r.From) == 0 {
		return true
	}
	return strings.Contains(from, r.From)
}

func (r *Rule) meetsSubject(subject string) bool {
	subject = strings.ToLower(subject)
	if len(r.Subject) == 0 {
		return true
	}
	return strings.Contains(subject, r.Subject)
}

// MeetsRule check if from and subject meets this rule
func (r *Rule) MeetsRule(from, subject string) bool {
	return r.meetsFrom(from) && r.meetsSubject(subject)
}

// GetChannel return the first channel with attempt the rules
func (f *Filter) GetChannel(from, subject string) string {
	for _, r := range *f {
		if r.MeetsRule(from, subject) {
			return r.Channel
		}
	}
	return ""
}

// Valid check if all rules is valid and fix
func (f *Filter) Valid() error {
	for _, r := range *f {
		r.Fix()
		if err := r.IsValid(); err != nil {
			return err
		}
	}
	return nil
}
