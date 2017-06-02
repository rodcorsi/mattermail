package mmail

import "testing"

func TestGetChannelFromSubject(t *testing.T) {
	assert := func(subject, expected string) {
		r := getChannelFromSubject(subject)
		if r != expected {
			t.Fatalf("Tested:%v expected:%v find:%v", subject, expected, r)
		}
	}

	assert("", "")
	assert("[#test]", "#test")
	assert("[#test.1]", "#test.1")
	assert("[@TeSt]", "@test")
	assert("[@Test.1]", "@test.1")
	assert("[@TeSt.1]", "@test.1")
	assert(" [#test]", "#test")
	assert("   [@test]", "@test")
	assert("[ #test]", "#test")
	assert("[   @test]", "@test")
	assert("[#test ]", "#test")
	assert("[@test    ]", "@test")
	assert("[@test] kjshdfsdh [#kjshdf]", "@test")
	assert("   [  #  sadfkj   ]  kjshdfsdh [#test]", "")
	assert("   [  @t-e_st   ]  kjshdfsdh [#kjshdf]", "@t-e_st")
	assert("[#test fsd   ]", "")
	assert("From:[@test]", "@test")
	assert("fwd:  [#test]", "#test")
	assert("foo baz  [@test]", "")
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

func TestNonASCII(t *testing.T) {
	assert := func(test, expected string) {
		result := NonASCII(test)
		if result != expected {
			t.Fatalf("Error NonASCII test:%v\nexpected:\n%q\n\nresult:\n%q", test, expected, result)
		}
	}

	assert("", "")
	assert("test", "test")
	assert("Fwd: =?UTF-8?B?Q290YcOnw6NvIGRlIEZlcnJhbWVudGFz?=", "Fwd: Cotação de Ferramentas")
	assert("=?utf-8?B?2KfbjNmGINuM2qkg2YXYqtmGINiz2KfYr9mHINin2LPYqg==?= =?utf-8?B?2KfbjNmGINuM2qkg2YXYqtmGINiz2KfYr9mHINin2LPYqg==?= =?utf-8?B?2YbYr9is?=", "این یک متن ساده است این یک متن ساده است ندج")
	assert("From: =?US-ASCII?Q?Keith_Moore?= <moore@cs.utk.edu>", "From: Keith Moore <moore@cs.utk.edu>")
	assert("To: =?ISO-8859-1?Q?Keld_J=F8rn_Simonsen?= <keld@dkuug.dk>", "To: Keld Jørn Simonsen <keld@dkuug.dk>")
	assert("CC: =?ISO-8859-1?Q?Andr=E9_?= Pirard <PIRARD@vm1.ulg.ac.be>", "CC: André  Pirard <PIRARD@vm1.ulg.ac.be>")
	assert("Subject: =?ISO-8859-1?B?SWYgeW91IGNhbiByZWFkIHRoaXMgeW8=?=", "Subject: If you can read this yo")
	assert("=?ISO-8859-2?B?dSB1bmRlcnN0YW5kIHRoZSBleGFtcGxlLg==?=", "u understand the example.")
	assert("From: =?ISO-8859-1?Q?Olle_J=E4rnefors?= <ojarnef@admin.kth.se>", "From: Olle Järnefors <ojarnef@admin.kth.se>")
	assert("From: =?ISO-8859-1?Q?Patrik_F=E4ltstr=F6m?= <paf@nada.kth.se>", "From: Patrik Fältström <paf@nada.kth.se>")
}
