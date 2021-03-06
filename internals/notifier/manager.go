package notifier

import (
	"errors"
	"sync"
)

// ClientManager is the websocket client pool manager
// It is used only to manage the client pool, and read/write on a specific client using raw byte slice message
type ClientManager struct {
	mutex   sync.RWMutex
	Clients map[Client]bool
}

// NewClientManager renders a new manager responsible of every connection
func NewClientManager() *ClientManager {
	return &ClientManager{
		Clients: make(map[Client]bool),
	}
}

// GetClients returns all clients of the manager
func (manager *ClientManager) GetClients() []Client {
	manager.mutex.RLock()
	clients := make([]Client, 0, len(manager.Clients))
	for k := range manager.Clients {
		clients = append(clients, k)
	}
	manager.mutex.RUnlock()
	return clients
}

// Register registers a new client
func (manager *ClientManager) Register(newClient Client) error {
	manager.mutex.Lock()
	if val, ok := manager.Clients[newClient]; ok {
		if val {
			return errors.New("This client already exists")
		}
	}
	manager.Clients[newClient] = true
	manager.mutex.Unlock()
	return nil
}

// Unregister unregisters a client
func (manager *ClientManager) Unregister(existentClient Client) error {
	manager.mutex.Lock()
	if val, ok := manager.Clients[existentClient]; !ok {
		if val {
			return errors.New("The client doesn't exist")
		}
	}
	delete(manager.Clients, existentClient)
	manager.mutex.Unlock()
	return nil
}
