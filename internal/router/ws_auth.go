package router

import (
	"net/http"
	"strings"
)

// wsAuthSubprotocol is the WebSocket subprotocol name used to negotiate JWT
// authentication. It must match the one used by the notifier upgrader.
const wsAuthSubprotocol = "bearer"

// TokenFromWebSocketProtocol extracts a JWT from the Sec-WebSocket-Protocol
// header. Browsers cannot set custom headers on a WebSocket handshake, so the
// token is passed as a subprotocol value: the client offers
// [wsAuthSubprotocol, <token>] and we read the token here. This avoids leaking
// the JWT in the URL query string (and therefore in server/proxy access logs).
func TokenFromWebSocketProtocol(r *http.Request) string {
	header := r.Header.Get("Sec-WebSocket-Protocol")
	if header == "" {
		return ""
	}
	for _, part := range strings.Split(header, ",") {
		value := strings.TrimSpace(part)
		if value == "" || value == wsAuthSubprotocol {
			continue
		}
		return value
	}
	return ""
}
