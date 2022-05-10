package router

import (
	"time"

	"github.com/crewjam/saml/samlsp"
)

const (
	defaultSessionCookieName = "token"
	defaultSessionMaxAge     = time.Hour
)

// CustomSessionProvider returns the default SessionProvider for the provided options,
// a CookieSessionProvider configured to store sessions in a cookie.
func CustomSessionProvider(opts samlsp.Options, cookieMaxAge time.Duration) samlsp.CookieSessionProvider {
	sessionMaxAge := defaultSessionMaxAge
	if cookieMaxAge > 0 {
		sessionMaxAge = cookieMaxAge
	}
	return samlsp.CookieSessionProvider{
		Name:     defaultSessionCookieName,
		Domain:   opts.URL.Host,
		MaxAge:   sessionMaxAge,
		HTTPOnly: true,
		Secure:   opts.URL.Scheme == "https",
		SameSite: opts.CookieSameSite,
		Codec:    samlsp.DefaultSessionCodec(opts),
	}
}
