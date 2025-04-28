package notifier

import (
	"errors"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/gorilla/websocket"
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
			Send: make(chan []byte, 1),
			User: user,
		},
		Socket:  conn,
		Receive: make(chan []byte, 1),
	}
}

var upgrader = &websocket.Upgrader{
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
	ticker := time.NewTicker(10 * time.Second)

	defer func() {
		ticker.Stop()
		destroyWebsocketClient(c)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				zap.L().Info("Notification nok write, closing")
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			zap.L().Info("Notification write msg", zap.ByteString("msg", message))
			c.Socket.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			// Send the Ping and return to close conn whether an error occurs
			if err := c.Socket.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// Read a message from one client and broadcast it to others
func (c *WebsocketClient) Read() {
	defer func() {
		destroyWebsocketClient(c)
	}()

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
				zap.L().Error("Read socket", zap.Error(err))
			}
			break
		}
		zap.L().Info("message received", zap.ByteString("message", message), zap.String("client", c.ID))
		_ = mt
		c.Receive <- message
	}
}

func destroyWebsocketClient(c *WebsocketClient) {
	if c == nil {
		return
	}
	err := C().Unregister(c)
	if err != nil {
		zap.L().Error("Could not unregister ws client", zap.Error(err), zap.String("id", c.ID))
	}
	err = c.Socket.Close()
	if err != nil {
		zap.L().Error("Could not unregister ws client", zap.Error(err), zap.String("id", c.ID))
	}
}
