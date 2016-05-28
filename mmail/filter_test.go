package mmail

import "testing"

func TestRuleIsValid(t *testing.T) {
	rule := &Rule{}
	valid := func() {
		if err := rule.IsValid(); err == nil {
			t.Fatal("this rule need to be invalid")
		}
	}

	valid()

	rule.From = "test@test.com"
	valid()

	rule.Subject = "subject"
	valid()

	rule.Channel = "#hey"
	if err := rule.IsValid(); err != nil {
		t.Fatal(err)
	}
}

func TestRuleMeetsRule(t *testing.T) {
	rule := &Rule{}

	rule.Channel = "#test"
	rule.From = "test@test.com"

	if rule.MeetsRule("", "") {
		t.Fatal("Do not attempt from rule")
	}

	if !rule.MeetsRule("test@test.com", "") {
		t.Fatal("Attempt from rule")
	}

	if rule.MeetsRule("other@test.com", "") {
		t.Fatal("Do not attempt from rule")
	}

	if !rule.MeetsRule("test@test.com", "ansdkjfhad") {
		t.Fatal("Attempt from rule subject need to be ignored")
	}

	if rule.MeetsRule("", "ansdkjfhad") {
		t.Fatal("Do not attempt from rule subject need to be ignored")
	}

	rule.Subject = "subject"
	rule.From = ""

	if rule.MeetsRule("", "") {
		t.Fatal("Do not attempt subject rule")
	}

	if rule.MeetsRule("test@test.com", "") {
		t.Fatal("Do not attempt subject rule, from need to be ignored")
	}

	if !rule.MeetsRule("test@test.com", "ansdf subject dfad") {
		t.Fatal("Attempt subject rule from need to be ignored")
	}

	if rule.MeetsRule("", "ansdkjfhad") {
		t.Fatal("Do not attempt subject rule")
	}

	rule.Subject = "subject"
	rule.From = "test@test.com"

	if rule.MeetsRule("", "") {
		t.Fatal("Do not attempt rules")
	}

	if rule.MeetsRule("dfadf", "sdfsafa") {
		t.Fatal("Do not attempt rules")
	}

	if rule.MeetsRule("test@test.com", "") {
		t.Fatal("Do not attempt all rules")
	}

	if rule.MeetsRule("", "subject") {
		t.Fatal("Do not attempt rules")
	}

	if !rule.MeetsRule("test@test.com", "asdH subject assdhj") {
		t.Fatal("Attempt all rules")
	}
}
