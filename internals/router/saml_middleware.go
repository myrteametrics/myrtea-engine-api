package router

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/google/uuid"
	gorillacontext "github.com/gorilla/context"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/roles"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/users"
	"go.uber.org/zap"
)

// SamlSPMiddleware wraps default samlsp.Middleware and override some specific func
type SamlSPMiddleware struct {
	*samlsp.Middleware
	Config SamlSPMiddlewareConfig
}

// NewSamlSP build a new SAML Service Provider middleware
func NewSamlSP(spRootURLStr string, entityID string, keyFile string, crtFile string, config SamlSPMiddlewareConfig) (*SamlSPMiddleware, error) {
	spRootURL, err := url.Parse(spRootURLStr)
	if err != nil {
		zap.L().Panic("Parse SP root URL", zap.Error(err))
		return nil, err
	}

	keyPair, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		zap.L().Panic("Load X509 Key Pair", zap.Error(err), zap.String("crtFile", crtFile), zap.String("keyFile", keyFile))
		return nil, err
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		zap.L().Panic("X509 ParseCertificate", zap.Error(err))
		return nil, err
	}

	metadata, err := GetIDPMetadata(config.MetadataMode, config.MetadataFilePath, config.MetadataURL)
	if err != nil {
		zap.L().Panic("GetIDPMetadata", zap.Error(err))
		return nil, err
	}

	onError := samlsp.DefaultOnError
	opts := samlsp.Options{
		URL:         *spRootURL,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: metadata,
		EntityID:    entityID,
		SignRequest: true,
		ForceAuthn:  false,
	}

	defaultMiddleware := &samlsp.Middleware{
		ServiceProvider: CustomServiceProvider(opts),
		Binding:         "",
		OnError:         onError,
		Session:         CustomSessionProvider(opts, config.CookieMaxAge),
	}
	defaultMiddleware.RequestTracker = CustomRequestTracker(opts, &defaultMiddleware.ServiceProvider)

	return &SamlSPMiddleware{
		Middleware: defaultMiddleware,
		Config:     config,
	}, nil
}

// RequireAccount is a HTTP middleware that requires that each request is
// associated with a valid session. If the request is not associated with a valid
// session, then rather than serve the request, the middleware redirects the user
// to start the SAML authentication flow.
func (m *SamlSPMiddleware) RequireAccount(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.Session.GetSession(r)
		if session != nil {
			r = r.WithContext(samlsp.ContextWithSession(r.Context(), session))

			handler.ServeHTTP(w, r)
			return
		}
		if err == samlsp.ErrNoSession {
			// SamlSPMiddleware custom function instead of DefaultMiddleware
			m.HandleStartAuthFlow(w, r)
			return
		}

		m.OnError(w, r, err)
	})
}

// HandleStartAuthFlow is called to start the SAML authentication process.
func (m *SamlSPMiddleware) HandleStartAuthFlow(w http.ResponseWriter, r *http.Request) {
	// If we try to redirect when the original request is the ACS URL we'll
	// end up in a loop. This is a programming error, so we panic here. In
	// general this means a 500 to the user, which is preferable to a
	// redirect loop.
	if r.URL.Path == m.ServiceProvider.AcsURL.Path {
		panic("don't wrap Middleware with RequireAccount")
	}

	var binding, bindingLocation string
	if m.Binding != "" {
		binding = m.Binding
		bindingLocation = m.ServiceProvider.GetSSOBindingLocation(binding)
	} else {
		binding = saml.HTTPRedirectBinding
		bindingLocation = m.ServiceProvider.GetSSOBindingLocation(binding)
		if bindingLocation == "" {
			binding = saml.HTTPPostBinding
			bindingLocation = m.ServiceProvider.GetSSOBindingLocation(binding)
		}
	}

	authReq, err := m.ServiceProvider.MakeAuthenticationRequest(bindingLocation, binding, m.ResponseBinding)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// relayState is limited to 80 bytes but also must be integrity protected.
	// this means that we cannot use a JWT because it is way to long. Instead
	// we set a signed cookie that encodes the original URL which we'll check
	// against the SAML response when we get it.
	relayState, err := m.RequestTracker.TrackRequest(w, r, authReq.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redirectURL, err := authReq.Redirect(relayState, &m.ServiceProvider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Authenticate-To", redirectURL.String())
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
}

// ContextMiddleware extracts a session from the request context and adds (if possible) a new user
// in the request context for further usage in the APIs
func (m *SamlSPMiddleware) ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := samlsp.SessionFromContext(r.Context()).(samlsp.SessionWithAttributes)

		zap.L().Debug("session", zap.Any("session", session))

		userID := samlsp.AttributeFromContext(r.Context(), m.Config.AttributeUserID)
		if userID == "" {
			zap.L().Error("Missing userID from session after SAML authentication")
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("invalid Session"))
			return
		}

		userDisplayName := samlsp.AttributeFromContext(r.Context(), m.Config.AttributeUserDisplayName)

		userGroups := session.GetAttributes()[m.Config.AttributeUserMemberOf] // .Get(<name>) doesn't work with slice
		userGroups = sliceDeduplicate(userGroups)
		userRoles := make([]roles.Role, 0)
		for _, userGroupName := range userGroups {
			role, found, err := roles.R().GetByName(userGroupName)
			if err != nil {
				zap.L().Error("Cannot get roles", zap.Error(err), zap.String("groupName", userGroupName))
				render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("internal Error"))
				return
			}
			if !found {
				continue
			}

			userRoles = append(userRoles, role)
		}

		userRoleUUIDs := make([]uuid.UUID, 0)
		for _, userRole := range userRoles {
			userRoleUUIDs = append(userRoleUUIDs, userRole.ID)
		}

		userPermissions, err := permissions.R().GetAllForRoles(userRoleUUIDs)
		if err != nil {
			zap.L().Error("Cannot get permissions", zap.Error(err))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("internal Error"))
			return
		}

		if m.Config.EnableMemberOfValidation && len(userRoles) == 0 {
			zap.L().Warn("User access denied", zap.String("reason", "no valid group found"), zap.String("userID", userID), zap.Strings("userGroups", userGroups))
			render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("access denied"))
			return
		}

		up := users.UserWithPermissions{
			User: users.User{
				Login:    userID,
				LastName: userDisplayName,
			},
			Roles:       userRoles,
			Permissions: userPermissions,
		}

		loggerR := r.Context().Value(models.ContextKeyLoggerR)
		if loggerR != nil {
			gorillacontext.Set(loggerR.(*http.Request), models.UserLogin, fmt.Sprintf("%s(%d)", up.User.Login, up.User.ID))
		}

		ctx := context.WithValue(r.Context(), models.ContextKeyUser, up)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminAuthentificator is a middle which check if the user is administrator (* * *)
func (m *SamlSPMiddleware) AdminAuthentificator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, found := handlers.GetUserFromContext(r)
		if !found {
			http.Error(w, http.StatusText(401), 403)
			return
		}
		if !user.HasPermission(permissions.New(permissions.All, permissions.All, permissions.All)) {
			http.Error(w, http.StatusText(403), 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}
