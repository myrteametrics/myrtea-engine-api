package router

import (
	"context"
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/security/apikey"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	roles2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/roles"
	users2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	ttlcache "github.com/myrteametrics/myrtea-sdk/v5/cache"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	gorillacontext "github.com/gorilla/context"
	"go.uber.org/zap"
)

// ContextMiddleware :
func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			zap.L().Warn("Get JWT infos from context", zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, errors.New("invalid JWT"))
			return
		}

		rawUserID := claims["id"]
		if rawUserID == nil {
			zap.L().Warn("Token found without user ID", zap.Any("claims", claims))
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, errors.New("invalid JWT"))
			return
		}

		rawUserIDStr, ok := rawUserID.(string)
		if !ok {
			zap.L().Warn("Cannot parse user ID", zap.Any("claims", claims))
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, errors.New("invalid JWT"))
			return
		}

		userID, _ := uuid.Parse(rawUserIDStr)

		user, found, err := users2.R().Get(userID)
		if err != nil {
			zap.L().Error("Cannot get user", zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, errors.New("internal Error"))
			return
		}
		if !found {
			zap.L().Error("User not found", zap.String("userID", rawUserID.(string)))
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, errors.New("invalid JWT"))
			return
		}

		roles, err := roles2.R().GetAllForUser(user.ID)
		if err != nil {
			zap.L().Error("Find Roles", zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, errors.New("internal Error"))
			return
		}

		userRoleUUIDs := make([]uuid.UUID, 0)
		for _, userRole := range roles {
			userRoleUUIDs = append(userRoleUUIDs, userRole.ID)
		}

		userPermissions, err := permissions.R().GetAllForRoles(userRoleUUIDs)
		if err != nil {
			zap.L().Error("Cannot get permissions", zap.Error(err))
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, errors.New("internal Error"))
			return
		}

		up := users2.UserWithPermissions{
			User:        user,
			Roles:       roles,
			Permissions: userPermissions,
		}

		loggerR := r.Context().Value(httputil.ContextKeyLoggerR)
		if loggerR != nil {
			gorillacontext.Set(loggerR.(*http.Request), httputil.UserLogin, fmt.Sprintf("%s(%s)", up.User.Login, up.User.ID))
		}

		ctx := context.WithValue(r.Context(), httputil.ContextKeyUser, up)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CustomAuthenticator is a default authentication middleware to enforce access from the
// Verifier middleware request context values. The Authenticator sends a 401 Unauthorized
// response for any unverified tokens and passes the good ones through. It's just fine
// until you decide to write something similar and customize your client response.
func CustomAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())

		if err != nil {
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, nil)
			return
		}

		// if token == nil || !token.Valid {
		if token == nil {
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, nil)
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}

// ContextMiddlewareApiKey :
func ContextMiddlewareApiKey(next http.Handler, Cache *ttlcache.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keyValue := r.Header.Get(HeaderKeyApiKey)

		if keyValue == "" {
			zap.L().Debug("API key missing from request")
			http.Error(w, "API key required (X-API-Key header)", http.StatusUnauthorized)
			return
		}

		var authKey apikey.APIKey

		Key, ok := Cache.Get(keyValue)
		authKey, isValid := Key.(apikey.APIKey)

		if !ok || !isValid {
			authKey, found, err := apikey.R().Validate(keyValue)
			if err != nil {
				zap.L().Error("API key validation error", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			if !found {
				zap.L().Debug("API key not found", zap.String("key", keyValue))
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			Cache.Set(keyValue, authKey)
		}

		role, found, err := roles2.R().Get(authKey.RoleID)
		if err != nil {
			zap.L().Error("Error retrieving role", zap.Error(err), zap.Any("roleID", authKey.RoleID))
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, errors.New("internal error"))
			return
		}
		if !found {
			zap.L().Debug("Role not found", zap.Any("roleID", authKey.RoleID))
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, errors.New("role not found"))
			return
		}

		userPermissions, err := permissions.R().GetAllForRole(role.ID)
		if err != nil {
			zap.L().Error("Error retrieving permissions", zap.Error(err), zap.Any("roleID", role.ID))
			httputil.Error(w, r, httputil.ErrAPISecurityMissingContext, errors.New("internal error"))
			return
		}

		up := users2.UserWithPermissions{
			User: users2.User{
				ID:        authKey.ID,
				Login:     "apikey-" + authKey.Name,
				Created:   authKey.CreatedAt,
				LastName:  "API Service",
				FirstName: authKey.Name,
			},
			Roles:       []roles2.Role{role},
			Permissions: userPermissions,
		}

		loggerR := r.Context().Value(httputil.ContextKeyLoggerR)
		if loggerR != nil {
			gorillacontext.Set(loggerR.(*http.Request), httputil.UserLogin, fmt.Sprintf("%s(%s)", up.User.Login, up.User.ID))
		}

		zap.L().Debug("API key authentication successful",
			zap.String("apikey", authKey.Name),
			zap.String("user", up.User.Login))
		ctx := context.WithValue(r.Context(), httputil.ContextKeyUser, up)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
