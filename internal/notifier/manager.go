package notifier

import (
	"errors"
	"sync"
)

// ClientManager is the websocket client pool manager
// It is used only to manage the client pool, and read/write on a specific client using raw byte slice message
type ClientManager struct {
	mutex sync.RWMutex
	// Clients is the set of every registered client
	Clients map[Client]bool
	// clientsByLogin indexes clients by user login to avoid a full pool scan on every send
	clientsByLogin map[string]map[Client]bool
}

// NewClientManager renders a new manager responsible for every connection
func NewClientManager() *ClientManager {
	return &ClientManager{
		Clients:        make(map[Client]bool),
		clientsByLogin: make(map[string]map[Client]bool),
	}
}

// GetClients returns all clients of the manager
func (manager *ClientManager) GetClients() []Client {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	clients := make([]Client, 0, len(manager.Clients))
	for k := range manager.Clients {
		clients = append(clients, k)
	}
	return clients
}

// GetClientsByLogin returns every client currently registered for the given user login
func (manager *ClientManager) GetClientsByLogin(login string) []Client {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	byLogin, ok := manager.clientsByLogin[login]
	if !ok {
		return nil
	}
	clients := make([]Client, 0, len(byLogin))
	for client := range byLogin {
		clients = append(clients, client)
	}
	return clients
}

// clientLogin returns the login of the client's user, or an empty string if none
func clientLogin(client Client) string {
	if user := client.GetUser(); user != nil {
		return user.Login
	}
	return ""
}

// Register registers a new client
func (manager *ClientManager) Register(newClient Client) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if manager.Clients[newClient] {
		return errors.New("this client already exists")
	}
	manager.Clients[newClient] = true

	login := clientLogin(newClient)
	if manager.clientsByLogin[login] == nil {
		manager.clientsByLogin[login] = make(map[Client]bool)
	}
	manager.clientsByLogin[login][newClient] = true
	return nil
}

// Unregister unregisters a client
func (manager *ClientManager) Unregister(existentClient Client) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if !manager.Clients[existentClient] {
		return errors.New("the client doesn't exist")
	}
	delete(manager.Clients, existentClient)

	login := clientLogin(existentClient)
	if byLogin, ok := manager.clientsByLogin[login]; ok {
		delete(byLogin, existentClient)
		if len(byLogin) == 0 {
			delete(manager.clientsByLogin, login)
		}
	}
	return nil
}
