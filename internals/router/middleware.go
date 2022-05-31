package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	gorillacontext "github.com/gorilla/context"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/users"
	"go.uber.org/zap"
)

// UnverifiedAuthenticator doc
// WARNING: Don't use this method unless you know what you're doing
// This method parses the token but doesn't validate the signature. It's only
// ever useful in cases where you know the signature is valid (because it has
// been checked previously in the stack) and you want to extract values from
// it.
func UnverifiedAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		findTokenFns := []func(r *http.Request) string{jwtauth.TokenFromQuery, jwtauth.TokenFromHeader, jwtauth.TokenFromCookie}
		var tokenStr string
		for _, fn := range findTokenFns {
			if tokenStr = fn(r); tokenStr != "" {
				break
			}
		}
		if tokenStr == "" {
			zap.L().Warn("No JWT string found in request")
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("Missing JWT"))
			return
		}

		token, err := jwt.Parse([]byte(tokenStr))
		if err != nil {
			zap.L().Warn("JWT string cannot be parsed") // , zap.String("jwt", tokenStr)) // Security issue if logged without check ?
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("Invalid JWT"))
			return
		}

		ctx = jwtauth.NewContext(ctx, token, err)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ContextMiddleware :
func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil {
			zap.L().Warn("Get JWT infos from context", zap.Error(err))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("Invalid JWT"))
			return
		}

		rawUserID := claims["id"]
		if rawUserID == nil {
			zap.L().Warn("Token found without user ID", zap.Any("claims", claims))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("Invalid JWT"))
			return
		}
		userID := int64(rawUserID.(float64))

		user, found, err := users.R().Get(userID)
		if err != nil {
			zap.L().Error("Cannot get user", zap.Error(err))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("Internal Error"))
			return
		}
		if !found {
			zap.L().Error("User not found", zap.Int64("id", userID))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("Invalid JWT"))
			return
		}

		gr, err := groups.R().GetGroupsOfUser(user.ID)
		if err != nil {
			zap.L().Error("Find Groups", zap.Error(err))
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("Internal Error"))
			return
		}

		ug := groups.UserWithGroups{
			User:   user,
			Groups: gr,
		}

		loggerR := r.Context().Value(models.ContextKeyLoggerR)
		if loggerR != nil {
			gorillacontext.Set(loggerR.(*http.Request), models.UserLogin, fmt.Sprintf("%s(%d)", ug.User.Login, ug.User.ID))
		}

		ctx := context.WithValue(r.Context(), models.ContextKeyUser, ug)
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
