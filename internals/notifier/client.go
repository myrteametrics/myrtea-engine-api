package notifier

import "github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"

// Client is used as an abstract notifier client, which could use SSE or WS implementations
type Client interface {
	GetUser() *groups.UserWithGroups
	GetSendChannel() chan []byte
	Read()
	Write()
}

// GenericClient is a standard notification client
type GenericClient struct {
	ID   string
	Send chan []byte
	User *groups.UserWithGroups
}

// GetSendChannel returns Send channel
func (c *GenericClient) GetSendChannel() chan []byte {
	return c.Send
}

// GetUser returns current client user
func (c *GenericClient) GetUser() *groups.UserWithGroups {
	return c.User
}
