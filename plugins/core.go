package plugin

import (
	"net/http"
)

// MyrteaPlugin is a standard interface for any myrtea plugins
type MyrteaPlugin interface {
	Start() error
	Stop() error
	Handler() http.Handler
	HandlerPrefix() string
}
