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
		Codec:    CustomSessionCodec(opts, sessionMaxAge),
		// Codec: samlsp.DefaultSessionCodec(opts),
	}
}

// CustomSessionCodec returns the custom SessionCodec for the provided options,
// a JWTSessionCodec configured to issue signed tokens.
func CustomSessionCodec(opts samlsp.Options, sessionMaxAge time.Duration) samlsp.JWTSessionCodec {
	codec := samlsp.DefaultSessionCodec(opts)
	codec.MaxAge = sessionMaxAge
	return codec
}
