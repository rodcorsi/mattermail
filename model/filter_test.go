package model

import (
	"testing"
)

func TestRule_Validate(t *testing.T) {
	rule := &Rule{}
	valid := func() {
		if err := rule.Validate(); err == nil {
			t.Fatal("this rule need to be invalid")
		}
	}

	valid()

	rule.From = "test@test.com"
	valid()

	rule.Subject = "subject"
	valid()

	rule.Channels = []string{"#hey"}
	if err := rule.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRule_Match(t *testing.T) {
	rule := &Rule{}

	rule.Channels = []string{"#test"}
	rule.From = "test@test.com"

	if rule.Match("", "", "") {
		t.Fatal("Do not attempt from rule")
	}

	if !rule.Match("test@test.com", "", "") {
		t.Fatal("Attempt from rule")
	}

	if rule.Match("other@test.com", "", "") {
		t.Fatal("Do not attempt from rule")
	}

	if !rule.Match("test@test.com", "ansdkjfhad", "") {
		t.Fatal("Attempt from rule subject need to be ignored")
	}

	if !rule.Match("test@test.com", "", "asdf") {
		t.Fatal("Attempt from rule folder need to be ignored")
	}

	if rule.Match("", "ansdkjfhad", "") {
		t.Fatal("Do not attempt from rule subject need to be ignored")
	}

	if rule.Match("", "", "asdf") {
		t.Fatal("Do not attempt from rule folder need to be ignored")
	}

	rule.Subject = "subject"
	rule.From = ""
	rule.Folder = ""

	if rule.Match("", "", "") {
		t.Fatal("Do not attempt subject rule")
	}

	if rule.Match("test@test.com", "", "") {
		t.Fatal("Do not attempt subject rule, from need to be ignored")
	}

	if !rule.Match("test@test.com", "ansdf subject dfad", "") {
		t.Fatal("Attempt subject rule from need to be ignored")
	}

	if rule.Match("", "ansdkjfhad", "") {
		t.Fatal("Do not attempt subject rule")
	}

	rule.Subject = "subject"
	rule.From = "test@test.com"
	rule.Folder = "test"

	if rule.Match("", "", "") {
		t.Fatal("Do not attempt rules")
	}

	if rule.Match("dfadf", "sdfsafa", "") {
		t.Fatal("Do not attempt rules")
	}

	if rule.Match("dfadf", "sdfsafa", "kljskdf") {
		t.Fatal("Do not attempt rules")
	}

	if rule.Match("test@test.com", "", "") {
		t.Fatal("Do not attempt all rules")
	}

	if rule.Match("", "subject", "") {
		t.Fatal("Do not attempt rules")
	}

	if rule.Match("", "", "test") {
		t.Fatal("Do not attempt rules")
	}

	if !rule.Match("test@test.com", "asdH subject assdhj", "test") {
		t.Fatal("Attempt all rules")
	}
}

func TestFilter_Fix(t *testing.T) {
	filter := &Filter{
		&Rule{
			Channels: []string{"  Channel "},
			Subject:  " Subject  ",
			From:     " Test@test.com ",
		},
	}

	filter.Fix()

	rule := (*filter)[0]

	if rule.Channels[0] != "#channel" {
		t.Fatal("Expected Channel: #channel result:", rule.Channels)
	}

	if rule.Subject != "subject" {
		t.Fatal("Expected Subject: subject result:", rule.Subject)
	}

	if rule.From != "test@test.com" {
		t.Fatal("Expected From: test@test.com result:", rule.From)
	}
}

func TestFilter_Validate(t *testing.T) {
	filter := &Filter{}

	if err := filter.Validate(); err == nil {
		t.Fatal("Expected error not nil for empty filter")
	}

	*filter = append(*filter, &Rule{})

	if err := filter.Validate(); err == nil {
		t.Fatal("Expected error not nil for filter with invalid rule")
	}

	*filter = append(Filter{}, &Rule{Subject: "X", Channels: []string{"#channel"}})
	if err := filter.Validate(); err != nil {
		t.Fatal("Expected error nil for valid filter err:", err.Error())
	}
}
