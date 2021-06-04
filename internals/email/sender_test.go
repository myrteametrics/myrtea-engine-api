package email

import (
	"bytes"
	"encoding/csv"
	"os"
	"testing"
)

func mockMessage(t *testing.T) Message {
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
		Body:            "<div>hello<div><div><table><tr><th>Code</th><th>Value</th></tr><tr><td>A</td><td>20</td></tr><tr><td>B</td><td>40</td></tr><tr><td>C</td><td>2500</td></tr></table><div>",
		Attachments:     attachments,
	}

	return message
}

func TestSenderSend(t *testing.T) {
	host := os.Getenv("TEST_SMTP_HOST")
	port := os.Getenv("TEST_SMTP_PORT")
	username := os.Getenv("TEST_SMTP_USERNAME")
	password := os.Getenv("TEST_SMTP_PASSWORD")

	sender := NewSender(username, password, host, port)
	err := sender.Send(mockMessage(t))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}

func TestSenderSendGmail(t *testing.T) {
	username := os.Getenv("TEST_SMTP_USERNAME")
	password := os.Getenv("TEST_SMTP_PASSWORD")
	host := "smtp.gmail.com"
	port := "587"

	sender := NewSender(username, password, host, port)
	err := sender.Send(mockMessage(t))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
