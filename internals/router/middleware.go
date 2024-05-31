package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	gorillacontext "github.com/gorilla/context"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/roles"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"go.uber.org/zap"
)

// ContextMiddleware :
func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			zap.L().Warn("Get JWT infos from context", zap.Error(err))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("invalid JWT"))
			return
		}

		rawUserID := claims["id"]
		if rawUserID == nil {
			zap.L().Warn("Token found without user ID", zap.Any("claims", claims))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("invalid JWT"))
			return
		}

		rawUserIDStr, ok := rawUserID.(string)
		if !ok {
			zap.L().Warn("Cannot parse user ID", zap.Any("claims", claims))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("invalid JWT"))
			return
		}

		userID, _ := uuid.Parse(rawUserIDStr)

		user, found, err := users.R().Get(userID)
		if err != nil {
			zap.L().Error("Cannot get user", zap.Error(err))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("internal Error"))
			return
		}
		if !found {
			zap.L().Error("User not found", zap.String("userID", rawUserID.(string)))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("invalid JWT"))
			return
		}

		roles, err := roles.R().GetAllForUser(user.ID)
		if err != nil {
			zap.L().Error("Find Roles", zap.Error(err))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("internal Error"))
			return
		}

		userRoleUUIDs := make([]uuid.UUID, 0)
		for _, userRole := range roles {
			userRoleUUIDs = append(userRoleUUIDs, userRole.ID)
		}

		userPermissions, err := permissions.R().GetAllForRoles(userRoleUUIDs)
		if err != nil {
			zap.L().Error("Cannot get permissions", zap.Error(err))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("internal Error"))
			return
		}

		up := users.UserWithPermissions{
			User:        user,
			Roles:       roles,
			Permissions: userPermissions,
		}

		loggerR := r.Context().Value(models.ContextKeyLoggerR)
		if loggerR != nil {
			gorillacontext.Set(loggerR.(*http.Request), models.UserLogin, fmt.Sprintf("%s(%s)", up.User.Login, up.User.ID))
		}

		ctx := context.WithValue(r.Context(), models.ContextKeyUser, up)
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
			render.Error(w, r, render.ErrAPISecurityMissingContext, nil)
			return
		}

		// if token == nil || !token.Valid {
		if token == nil {
			render.Error(w, r, render.ErrAPISecurityMissingContext, nil)
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}
