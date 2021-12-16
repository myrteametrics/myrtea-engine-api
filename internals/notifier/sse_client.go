package notifier

import (
	"fmt"
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/users"
	uuid "github.com/satori/go.uuid"
)

// SSEClient represents a single specific server-sent event connection
type SSEClient struct {
	GenericClient
	w http.ResponseWriter
}

// BuildSSEClient build and returns a new SSEClient
func BuildSSEClient(w http.ResponseWriter, user *users.UserWithPermissions) (*SSEClient, error) {
	return &SSEClient{
		GenericClient: GenericClient{
			ID:   uuid.NewV4().String(),
			User: user,
			Send: make(chan []byte),
		},
		w: w,
	}, nil
}

func (c *SSEClient) Write() {
	flusher, ok := c.w.(http.Flusher)
	if !ok {
		http.Error(c.w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	for {
		// SSE compatible format for javascript EventSource() ("data: <content>\n\n")
		fmt.Fprintf(c.w, "data: %s\n\n", <-c.Send)
		flusher.Flush()
	}
}

func (c *SSEClient) Read() {}
