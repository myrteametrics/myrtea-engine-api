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
	gorillacontext "github.com/gorilla/context"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-sdk/v4/security"
	"github.com/spf13/viper"
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

	// for backwards compatibility, support IDPMetadataURL
	if opts.IDPMetadataURL != nil && opts.IDPMetadata == nil {
		httpClient := opts.HTTPClient
		if httpClient == nil {
			httpClient = http.DefaultClient
		}
		metadata, err := samlsp.FetchMetadata(context.TODO(), httpClient, *opts.IDPMetadataURL)
		if err != nil {
			return nil, err
		}
		opts.IDPMetadata = metadata
	}

	defaultMiddleware := &samlsp.Middleware{
		ServiceProvider: CustomServiceProvider(opts),
		Binding:         "",
		OnError:         onError,
		Session:         samlsp.DefaultSessionProvider(opts),
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

	authReq, err := m.ServiceProvider.MakeAuthenticationRequest(bindingLocation)
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

	redirectURL := authReq.Redirect(relayState)
	w.Header().Add("Authenticate-To", redirectURL.String())
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
}

// ContextMiddleware extracts a session from the request context and adds (if possible) a new user
// in the request context for further usage in the APIs
func (m *SamlSPMiddleware) ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := samlsp.SessionFromContext(r.Context()).(samlsp.SessionWithAttributes)

		userID := samlsp.AttributeFromContext(r.Context(), m.Config.AttributeUserID)
		if userID == "" {
			zap.L().Error("Missing userID from session after SAML authentication")
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("Invalid Session"))
			return
		}

		userDisplayName := samlsp.AttributeFromContext(r.Context(), m.Config.AttributeUserDisplayName)

		userGroups := session.GetAttributes()[m.Config.AttributeUserMemberOf] // .Get(<name>) doesn't work with slice
		userGroups = sliceDeduplicate(userGroups)
		userMemberOf := make([]groups.GroupOfUser, 0)
		for _, userGroupName := range userGroups {
			g, found, err := groups.R().GetByName(userGroupName)
			if err != nil {
				zap.L().Error("Cannot get group", zap.Error(err), zap.String("groupName", userGroupName))
				render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("Internal Error"))
				return
			}
			if !found {
				continue
			}

			userMemberOf = append(userMemberOf, groups.GroupOfUser{
				ID:       g.ID,
				Name:     g.Name,
				UserRole: 1,
			})
		}

		if m.Config.EnableMemberOfValidation && len(userMemberOf) == 0 {
			zap.L().Warn("User access denied", zap.String("reason", "no valid group found"), zap.String("userID", userID))
			render.Error(w, r, render.ErrAPISecurityNoRights, errors.New("Access denied"))
			return
		}

		isAdmin := 0
		for member := range userMemberOf {
			if member.Name == viper.GetString("AUTHENTICATION_SAML_ADMIN_GROUP_NAME") {
				isAdmin = 1
			}
		}

		ug := groups.UserWithGroups{
			User: security.User{
				Login:    userID,
				LastName: userDisplayName,
				Role:     isAdmin,
			},
			Groups: userMemberOf,
		}

		loggerR := r.Context().Value(models.ContextKeyLoggerR)
		if loggerR != nil {
			gorillacontext.Set(loggerR.(*http.Request), models.UserLogin, fmt.Sprintf("%s(%d)", ug.User.Login, ug.User.ID))
		}

		ctx := context.WithValue(r.Context(), models.ContextKeyUser, ug)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
