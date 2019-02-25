package pmmime

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"strings"

	"encoding/base64"
	htmlsets "golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
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

func getEncoding(charset string) (enc encoding.Encoding, err error) {
	enc, _ = htmlsets.Lookup(charset)
	if enc == nil {
		err = fmt.Errorf("Can not get encodig for '%s'", charset)
	}
	return
}

func selectDecoder(charset string) (decoder *encoding.Decoder, err error) {
	var enc encoding.Encoding
	lcharset := strings.Trim(strings.ToLower(charset), " \t\r\n")
	switch lcharset {
	case "utf-7", "unicode-1-1-utf-7":
		return NewUtf7Decoder(), nil
	default:
		enc, err = getEncoding(lcharset)
	}
	if err == nil {
		decoder = enc.NewDecoder()
	}
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
