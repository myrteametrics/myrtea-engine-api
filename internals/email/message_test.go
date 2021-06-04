package email

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"testing"
)

func TestDetectContentType(t *testing.T) {
	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	w.Comma = ','
	w.Write([]string{"a", "b", "c", "d"})
	w.Write([]string{"1", "2", "3", "4"})
	w.Write([]string{"5", "6", "7", "8"})
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}
	data := b.Bytes()
	t.Log(http.DetectContentType(data))
}

func TestToBytes(t *testing.T) {

	b := new(bytes.Buffer)
	w := csv.NewWriter(b)

	w.Write([]string{"a", "b", "c", "d"})
	w.Write([]string{"1", "2", "3", "4"})
	w.Write([]string{"5", "6", "7", "8"})
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}

	attachments := make([]MessageAttachment, 0)
	attachments = append(attachments, MessageAttachment{FileName: "fichier1.csv", Mime: "application/octet-stream", Content: b.Bytes()})
	attachments = append(attachments, MessageAttachment{FileName: "fichier2.csv", Mime: "application/octet-stream", Content: b.Bytes()})

	message := Message{
		To:              []string{"receiver1@gmail.com", "receiver2@gmail.com"},
		CC:              []string{},
		BCC:             []string{},
		Subject:         "Complex subject title - 123456",
		BodyContentType: "text/html",
		Body:            "<body>\n\t<h2>Hello !<h2/>\n\t<p>Some random content...</p>\n</body>",
		Attachments:     attachments,
	}
	t.Log("\n" + string(message.ToBytes()))
}
