package notification

//FrontNotification data structure represente the notification and her current state
type FrontNotification struct {
	Notification
	IsRead bool
}

// Notification is a general interface for all notifications types
type Notification interface {
	ToBytes() ([]byte, error)
}
