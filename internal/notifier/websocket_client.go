package notifier

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"

	"go.uber.org/zap"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is the maximum time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// pongWait is the maximum time we wait for a pong answer from the peer.
	pongWait = 60 * time.Second
	// pingPeriod is the interval at which pings are sent. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// sendBufferSize is the buffered capacity of the client send channel. A
	// large-enough buffer absorbs bursts while still allowing the notifier to
	// detect a genuinely stuck consumer (full buffer => message dropped).
	sendBufferSize = 256
	// wsAuthSubprotocol is the WebSocket subprotocol name used to negotiate
	// JWT authentication. The client offers [wsAuthSubprotocol, <token>] and
	// the server echoes wsAuthSubprotocol back to complete the handshake.
	wsAuthSubprotocol = "bearer"
)

// WebsocketClient structure represents a specific websocket connection, used by the manager
type WebsocketClient struct {
	GenericClient
	Socket  *websocket.Conn
	Receive chan []byte
}

// NewWebsocketClient creates a new client object containing the new connection
func NewWebsocketClient(conn *websocket.Conn, user *users.UserWithPermissions) *WebsocketClient {
	return &WebsocketClient{
		GenericClient: GenericClient{
			ID:   uuid.New().String(),
			Send: make(chan []byte, sendBufferSize),
			User: user,
		},
		Socket:  conn,
		Receive: make(chan []byte, 1),
	}
}

var upgrader = &websocket.Upgrader{
	// Allow the JWT to be passed through the WebSocket subprotocol so that it
	// does not have to travel in the URL query string.
	Subprotocols: []string{wsAuthSubprotocol},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// BuildWebsocketClient renders a new client after getting a new connection established
func BuildWebsocketClient(w http.ResponseWriter, r *http.Request, user *users.UserWithPermissions) (*WebsocketClient, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return NewWebsocketClient(conn, user), nil
}

// Write a message on a client socket
func (c *WebsocketClient) Write() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		destroyWebsocketClient(c)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Socket.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				zap.L().Info("Notification nok write, closing")
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			zap.L().Debug("Notification write msg", zap.ByteString("msg", message))
			if err := c.Socket.WriteMessage(websocket.TextMessage, message); err != nil {
				zap.L().Debug("Notification write failed", zap.Error(err))
				return
			}
		case <-ticker.C:
			// Send the Ping and return to close conn whether an error occurs
			_ = c.Socket.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Socket.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// Read consumes incoming messages from the client. Its main purpose is to keep
// the connection alive and to detect dead connections: a read deadline is set
// and extended each time a pong is received, so a peer that stops answering
// pings is detected and pruned instead of lingering in the client pool.
func (c *WebsocketClient) Read() {
	defer func() {
		destroyWebsocketClient(c)
	}()

	_ = c.Socket.SetReadDeadline(time.Now().Add(pongWait))
	c.Socket.SetPongHandler(func(string) error {
		return c.Socket.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		mt, message, err := c.Socket.ReadMessage()
		if err != nil {
			var closeError *websocket.CloseError
			switch {
			case errors.As(err, &closeError):
				if closeError.Code != websocket.CloseNormalClosure && closeError.Code != websocket.CloseGoingAway {
					zap.L().Error("Read socket", zap.Error(err))
				}
			default:
				zap.L().Debug("Read socket closed", zap.Error(err))
			}
			break
		}
		zap.L().Debug("message received", zap.ByteString("message", message), zap.String("client", c.ID))
		_ = mt
		// Forward to Receive without blocking: there is no consumer in
		// production, so a full/unconsumed channel must not stall the read
		// loop (which would defeat the liveness detection).
		select {
		case c.Receive <- message:
		default:
		}
	}
}

func destroyWebsocketClient(c *WebsocketClient) {
	if c == nil {
		return
	}
	err := C().Unregister(c)
	if err != nil {
		zap.L().Debug("Could not unregister ws client", zap.Error(err), zap.String("id", c.ID))
	}
	err = c.Socket.Close()
	if err != nil {
		zap.L().Debug("Could not close ws client socket", zap.Error(err), zap.String("id", c.ID))
	}
}
