package mimeparser

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/textproto"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type VisitAcceptor interface {
	Accept(partReader io.Reader, header textproto.MIMEHeader, hasPlainSibling bool, isFirst, isLast bool)
}

func VisitAll(part io.Reader, h textproto.MIMEHeader, accepter VisitAcceptor) (err error) {
	return visit(part, h, accepter, false)
}

func IsLeaf(h textproto.MIMEHeader) bool {
	return !strings.HasPrefix(h.Get("Content-Type"), "multipart/")
}

func visit(part io.Reader, h textproto.MIMEHeader, accepter VisitAcceptor, hasPlainSibling bool) (err error) {

	parentMediaType, params, err := mime.ParseMediaType(h.Get("Content-Type"))
	if err != nil {
		return
	}

	accepter.Accept(part, h, hasPlainSibling, true, false)

	if !IsLeaf(h) {
		var multiparts []io.Reader
		var multipartHeaders []textproto.MIMEHeader
		if multiparts, multipartHeaders, err = getMultipartParts(part, params); err != nil {
			return
		} else {
			hasPlainChild := false
			for _, header := range multipartHeaders {
				mediaType, _, _ := mime.ParseMediaType(header.Get("Content-Type"))
				if mediaType == "text/plain" {
					hasPlainChild = true
				}
			}
			if hasPlainSibling && parentMediaType == "multipart/related" {
				hasPlainChild = true
			}

			for i, p := range multiparts {
				if err = visit(p, multipartHeaders[i], accepter, hasPlainChild); err != nil {
					return
				}
				accepter.Accept(part, h, hasPlainSibling, false, i == (len(multiparts)-1))
			}
		}
	}
	return

}

func getAllChildParts(part io.Reader, h textproto.MIMEHeader) (parts []io.Reader, headers []textproto.MIMEHeader, err error) {

	mediaType, params, err := mime.ParseMediaType(h.Get("Content-Type"))
	if err != nil {
		return
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		var multiparts []io.Reader
		var multipartHeaders []textproto.MIMEHeader
		if multiparts, multipartHeaders, err = getMultipartParts(part, params); err != nil {
			return
		}
		if strings.Contains(mediaType, "alternative") {
			var chosenPart io.Reader
			var chosenHeader textproto.MIMEHeader
			if chosenPart, chosenHeader, err = pickAlternativePart(multiparts, multipartHeaders); err != nil {
				return
			}
			var childParts []io.Reader
			var childHeaders []textproto.MIMEHeader
			if childParts, childHeaders, err = getAllChildParts(chosenPart, chosenHeader); err != nil {
				return
			}
			parts = append(parts, childParts...)
			headers = append(headers, childHeaders...)
		} else {
			for i, p := range multiparts {
				var childParts []io.Reader
				var childHeaders []textproto.MIMEHeader
				if childParts, childHeaders, err = getAllChildParts(p, multipartHeaders[i]); err != nil {
					return
				}
				parts = append(parts, childParts...)
				headers = append(headers, childHeaders...)
			}
		}
	} else {
		parts = append(parts, part)
		headers = append(headers, h)
	}
	return
}

func getMultipartParts(r io.Reader, params map[string]string) (parts []io.Reader, headers []textproto.MIMEHeader, err error) {
	mr := multipart.NewReader(r, params["boundary"])
	parts = []io.Reader{}
	headers = []textproto.MIMEHeader{}
	var p *multipart.Part
	for {
		p, err = mr.NextPart()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return
		}
		b, _ := ioutil.ReadAll(p)
		buffer := bytes.NewBuffer(b)

		parts = append(parts, buffer)
		headers = append(headers, p.Header)
	}
	return
}

func pickAlternativePart(parts []io.Reader, headers []textproto.MIMEHeader) (part io.Reader, h textproto.MIMEHeader, err error) {

	for i, h := range headers {
		mediaType, _, err := mime.ParseMediaType(h.Get("Content-Type"))
		if err != nil {
			continue
		}
		if strings.HasPrefix(mediaType, "multipart/") {
			return parts[i], headers[i], nil
		}
	}
	for i, h := range headers {
		mediaType, _, err := mime.ParseMediaType(h.Get("Content-Type"))
		if err != nil {
			continue
		}
		if mediaType == "text/html" {
			return parts[i], headers[i], nil
		}
	}
	for i, h := range headers {
		mediaType, _, err := mime.ParseMediaType(h.Get("Content-Type"))
		if err != nil {
			continue
		}
		if mediaType == "text/plain" {
			return parts[i], headers[i], nil
		}
	}
	//if we get all the way here, part will be nil
	return
}

