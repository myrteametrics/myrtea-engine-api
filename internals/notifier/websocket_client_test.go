package notifier

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestNewWSClient(t *testing.T) {
	ReplaceGlobals(NewNotifier())

	// Server-side initialisation
	var client *WebsocketClient
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		client, err = BuildWebsocketClient(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer s.Close()

	// Client-side initialisation
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http"), nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// Tests
	if client == nil {
		t.Fatal("Client not built")
	}
}

func TestWSClientRead(t *testing.T) {
	ReplaceGlobals(NewNotifier())

	// Server-side initialisation
	var client *WebsocketClient
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		client, err = BuildWebsocketClient(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		go client.Read()
	}))
	defer s.Close()

	// Client-side initialisation
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http"), nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// Tests
	for i := 0; i < 10; i++ {
		if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
			t.Fatalf("%v", err)
		}

		// Read message directly from the client Receive channel
		message, ok := <-client.Receive
		if !ok {
			t.Fatalf("Cannot read Receive channel")
		}
		if string(message) != "hello" {
			t.Fatalf("bad message")
		}
	}
}

func TestWSClientWrite(t *testing.T) {
	ReplaceGlobals(NewNotifier())

	// Server-side initialisation
	var client *WebsocketClient
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		client, err = BuildWebsocketClient(w, r, nil)
		if err != nil {
			t.Fatal(err)
		}
		go client.Write()
	}))
	defer s.Close()

	// Client-side initialisation
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(s.URL, "http"), nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// Tests
	for i := 0; i < 10; i++ {
		// Send message directly on the client Send channel
		client.Send <- []byte("hello")

		mt, message, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if mt != websocket.TextMessage {
			t.Fatalf("Invalid message type")
		}
		if string(message) != "hello" {
			t.Fatalf("bad message")
		}
	}
}
