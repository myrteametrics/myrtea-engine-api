package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

type Message struct {
	To              []string
	CC              []string
	BCC             []string
	Subject         string
	BodyContentType string // text/plain, text/html
	Body            string
	Attachments     map[string][]byte
}

func NewMessage(subject string, bodyContentType string, body string) Message {
	return Message{
		Subject:         subject,
		BodyContentType: bodyContentType,
		Body:            body,
		Attachments:     make(map[string][]byte),
	}
}

func (m *Message) AttachFileBytes(fileName string, b []byte) {
	m.Attachments[fileName] = b
}

func (m *Message) AttachFile(src string) error {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	_, fileName := filepath.Split(src)
	m.AttachFileBytes(fileName, b)
	return nil
}

func (m *Message) ToBytes() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("MIME-Version: 1.0\n")
	buf.WriteString(fmt.Sprintf("Subject: %s\n", m.Subject))
	buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(m.To, ",")))
	if len(m.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\n", strings.Join(m.CC, ",")))
	}
	if len(m.BCC) > 0 {
		buf.WriteString(fmt.Sprintf("Bcc: %s\n", strings.Join(m.BCC, ",")))
	}

	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	// if len(m.Attachments) > 0 {
	// 	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n", boundary))
	// }
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n", boundary))

	buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
	buf.WriteString(fmt.Sprintf("Content-Type: %s; charset=\"UTF-8\"\n", m.BodyContentType))
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\n")
	buf.WriteString(m.Body)

	if len(m.Attachments) > 0 {
		for k, v := range m.Attachments {
			buf.WriteString(fmt.Sprintf("\n\n\n--%s\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\n", http.DetectContentType(v)))
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\n", k))
			buf.WriteString("Content-Transfer-Encoding: base64\n")

			b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
			base64.StdEncoding.Encode(b, v)
			buf.Write(b)
			// buf.WriteString(fmt.Sprintf("\n--%s", boundary))
		}
		// buf.WriteString("--")
		buf.WriteString(fmt.Sprintf("\n\n\n--%s--", boundary))
	}

	return buf.Bytes()
}
