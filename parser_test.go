package gomime

import (
	"bytes"

	"io/ioutil"
	"net/mail"

	"net/textproto"
	"strings"
	"testing"
)

func minimalParse(mimeBody string) (readBody string, plainContents string, err error) {
	mm, err := mail.ReadMessage(strings.NewReader(mimeBody))
	if err != nil {
		return
	}

	h := textproto.MIMEHeader(mm.Header)
	mmBodyData, err := ioutil.ReadAll(mm.Body)
	if err != nil {
		return
	}

	printAccepter := NewMIMEPrinter()
	plainTextCollector := NewPlainTextCollector(printAccepter)
	visitor := NewMimeVisitor(plainTextCollector)
	err = VisitAll(bytes.NewReader(mmBodyData), h, visitor)

	readBody = printAccepter.String()
	plainContents = plainTextCollector.GetPlainText()

	return readBody, plainContents, err
}

func androidParse(mimeBody string) (body, headers string, atts, attHeaders []string, err error) {
	mm, err := mail.ReadMessage(strings.NewReader(mimeBody))
	if err != nil {
		return
	}

	h := textproto.MIMEHeader(mm.Header)
	mmBodyData, err := ioutil.ReadAll(mm.Body)

	printAccepter := NewMIMEPrinter()
	bodyCollector := NewBodyCollector(printAccepter)
	attachmentsCollector := NewAttachmentsCollector(bodyCollector)
	mimeVisitor := NewMimeVisitor(attachmentsCollector)
	err = VisitAll(bytes.NewReader(mmBodyData), h, mimeVisitor)

	body, _ = bodyCollector.GetBody()
	headers = bodyCollector.GetHeaders()
	atts = attachmentsCollector.GetAttachments()
	attHeaders = attachmentsCollector.GetAttHeaders()

	return
}

func TestParseBoundaryIsEmpty(t *testing.T) {
	testMessage :=
		`Date: Sun, 10 Mar 2019 11:10:06 -0600
In-Reply-To: <abcbase64@protonmail.com>
X-Original-To: enterprise@protonmail.com
References: <abc64@unicoderns.com> <abc63@protonmail.com> <abc64@protonmail.com> <abc65@mail.gmail.com> <abc66@protonmail.com>
To: "ProtonMail" <enterprise@protonmail.com>
X-Pm-Origin: external
Delivered-To: enterprise@protonmail.com
Content-Type: multipart/mixed; boundary=ac7e36bd45425e70b4dab2128f34172e4dc3f9ff2eeb47e909267d4252794ec7
Reply-To: XYZ <xyz@xyz.com>
Mime-Version: 1.0
Subject: Encrypted Message
Return-Path: <xyz@xyz.com>
From: XYZ <xyz@xyz.com>
X-Pm-Conversationid-Id: gNX9bDPLmBgFZ-C3Tdlb628cas1Xl0m4dql5nsWzQAEI-WQv0ytfwPR4-PWELEK0_87XuFOgetc239Y0pjPYHQ==
X-Pm-Date: Sun, 10 Mar 2019 18:10:06 +0100
Message-Id: <68c11e46-e611-d9e4-edc1-5ec96bac77cc@unicoderns.com>
X-Pm-Transfer-Encryption: TLSv1.2 with cipher ECDHE-RSA-AES256-GCM-SHA384 (256/256 bits)
X-Pm-External-Id: <68c11e46-e611-d9e4-edc1-5ec96bac77cc@unicoderns.com>
X-Pm-Internal-Id: _iJ8ETxcqXTSK8IzCn0qFpMUTwvRf-xJUtldRA1f6yHdmXjXzKleG3F_NLjZL3FvIWVHoItTxOuuVXcukwwW3g==
Openpgp: preference=signencrypt
User-Agent: Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Thunderbird/60.4.0
X-Pm-Content-Encryption: end-to-end

--ac7e36bd45425e70b4dab2128f34172e4dc3f9ff2eeb47e909267d4252794ec7
Content-Disposition: inline
Content-Transfer-Encoding: quoted-printable
Content-Type: multipart/mixed; charset=utf-8

Content-Type: multipart/mixed; boundary="xnAIW3Turb9YQZ2rXc2ZGZH45WepHIZyy";
 protected-headers="v1"
From: XYZ <xyz@xyz.com>
To: "ProtonMail" <enterprise@protonmail.com>
Subject: Encrypted Message
Message-ID: <68c11e46-e611-d9e4-edc1-5ec96bac77cc@unicoderns.com>
References: <abc64@unicoderns.com> <abc63@protonmail.com> <abc64@protonmail.com> <abc65@mail.gmail.com> <abc66@protonmail.com>
In-Reply-To: <abcbase64@protonmail.com>

--xnAIW3Turb9YQZ2rXc2ZGZH45WepHIZyy
Content-Type: text/rfc822-headers; protected-headers="v1"
Content-Disposition: inline

From: XYZ <xyz@xyz.com>
To: ProtonMail <enterprise@protonmail.com>
Subject: Re: Encrypted Message

--xnAIW3Turb9YQZ2rXc2ZGZH45WepHIZyy
Content-Type: multipart/alternative;
 boundary="------------F9E5AA6D49692F51484075E3"
Content-Language: en-US

This is a multi-part message in MIME format.
--------------F9E5AA6D49692F51484075E3
Content-Type: text/plain; charset=utf-8
Content-Transfer-Encoding: quoted-printable

Hi ...

--------------F9E5AA6D49692F51484075E3
Content-Type: text/html; charset=utf-8
Content-Transfer-Encoding: quoted-printable

<html>
  <head>
  </head>
  <body text=3D"#000000" bgcolor=3D"#FFFFFF">
    <p>Hi ..  </p>
  </body>
</html>

--------------F9E5AA6D49692F51484075E3--

--xnAIW3Turb9YQZ2rXc2ZGZH45WepHIZyy--

--ac7e36bd45425e70b4dab2128f34172e4dc3f9ff2eeb47e909267d4252794ec7--


`

	body, content, err := minimalParse(testMessage)
	if err == nil {
		t.Fatal("should have error but is", err)
	}
	t.Log("==BODY==")
	t.Log(body)
	t.Log("==CONTENT==")
	t.Log(content)
}

func TestParse(t *testing.T) {
	testMessage :=
		`From: John Doe <example@example.com>
MIME-Version: 1.0
Content-Type: multipart/mixed;
        boundary="XXXXboundary text"

This is a multipart message in MIME format.

--XXXXboundary text
Content-Type: text/plain; charset=utf-8

this is the body text

--XXXXboundary text
Content-Type: text/html; charset=utf-8

<html><body>this is the html body text</body></html>

--XXXXboundary text
Content-Type: text/plain; charset=utf-8
Content-Disposition: attachment;
        filename="test.txt"

this is the attachment text

--XXXXboundary text
Content-Type: image/jpeg; name="images-2.jpg"
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="images-2.jpg"

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
0//Z
--XXXXboundary text--


`
	_, _, _, _, err := androidParse(testMessage)
	if err != nil {
		t.Error("parse error", err)
	}
}
