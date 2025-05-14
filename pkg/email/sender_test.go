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
		Body:            `<table cellspacing="0" cellpadding="1" style="border-collapse: collapse;"><tr><th style="border: 1px solid rgb(0,0,0);padding: 5px;font-weight: bold;">Code Site</th><th style="border: 1px solid rgb(0,0,0);padding: 5px;font-weight: bold;">Nb colis manquant</th></tr><tr><td style="border: 1px solid rgb(0,0,0);padding: 5px;">bucket1</td><td style="border: 1px solid rgb(0,0,0);padding: 5px;">10</td></tr><tr><td style="border: 1px solid rgb(0,0,0);padding: 5px;">bucket2</td><td style="border: 1px solid rgb(0,0,0);padding: 5px;">10</td></tr><tr><td style="border: 1px solid rgb(0,0,0);padding: 5px;">bucket3</td><td style="border: 1px solid rgb(0,0,0);padding: 5px;">10</td></tr></table>`,
		Attachments:     attachments,
	}

	return message
}

func TestSenderSend(t *testing.T) {
	t.Skip() // Development test
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
	t.Skip() // Development test
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
