package routeroidc

import (
	"net/http"
	"strings"
)

const (
	tokenPrefix       = "Bearer "
	tokenKey          = "jwt"
	wsAuthSubprotocol = "bearer"
)

// tokenFromWebSocketProtocol extracts a JWT from Sec-WebSocket-Protocol.
// Browsers cannot set custom headers on a WS handshake, so the token is
// passed as a subprotocol value alongside the sentinel "bearer" subprotocol.
func tokenFromWebSocketProtocol(r *http.Request) string {
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
