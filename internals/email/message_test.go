package email

import (
	"bytes"
	"encoding/csv"
	"testing"
)

func TestToBytes(t *testing.T) {
	attachments := make(map[string][]byte)

	b := new(bytes.Buffer)
	w := csv.NewWriter(b)
	w.Comma = ';'
	w.Write([]string{"a", "b", "c", "d"})
	w.Write([]string{"1", "2", "3", "4"})
	w.Write([]string{"5", "6", "7", "8"})
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}
	attachments["fichier1.csv"] = b.Bytes()
	attachments["fichier2.csv"] = b.Bytes()

	message := Message{
		To:              []string{"destinataire1@gmail.com", "destinataire2@gmail.com"},
		CC:              []string{},
		BCC:             []string{},
		Subject:         "Complex subject title - 123456",
		BodyContentType: "text/html",
		Body:            "<body>\n\t<h2>Hello !<h2/>\n\t<p>Some random content...</p>\n</body>",
		Attachments:     attachments,
	}
	t.Log("\n" + string(message.ToBytes()))
}
