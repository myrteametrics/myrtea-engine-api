package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/multipart"
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
	Attachments     []MessageAttachment
}

type MessageAttachment struct {
	FileName string
	Mime     string
	Content  []byte
}

func NewMessage(subject string, bodyContentType string, body string) Message {
	return Message{
		Subject:         subject,
		BodyContentType: bodyContentType,
		Body:            body,
		Attachments:     make([]MessageAttachment, 0),
	}
}

func (m *Message) AttachFileBytes(fileName string, mime string, content []byte) {
	m.Attachments = append(m.Attachments, MessageAttachment{FileName: fileName, Mime: mime, Content: content})
}

func (m *Message) AttachFile(src string, mime string, fileName string) error {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	_fileName := fileName
	if _fileName == "" {
		_, _fileName = filepath.Split(src)
	}
	m.AttachFileBytes(_fileName, mime, b)
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
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n", boundary))

	buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
	buf.WriteString(fmt.Sprintf("Content-Type: %s; charset=\"UTF-8\"\n", m.BodyContentType))
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\n")
	buf.WriteString("\n")
	buf.WriteString(strings.ReplaceAll(m.Body, "=\"", "=3D\""))
	buf.WriteString("\n")

	if len(m.Attachments) > 0 {
		for _, attachment := range m.Attachments {
			buf.WriteString("\n")
			buf.WriteString("\n")
			buf.WriteString(fmt.Sprintf("--%s\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\n", attachment.Mime, attachment.FileName))
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\n", attachment.FileName))
			buf.WriteString("Content-Transfer-Encoding: base64\n")
			buf.WriteString("\n")
			b := make([]byte, base64.StdEncoding.EncodedLen(len(attachment.Content)))
			base64.StdEncoding.Encode(b, attachment.Content)
			buf.Write(b)
			buf.WriteString("\n")
		}
	}

	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("--%s--", boundary))

	return buf.Bytes()
}
