package notifier

import (
	"github.com/google/uuid"
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
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
	defer func() {
		destroyWebsocketClient(c)
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Socket.WriteMessage(websocket.TextMessage, message)
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
			switch err.(type) {
			case *websocket.CloseError:
				e := err.(*websocket.CloseError)
				if e.Code != websocket.CloseNormalClosure && e.Code != websocket.CloseGoingAway {
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
	C().Unregister(c)
	c.Socket.Close()
}
