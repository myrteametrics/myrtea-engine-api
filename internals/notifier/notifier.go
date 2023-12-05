package notifier

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/notifier/notification"
	"go.uber.org/zap"
)

var (
	_globalNotifierMu sync.RWMutex
	_globalNotifier   *Notifier
)

// C is used to access the global notifier singleton
func C() *Notifier {
	_globalNotifierMu.RLock()
	defer _globalNotifierMu.RUnlock()

	notifier := _globalNotifier
	return notifier
}

// ReplaceGlobals affect a new notifier to the global notifier singleton
func ReplaceGlobals(notifier *Notifier) func() {
	_globalNotifierMu.Lock()
	defer _globalNotifierMu.Unlock()

	prev := _globalNotifier
	_globalNotifier = notifier
	return func() { ReplaceGlobals(prev) }
}

// Notifier is the main struct used to send notifications
type Notifier struct {
	clientManager *ClientManager
	cache         map[string]time.Time
	// queue / cache / batch system
}

// NewNotifier returns a pointer to a new instance of Notifier
func NewNotifier() *Notifier {
	cm := NewClientManager()
	return &Notifier{
		clientManager: cm,
		cache:         make(map[string]time.Time),
	}
}

// Register add a new client to the client manager pool
func (notifier *Notifier) Register(client Client) error {
	zap.L().Info("Client registered", zap.Any("user", client.GetUser()))
	return notifier.clientManager.Register(client)
}

// Unregister disconnect an existing client from the client manager pool
func (notifier *Notifier) Unregister(client Client) error {
	zap.L().Info("Client unregistered", zap.Any("user", client.GetUser()))
	return notifier.clientManager.Unregister(client)
}

// verifyCache check if a notification has already been sent
func (notifier *Notifier) verifyCache(key string, timeout time.Duration) bool {
	if val, ok := notifier.cache[key]; ok && time.Now().UTC().Before(val) {
		return false
	}
	notifier.cache[key] = time.Now().UTC().Add(timeout)
	return true
}

func (notifier *Notifier) CleanCache() {
	for key, val := range notifier.cache {
		if time.Now().UTC().After(val) {
			delete(notifier.cache, key)
		}
	}
}

// TODO: renew this
//// SendToRoles send a notification to every user related to the input list of roles
//func (notifier *Notifier) SendToRoles(cacheKey string, timeout time.Duration, notif notification.Notification, roles []uuid.UUID) {
//
//	zap.L().Debug("notifier.SendToRoles", zap.Any("roles", roles), zap.Any("notification", notif))
//
//	if cacheKey != "" && !notifier.verifyCache(cacheKey, timeout) {
//		zap.L().Debug("Notification send skipped")
//		return
//	}
//
//	id, err := notification.R().Create(notif, "")
//	if err != nil {
//		zap.L().Error("Add notification to history", zap.Error(err))
//		return
//	}
//
//	notifFull, err := notification.R().Get(id)
//	if notifFull == nil {
//		zap.L().Error("Notification not found after creation", zap.Int64("id", id))
//	}
//
//	// FIXME: This should be fully reworking after security refactoring and removal of groups
//
//	// if roles != nil && len(roles) > 0 {
//	// 	clients := make(map[Client]bool, 0)
//	// 	for _, roleID := range roles {
//	// 		roleClients := notifier.findClientsByRoleID(roleID)
//	// 		for _, client := range roleClients {
//	// 			clients[client] = true
//	// 		}
//	// 	}
//	// 	for client := range clients {
//	// 		notifier.sendToClient(notifFull, client)
//	// 	}
//	// }
//}

// sendToClient convert and send a notification to a specific client
// Every multiplexing function must call this function in the end to send message
func (notifier *Notifier) sendToClient(notif notification.Notification, client Client) {
	message, err := notif.ToBytes()
	if err != nil {
		zap.L().Error("notif.ToBytes()", zap.Error(err))
		return
	}

	notifier.Send(message, client)
}

// Broadcast send a notification to every connected client
func (notifier *Notifier) Broadcast(notif notification.Notification) {
	for _, client := range notifier.clientManager.GetClients() {
		notifier.sendToClient(notif, client)
	}
}

// SendToUsers send a notification to users corresponding the input ids
func (notifier *Notifier) SendToUsers(notif notification.Notification, users []users.UserWithPermissions) {
	if users != nil && len(users) > 0 {
		for _, user := range users {
			clients := notifier.findClientsByUserLogin(user.Login)
			for _, client := range clients {
				notifier.sendToClient(notif, client)
			}
		}
	}
}

// SendToUser send a notification to a specific user
func (notifier *Notifier) SendToUser(notif notification.Notification, user users.UserWithPermissions) error {
	id, err := notification.R().Create(notif, user.Login)
	if err != nil {
		zap.L().Error("Add notification to history", zap.Error(err))
		return err
	}

	notif = notif.SetId(id)
	clients := notifier.findClientsByUserLogin(user.Login)
	for _, client := range clients {
		notifier.sendToClient(notif, client)
	}
	return nil
}

// Send a byte slices to a specific websocket client
func (notifier *Notifier) Send(message []byte, client Client) {
	if client != nil {
		client.GetSendChannel() <- message
	}
}

// findClientsByUserLogin returns a list of clients corresponding to the input login
func (notifier *Notifier) findClientsByUserLogin(login string) []Client {
	clients := make([]Client, 0)
	for _, client := range notifier.clientManager.GetClients() {
		if client.GetUser() != nil && client.GetUser().Login == login {
			clients = append(clients, client)
		}
	}
	return clients
}

// findClientsByRoleID returns a list of clients corresponding to the input role id
func (notifier *Notifier) findClientsByRoleID(id uuid.UUID) []Client {
	clients := make([]Client, 0)
	for _, client := range notifier.clientManager.GetClients() {
		if client.GetUser() != nil {
			for _, role := range client.GetUser().Roles {
				if role.ID == id {
					clients = append(clients, client)
				}
			}
		}
	}
	return clients
}
