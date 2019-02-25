package pmmime

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/text/encoding/htmlindex"
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

func TestGetEncoding(t *testing.T) {
	// all MIME charset with aliases can be found here https://www.iana.org/assignments/character-sets/character-sets.xhtml
	mimesets := map[string][]string{
		"utf-8": []string{ // MIB 16
			"utf8",
			"csutf8",
			"us-ascii", // MIB 3
			"iso-ir-6",
			"unicode-1-1-utf-8",
			"iso-utf-8",
			"utf8mb4",
			"ansi_x3.4-1968",
			"ansi_x3.4-1986",
			"iso_646.irv:1991",
			"iso646-us",
			"us",
			"ibm367",
			"cp367",
			"csascii",
			"ascii",
		},
		"gbk": []string{
			"gb2312", // MIB 2025
		},
		//"utf7": []string{"utf-7", "unicode-1-1-utf-7"},
		"iso-8859-1": []string{ // MIB 4
			"iso8859-1",
			"so-ir-100",
			"iso_8859-1",
			"latin1",
			"l1",
			"ibm819",
			"cp819",
			"csisolatin1",
		},
		"iso-8859-2": []string{ // MIB 5
			"iso-ir-101",
			"iso_8859-2",
			"iso8859-2",
			"latin2",
			"l2",
			"csisolatin2",
			"ibm852",
		},
		"iso-8859-3": []string{ // MIB 6
			"iso-ir-109",
			"iso_8859-3",
			"latin3",
			"l3",
			"csisolatin3",
		},
		"iso-8859-4": []string{ // MIB 7
			"iso-ir-110",
			"iso_8859-4",
			"latin4",
			"l4",
			"csisolatin4",
		},
		"iso-8859-5": []string{ // MIB 8
			"iso-ir-144",
			"iso_8859-5",
			"cyrillic",
			"csisolatincyrillic",
		},
		"iso-8859-6": []string{ // MIB 9
			"iso-ir-127",
			"iso_8859-6",
			"ecma-114",
			"asmo-708",
			"arabic",
			"csisolatinarabic"},
		"iso-8859-6e": []string{ // MIB 81
			"csiso88596e",
			"iso-8859-6-e"},
		"iso-8859-6i": []string{ // MIB 82
			"csiso88596i",
			"iso-8859-6-i"},
		"iso-8859-7": []string{ // MIB 10
			"iso-ir-126",
			"iso_8859-7",
			"elot_928",
			"ecma-118",
			"greek",
			"greek8",
			"csisolatingreek"},
		"iso-8859-8": []string{ // MIB 11
			"iso-ir-138",
			"iso_8859-8",
			"iso-8859-8-i", // Hebrew, the "i" means right-to-left, probably unnecessary with ISO cleaning above
			"hebrew",
			"csisolatinhebrew"},
		"iso-8859-8e": []string{ // MIB 84
			"csiso88598e",
			"iso-8859-8-e"},
		"iso-8859-8i": []string{ // MIB 85
			"csiso88598i",
			"iso-8859-8-i"},
		"iso-8859-10": []string{ // MIB 13
			"iso-ir-157",
			"l6",
			"iso_8859-10:1992",
			"csisolatin6",
			"latin6"},
		"iso-8859-13": []string{ // MIB 109
			"csiso885913"},
		"iso-8859-14": []string{ // MIB 110
			"iso-ir-199",
			"iso_8859-14:1998",
			"iso_8859-14",
			"latin8",
			"iso-celtic",
			"l8",
			"csiso885914"},
		"iso-8859-15": []string{ // MIB 111
			"iso_8859-15",
			"latin-9",
			"csiso885915",
			"ISO8859-15"},
		"iso-8859-16": []string{ // MIB 112
			"iso-ir-226",
			"iso_8859-16:2001",
			"iso_8859-16",
			"latin10",
			"l10",
			"csiso885916",
		},
		"windows-874": []string{ // MIB 2109
			"cswindows874",
			"cp874",
			"iso-8859-11",
			"tis-620",
		},
		"windows-1250": []string{ // MIB 2250
			"cswindows1250",
			"cp1250",
		},
		"windows-1251": []string{ // MIB 2251
			"cswindows1251",
			"cp1251",
		},
		"windows-1252": []string{ // MIB 2252
			"cswindows1252",
			"cp1252",
			"3dwindows-1252",
			"we8mswin1252",
		},
		"windows-1253": []string{"cswindows1253", "cp1253"},        // MIB 2253
		"windows-1254": []string{"cswindows1254", "cp1254"},        // MIB 2254
		"windows-1255": []string{"cswindows1255", "cp1255"},        // MIB 2255
		"windows-1256": []string{"cswindows1256", "cp1256"},        // MIB 2256
		"windows-1257": []string{"cswindows1257", "cp1257"},        // MIB 2257
		"windows-1258": []string{"cswindows1258", "cp1258"},        // MIB 2257 "koi8-r":       []string{"cskoi8r"},            // MIB 2084
		"koi8-u":       []string{"cskoi8u"},                        // MIB 2088
		"macintosh":    []string{"mac", "macroman", "csmacintosh"}, // MIB 2027
		"uhc": []string{ // Korea
			"ks_c_5601-1987",
			"ksc5601",
			"cp949",
		},
		"big5": []string{
			"big5",
		},
		"euckr": []string{
			"euc-kr", // MIB 38
			"ibm-euckr",
		},
		"euccn": []string{
			"ibm-euccn",
		},
		"eucjp": []string{
			"ibm-eucjp",
		},
		"iso2022jp": []string{
			"csiso2022jp",
			"iso-2022-jp", // MIB 39
		},

		"unkown": []string{
			"cp850",
			"cp858",            // "cp850"  Mostly correct except for the Euro sign
			"ansi_x3.110-1983", // MIB 74 // utf 8 probably
			"zht16mswin950",    // cp950
			"cp950",
		},
	}

	for expected, names := range mimesets {
		for _, name := range names {
			enc, err := getEncoding(name)
			if err != nil {
				t.Errorf("Error while get encoding %v: %v", name, err)
			}
			canonical, err := htmlindex.Name(enc)
			if err != nil {
				t.Errorf("Error while get canonical %v: %v", name, err)
			}
			if canonical != expected {
				t.Errorf("Expected %v but have %v", expected, canonical)
			}
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

		// add more from mutations of https://en.wikipedia.org/wiki/World_Wide_Web
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
