package notifier

// import "github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
import "github.com/myrteametrics/myrtea-engine-api/v4/internals/security/users"

// Client is used as an abstract notifier client, which could use SSE or WS implementations
type Client interface {
	GetUser() *users.UserWithPermissions
	GetSendChannel() chan []byte
	Read()
	Write()
}

// GenericClient is a standard notification client
type GenericClient struct {
	ID   string
	Send chan []byte
	User *users.UserWithPermissions
}

// GetSendChannel returns Send channel
func (c *GenericClient) GetSendChannel() chan []byte {
	return c.Send
}

// GetUser returns current client user
func (c *GenericClient) GetUser() *users.UserWithPermissions {
	return c.User
}
