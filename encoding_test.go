package mimeparser

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestDecodeHeader(t *testing.T) {
	testData := []struct{ raw, expected string }{
		{
			"",
			"",
		},
		{
			"=?UTF-8?B?w4TDi8OPw5bDnA==?= =?UTF-8?B?IMOkw6vDr8O2w7w=?=",
			"ÄËÏÖÜ äëïöü",
		},
		{
			"=?ISO-8859-2?B?xMtJ1tw=?= =?ISO-8859-2?B?IOTrafb8?=",
			"ÄËIÖÜ äëiöü",
		},
		{
			"=?uknown?B?xMtJ1tw=?= =?ISO-8859-2?B?IOTrafb8?=",
			"=?uknown?B?xMtJ1tw=?= =?ISO-8859-2?B?IOTrafb8?=",
		},
	}

	for _, val := range testData {
		if decoded, err := DecodeHeader(val.raw); strings.Compare(val.expected, decoded) != 0 {
			t.Error("Incorrect decoding of header", val.raw, "expected", val.expected, "but have", decoded, ". Error", err)
		} else {
			fmt.Println("Header", val.raw, "successfully decoded", decoded, ". Error", err)
		}
	}
}

// sample text for UTF8 http://www.columbia.edu/~fdc/utf8/index.html
func TestEncodeReader(t *testing.T) {
	// define test data
	testData := []struct {
		params   map[string]string
		original []byte
		message  string
	}{
		// russian
		{
			map[string]string{"charset": "koi8-r"},
			//     а, з, б, у, к, а, а, б, в, г, д, е, ё
			[]byte{0xC1, 0xDA, 0xC2, 0xD5, 0xCB, 0xC1, 0xC1, 0xC2, 0xD7, 0xC7, 0xC4, 0xC5, 0xA3},
			"азбукаабвгдеё",
		},
		{
			map[string]string{"charset": "KOI8-R"},
			[]byte{0xC1, 0xDA, 0xC2, 0xD5, 0xCB, 0xC1, 0xC1, 0xC2, 0xD7, 0xC7, 0xC4, 0xC5, 0xA3},
			"азбукаабвгдеё",
		},
		{
			map[string]string{"charset": "csKOI8R"},
			[]byte{0xC1, 0xDA, 0xC2, 0xD5, 0xCB, 0xC1, 0xC1, 0xC2, 0xD7, 0xC7, 0xC4, 0xC5, 0xA3},
			"азбукаабвгдеё",
		},
		{
			map[string]string{"charset": "koi8-u"},
			[]byte{0xC1, 0xDA, 0xC2, 0xD5, 0xCB, 0xC1, 0xC1, 0xC2, 0xD7, 0xC7, 0xC4, 0xC5, 0xA3},
			"азбукаабвгдеё",
		},
		{
			map[string]string{"charset": "iso-8859-5"},
			//     а    , з    , б    , у    , к    , а    , а    , б    , в    , г    , д    , е    , ё
			[]byte{0xD0, 0xD7, 0xD1, 0xE3, 0xDA, 0xD0, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xF1},
			"азбукаабвгдеё",
		},
		{
			map[string]string{"charset": "csWrong"},
			[]byte{0xD0, 0xD7, 0xD1, 0xE3, 0xDA, 0xD0, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6},
			"",
		},
		{
			map[string]string{"charset": "utf8"},
			[]byte{0xD0, 0xB0, 0xD0, 0xB7, 0xD0, 0xB1, 0xD1, 0x83, 0xD0, 0xBA, 0xD0, 0xB0, 0xD0, 0xB0, 0xD0, 0xB1, 0xD0, 0xB2, 0xD0, 0xB3, 0xD0, 0xB4, 0xD0, 0xB5, 0xD1, 0x91},
			"азбукаабвгдеё",
		},
		// czechoslovakia
		{
			map[string]string{"charset": "windows-1250"},
			[]byte{225, 228, 232, 233, 236, 244},
			"áäčéěô",
		},
		// umlauts
		{
			map[string]string{"charset": "iso-8859-1"},
			[]byte{196, 203, 214, 220, 228, 235, 246, 252},
			"ÄËÖÜäëöü",
		},
		// latvia
		{
			map[string]string{"charset": "iso-8859-4"},
			[]byte{224, 239, 243, 182, 254},
			"āīķļū",
		},
		{ // encoded by https://www.motobit.com/util/charset-codepage-conversion.asp
			map[string]string{"charset": "utf7"},
			[]byte("He wes Leovena+APA-es sone -- li+APA-e him be Drihten.+A6QDtw- +A7MDuwPOA8MDwwOx- +A7wDvwPF- +A60DtAPJA8MDsQO9- +A7UDuwO7A7cDvQO5A7oDrg-. +BCcENQRABD0ENQQ7BDg- +BDgENwQxBEs- +BDcENAQ1BEEETA- +BDg- +BEIEMAQ8-,+BCcENQRABD0ENQQ7BDg- +BDgENwQxBEs- +BDcENAQ1BEEETA- +BDg- +BEIEMAQ8-,+C68LvguuC7ELvwuoC80LpA- +C64Lygu0C78LlQuzC78LsgvH- +C6QLrgu/C7QLzQuuC8oLtAu/- +C6oLywuyC80- +C4cLqQu/C6QLvgu1C6QLwQ- +C44LmQvNC5ULwQuuC80- +C5ULvgujC8sLrgvN-."),
			"He wes Leovenaðes sone -- liðe him be Drihten.Τη γλώσσα μου έδωσαν ελληνική. Чернели избы здесь и там,Чернели избы здесь и там,யாமறிந்த மொழிகளிலே தமிழ்மொழி போல் இனிதாவது எங்கும் காணோம்.",
		},
	}

	// run tests
	for _, val := range testData {
		fmt.Println("Testing ", val)
		expected := []byte(val.message)
		decoded, err := DecodeCharset(val.original, val.params)
		if len(expected) == 0 {
			if err == nil {
				t.Error("Expected err but have ", err)
			} else {
				fmt.Println("Expected err: ", err)
				continue
			}
		} else {
			if err != nil {
				t.Error("Expected ok but have ", err)
			}
		}

		if bytes.Equal(decoded, expected) {
			fmt.Println("Succesfull decoding of ", val.params, ":", string(decoded))
		} else {
			t.Error("Wrong encoding of ", val.params, ".Expected ", expected, " but have ", decoded)
		}
		if strings.Compare(val.message, string(decoded)) != 0 {
			t.Error("Wrong message for ", val.params, ".Expected ", val.message, " but have ", string(decoded))
		}
	}
}