// Parse address comment as defined in http://tools.wordtothewise.com/rfc/822
// FIXME: Does not work for address groups
// NOTE: This should be removed for go>1.10 (please check)
func parseAddressComment(raw string) string {
	parsed := []string{}
	for _, item := range regexp.MustCompile("[,;]").Split(raw, -1) {
		re := regexp.MustCompile("[(][^)]*[)]")
		comments := strings.Join(re.FindAllString(item, -1), " ")
		comments = strings.Replace(comments, "(", "", -1)
		comments = strings.Replace(comments, ")", "", -1)
		withoutComments := re.ReplaceAllString(item, "")
		addr, err := mail.ParseAddress(withoutComments)
		if err != nil {
			continue
		}
		if addr.Name == "" {
			addr.Name = comments
		}
		parsed = append(parsed, addr.String())
	}
	return strings.Join(parsed, ", ")
}

func checkHeaders(headers []textproto.MIMEHeader) bool {
	foundAttachment := false

	for i := 0; i < len(headers); i++ {
		h := headers[i]

		mediaType, _, _ := mime.ParseMediaType(h.Get("Content-Type"))

		if !strings.HasPrefix(mediaType, "text/") {
			foundAttachment = true
		} else if foundAttachment {
			//this means that there is a text part after the first attachment, so we will have to convert the body from plain->HTML
			return true
		}
	}
	return false
}

func decodePart(partReader io.Reader, header textproto.MIMEHeader) (decodedPart io.Reader) {
	decodedPart = DecodeContentEncoding(partReader, header.Get("Content-Transfer-Encoding"))
	if decodedPart == nil {
		log.Warnf("Unsupported Content-Transfer-Encoding '%v'", header.Get("Content-Transfer-Encoding"))
		decodedPart = partReader
	}
	return
}

// ===================== MIME Printer ===================================
// Simply print resulting MIME tree into text form
type stack []string

func (s stack) Push(v string) stack {
	return append(s, v)
}
func (s stack) Pop() (stack, string) {
	l := len(s)
	return s[:l-1], s[l-1]
}
func (s stack) Peek() string {
	return s[len(s)-1]
}

type MIMEPrinter struct {
	result        *bytes.Buffer
	boundaryStack stack
}

func NewMIMEPrinter() (pd *MIMEPrinter) {
	return &MIMEPrinter{
		result:        bytes.NewBuffer([]byte("")),
		boundaryStack: stack{},
	}
}

func (pd *MIMEPrinter) Accept(partReader io.Reader, header textproto.MIMEHeader, hasPlainSibling bool, isFirst, isLast bool) {
	if isFirst {
		http.Header(header).Write(pd.result)
		pd.result.Write([]byte("\n"))
		if IsLeaf(header) {
			pd.result.ReadFrom(partReader)
		} else {
			_, params, _ := mime.ParseMediaType(header.Get("Content-Type"))
			boundary := params["boundary"]
			pd.boundaryStack = pd.boundaryStack.Push(boundary)
			pd.result.Write([]byte("\nThis is a multi-part message in MIME format.\n--" + boundary + "\n"))
		}
	} else {
		if !isLast {
			pd.result.Write([]byte("\n--" + pd.boundaryStack.Peek() + "\n"))
		} else {
			var boundary string
			pd.boundaryStack, boundary = pd.boundaryStack.Pop()
			pd.result.Write([]byte("\n--" + boundary + "--\n.\n"))
		}
	}
}

func (pd *MIMEPrinter) String() string {
	return pd.result.String()
}

// ======================== PlainText Collector  =========================
// Collect contents of all non-attachment text/plain parts and return
// it is a string

type PlainTextCollector struct {
	target            VisitAcceptor
	plainTextContents *bytes.Buffer
}

func NewPlainTextCollector(targetAccepter VisitAcceptor) *PlainTextCollector {
	return &PlainTextCollector{
		target:            targetAccepter,
		plainTextContents: bytes.NewBuffer([]byte("")),
	}
}

func (ptc PlainTextCollector) Accept(partReader io.Reader, header textproto.MIMEHeader, hasPlainSibling bool, isFirst, isLast bool) {
	if isFirst {
		if IsLeaf(header) {
			mediaType, params, _ := mime.ParseMediaType(header.Get("Content-Type"))
			disp, _, _ := mime.ParseMediaType(header.Get("Content-Disposition"))
			if mediaType == "text/plain" && disp != "attachment" {
				partData, _ := ioutil.ReadAll(partReader)
				decodedPart := decodePart(bytes.NewReader(partData), header)

				if buffer, err := ioutil.ReadAll(decodedPart); err == nil {
					buffer, err = DecodeCharset(buffer, params)
					if err != nil {
						log.Warnln("Decode charset error:", err)
					}
					ptc.plainTextContents.Write(buffer)
				}

				ptc.target.Accept(bytes.NewReader(partData), header, hasPlainSibling, isFirst, isLast)
				return
			}
		}
	}
	ptc.target.Accept(partReader, header, hasPlainSibling, isFirst, isLast)
}

func (ptc PlainTextCollector) GetPlainText() string {
	return ptc.plainTextContents.String()
}
