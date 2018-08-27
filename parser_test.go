package mimeparser

import (
	"bytes"
	"fmt"

	"io/ioutil"
	"net/mail"

	"net/textproto"
	"strings"
	"testing"
)

func MinimalParse(mimeBody string) (readBody string, plainContents string, err error) {

	mm, err := mail.ReadMessage(strings.NewReader(mimeBody))
	if err != nil {
		return
	}

	h := textproto.MIMEHeader(mm.Header)
	mmBodyData, err := ioutil.ReadAll(mm.Body)

	printAccepter := NewMIMEPrinter()
	plainTextCollector := NewPlainTextCollector(printAccepter)
	err = VisitAll(bytes.NewReader(mmBodyData), h, plainTextCollector)

	readBody = printAccepter.String()
	plainContents = plainTextCollector.GetPlainText()

	return readBody, plainContents, err
}

func AndroidParse(mimeBody string) (readBody string, plainContents string, err error) {
	// TODO: samle, edit as needed
	return MinimalParse(mimeBody)
}

func TestParse(t *testing.T) {
	testMessage :=
		`From: John Doe <example@example.com>
MIME-Version: 1.0
Content-Type: multipart/mixed;
        boundary="XXXXboundary text"

This is a multipart message in MIME format.

--XXXXboundary text 
Content-Type: text/plain

this is the body text

--XXXXboundary text 
Content-Type: text/plain;
Content-Disposition: attachment;
        filename="test.txt"

this is the attachment text

--XXXXboundary text--


`
	fmt.Println("+++")
	body, plain, err := MinimalParse(testMessage)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(body)
	fmt.Println(plain)
}
