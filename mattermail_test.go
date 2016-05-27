package main

import "testing"

func TestGetChannelFromSubject(t *testing.T) {
	assert := func(subject, expected string) {
		r := getChannelFromSubject(subject)
		if r != expected {
			t.Fatalf("Tested:%v expected:%v find:%v", subject, expected, r)
		}
	}

	assert("[#test]", "test")
	assert("[#TeSt]", "test")
	assert(" [#test]", "test")
	assert("   [#test]", "test")
	assert("[ #test]", "test")
	assert("[   #test]", "test")
	assert("[#test ]", "test")
	assert("[#test    ]", "test")
	assert("[#test] kjshdfsdh [#kjshdf]", "test")
	assert("   [  #  test   ]  kjshdfsdh [#kjshdf]", "test")
	assert("   [  #  t-e_st   ]  kjshdfsdh [#kjshdf]", "t-e_st")
	assert("[#test fsd   ]", "")
}

func TestGetUserFromSubject(t *testing.T) {
	assert := func(subject, expected string) {
		r := getUserFromSubject(subject)
		if r != expected {
			t.Fatalf("Tested:%v expected:%v find:%v", subject, expected, r)
		}
	}

	assert("[@test]", "test")
	assert("[@TeSt]", "test")
	assert(" [@test]", "test")
	assert("   [@test]", "test")
	assert("[ @test]", "test")
	assert("[   @test]", "test")
	assert("[@test ]", "test")
	assert("[@test    ]", "test")
	assert("[@test] kjshdfsdh [#kjshdf]", "test")
	assert("   [  @  test   ]  kjshdfsdh [@kjshdf]", "test")
	assert("   [  @  t-e_st   ]  kjshdfsdh [@kjshdf]", "t-e_st")
	assert("[@test fsd   ]", "")
}
