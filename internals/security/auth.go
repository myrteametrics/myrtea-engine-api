package security

import "github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"

// Auth refers to a generic interface which must be implemented by every authentication backend
type Auth interface {
	Authenticate(string, string) (users.User, bool, error)
}
