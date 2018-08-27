package mimeparser

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/quotedprintable"
	"strings"

	"encoding/base64"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var wordDec = &mime.WordDecoder{
	CharsetReader: func(charset string, input io.Reader) (io.Reader, error) {
		dec, err := selectDecoder(charset)
		if err != nil {
			return nil, err
		}
		if dec == nil { // utf-8
			return input, nil
		}
		return dec.Reader(input), nil
	},
}

func selectDecoder(charset string) (decoder *encoding.Decoder, err error) {
	var enc encoding.Encoding
	// all MIME charset with aliases can be found here https://www.iana.org/assignments/character-sets/character-sets.xhtml
	switch strings.ToLower(charset) {
	case "utf-8", // MIB 16
		"utf8",
		"csutf8",
		"us-ascii", // MIB 3
		"iso-ir-6",
		"ansi_x3.4-1968",
		"ansi_x3.4-1986",
		"iso_646.irv:1991",
		"iso646-us",
		"us",
		"ibm367",
		"cp367",
		"csascii",
		"ascii",
		"ks_c_5601-1987",   // MIB 36
		"euc-kr",           // MIB 38
		"iso-2022-jp",      // MIB 39
		"ansi_x3.110-1983", // MIB 74
		"gb2312",           // MIB 2025
		"":                 // Nothing to do
		return
	case "utf7", "utf-7":
		return NewUtf7Decoder(), nil
	case "iso-8859-1", // MIB 4
		"iso8859-1",
		"so-ir-100",
		"iso_8859-1",
		"latin1",
		"l1",
		"ibm819",
		"cp819",
		"csisolatin1":
		enc = charmap.ISO8859_1
	case "iso-8859-2", // MIB 5
		"iso-ir-101",
		"iso_8859-2",
		"iso8859-2",
		"latin2",
		"l2",
		"csisolatin2":
		enc = charmap.ISO8859_2
	case "iso-8859-3", // MIB 6
		"iso-ir-109",
		"iso_8859-3",
		"latin3",
		"l3",
		"csisolatin3":
		enc = charmap.ISO8859_3
	case "iso-8859-4", // MIB 7
		"iso-ir-110",
		"iso_8859-4",
		"latin4",
		"l4",
		"csisolatin4":
		enc = charmap.ISO8859_4
	case "iso-8859-5", // MIB 8
		"iso-ir-144",
		"iso_8859-5",
		"cyrillic",
		"csisolatincyrillic":
		enc = charmap.ISO8859_5
	case "iso-8859-6", // MIB 9
		"iso-ir-127",
		"iso_8859-6",
		"ecma-114",
		"asmo-708",
		"arabic",
		"csisolatinarabic":
		enc = charmap.ISO8859_6
	case "iso-8859-6e", // MIB 81
		"csiso88596e",
		"iso-8859-6-e":
		enc = charmap.ISO8859_6E
	case "iso-8859-6i", // MIB 82
		"csiso88596i",
		"iso-8859-6-i":
		enc = charmap.ISO8859_6I
	case "iso-8859-7", // MIB 10
		"iso-ir-126",
		"iso_8859-7",
		"elot_928",
		"ecma-118",
		"greek",
		"greek8",
		"csisolatingreek":
		enc = charmap.ISO8859_7
	case "iso-8859-8", // MIB 11
		"iso-ir-138",
		"iso_8859-8",
		"hebrew",
		"csisolatinhebrew":
		enc = charmap.ISO8859_8
	case "iso-8859-8e", // MIB 84
		"csiso88598e",
		"iso-8859-8-e":
		enc = charmap.ISO8859_8E
	case "iso-8859-8i", // MIB 85
		"csiso88598i",
		"iso-8859-8-i":
		enc = charmap.ISO8859_8I
	case "iso-8859-10", // MIB 13
		"iso-ir-157",
		"l6",
		"iso_8859-10:1992",
		"csisolatin6",
		"latin6":
		enc = charmap.ISO8859_10
	case "iso-8859-13", // MIB 109
		"csiso885913":
		enc = charmap.ISO8859_13
	case "iso-8859-14", // MIB 110
		"iso-ir-199",
		"iso_8859-14:1998",
		"iso_8859-14",
		"latin8",
		"iso-celtic",
		"l8",
		"csiso885914":
		enc = charmap.ISO8859_14
	case "iso-8859-15", // MIB 111
		"iso_8859-15",
		"latin-9",
		"csiso885915",
		"ISO8859-15":
		enc = charmap.ISO8859_15
	case "iso-8859-16", // MIB 112
		"iso-ir-226",
		"iso_8859-16:2001",
		"iso_8859-16",
		"latin10",
		"l10",
		"csiso885916":
		enc = charmap.ISO8859_16
	case "windows-874", "cswindows874": // MIB 2109
		enc = charmap.Windows874
	case "windows-1250", "cswindows1250": // MIB 2250
		enc = charmap.Windows1250
	case "windows-1251", "cswindows1251": // MIB 2251
		enc = charmap.Windows1251
	case "windows-1252", "cswindows1252", "cp1252", "3dwindows-1252": // MIB 2252
		enc = charmap.Windows1252
	case "windows-1253", "cswindows1253": // MIB 2253
		enc = charmap.Windows1253
	case "windows-1254", "cswindows1254": // MIB 2254
		enc = charmap.Windows1254
	case "windows-1255", "cswindows1255": // MIB 2255
		enc = charmap.Windows1255
	case "windows-1256", "cswindows1256": // MIB 2256
		enc = charmap.Windows1256
	case "windows-1257", "cswindows1257": // MIB 2257
		enc = charmap.Windows1257
	case "koi8-r", "cskoi8r": // MIB 2084
		enc = charmap.KOI8R
	case "koi8-u", "cskoi8u": // MIB 2088
		enc = charmap.KOI8U
	case "macintosh", "mac", "csmacintosh": // MIB 2027
		enc = charmap.Macintosh
	default:
		err = errors.New("Unsupported charset " + charset)
		return
	}
	decoder = enc.NewDecoder()
	return
}

func DecodeHeader(raw string) (decoded string, err error) {
	if decoded, err = wordDec.DecodeHeader(raw); err != nil {
		decoded = raw
	}
	return
}

func EncodeHeader(s string) string {
	return mime.QEncoding.Encode("utf-8", s)
}

func DecodeCharset(original []byte, parameters map[string]string) ([]byte, error) {
	charset, ok := parameters["charset"]
	decoder, err := selectDecoder(charset)
	if len(original) == 0 || !ok || decoder == nil {
		return original, err
	}

	utf8 := make([]byte, len(original))
	nDst, nSrc, err := decoder.Transform(utf8, original, false)
	for err == transform.ErrShortDst {
		utf8 = make([]byte, (nDst/nSrc+1)*len(original))
		nDst, nSrc, err = decoder.Transform(utf8, original, false)
	}
	if err != nil {
		return original, err
	}
	utf8 = bytes.Trim(utf8, "\x00")

	return utf8, nil
}

func DecodeContentEncoding(r io.Reader, contentEncoding string) (d io.Reader) {
	switch strings.ToLower(contentEncoding) {
	case "quoted-printable":
		d = quotedprintable.NewReader(r)
	case "base64":
		d = base64.NewDecoder(base64.StdEncoding, r)
	case "7bit", "8bit", "binary", "": // Nothing to do
		d = r
	}
	return
}
