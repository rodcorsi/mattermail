package mmail

import "testing"

func TestGetChannelFromSubject(t *testing.T) {
	assert := func(subject, expected string) {
		r := getChannelFromSubject(subject)
		if r != expected {
			t.Fatalf("Tested:%v expected:%v find:%v", subject, expected, r)
		}
	}

	assert("[#test]", "#test")
	assert("[#TeSt]", "#test")
	assert(" [#test]", "#test")
	assert("   [#test]", "#test")
	assert("[ #test]", "#test")
	assert("[   #test]", "#test")
	assert("[#test ]", "#test")
	assert("[#test    ]", "#test")
	assert("[#test] kjshdfsdh [#kjshdf]", "#test")
	assert("   [  #  test   ]  kjshdfsdh [#kjshdf]", "")
	assert("   [  #t-e_st   ]  kjshdfsdh [#kjshdf]", "#t-e_st")
	assert("[#test fsd   ]", "")

	assert("[@test]", "@test")
	assert("[@TeSt]", "@test")
	assert(" [@test]", "@test")
	assert("   [@test]", "@test")
	assert("[ @test]", "@test")
	assert("[   @test]", "@test")
	assert("[@test ]", "@test")
	assert("[@test    ]", "@test")
	assert("[@test] kjshdfsdh [#kjshdf]", "@test")
	assert("   [  @  test   ]  kjshdfsdh [@kjshdf]", "")
	assert("   [  @t-e_st   ]  kjshdfsdh [@kjshdf]", "@t-e_st")
	assert("[@test fsd   ]", "")
}

func TestReadLines(t *testing.T) {
	testCount := 0
	assert := func(lines string, nmax int, expected string) {
		testCount++
		r := readLines(lines, nmax)
		if r != expected {
			t.Fatalf("Error readLines test:%v\nexpected:\n%q\n\nreturned:\n%q", testCount, expected, r)
		}
	}

	var lines string

	lines = ""

	assert(lines, 10, "")
	assert(lines, 0, "")

	lines = "AAA\nBBB\nCCC"

	assert(lines, 10, lines)
	assert(lines, 3, lines)
	assert(lines, 0, "")
	assert(lines, 2, "AAA\nBBB")

	lines = "AAA\n"
	assert(lines, 1, lines)
	assert(lines, 2, lines)

	lines = "\n"
	assert(lines, 1, lines)
	assert(lines, 2, lines)
	assert(lines, 10, lines)
	assert(lines, 0, "")

	lines = "\r\n"
	assert(lines, 1, lines)
	assert(lines, 2, lines)
	assert(lines, 10, lines)
	assert(lines, 0, "")

	lines = "AAA\r\nBBB\nCCC"
	assert(lines, 1, "AAA")
	assert(lines, 2, lines)

	lines = "AAA\r\n"
	assert(lines, 1, lines)
	assert(lines, 2, lines)
}
