package plugin

import (
	"net/http"
)

// MyrteaPlugin is a standard interface for any myrtea plugins
type MyrteaPlugin interface {
	ServicePort() int
	HandlerPrefix() string
	Handler() http.Handler
	Start() error
	Stop() error
}
