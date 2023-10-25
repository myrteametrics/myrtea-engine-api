package oidcAuth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	gorillacontext "github.com/gorilla/context"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/roles"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/utils"
	"go.uber.org/zap"
)

func OIDCMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		var tokenStr string
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			zap.L().Warn("No token string found in request")
			render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("missing token"))
			return
		}

		// Check the token with the OIDC server
		instanceOidc, err := GetOidcInstance()
		if err != nil {
			zap.L().Error("", zap.Error(err))
			render.Error(w, r, render.ErrAPIProcessError, err)
			return
		}
		idToken, err := instanceOidc.Provider.Verifier(&oidc.Config{ClientID: instanceOidc.OidcConfig.ClientID}).Verify(r.Context(), tokenStr)
		if err != nil {
			zap.L().Error("Invalid OIDC auth Token", zap.Error(err))
			render.Error(w, r, render.ErrAPIInvalidAuthToken, err)
			return
		}

		// Check if the token has expired
		if idToken.Expiry.Before(time.Now()) {
			zap.L().Error("OIDC auth Token expired")
			render.Error(w, r, render.ErrAPIExpiredAuthToken, errors.New("expired auth token"))
			return
		}

		// Inject the id token into the context for the following handlers
		ctx := context.WithValue(r.Context(), "idToken", idToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the id token in the context
		idToken, ok := r.Context().Value("idToken").(*oidc.IDToken)
		if !ok {
			zap.L().Error("OIDC auth Missing idToken from context")
			render.Error(w, r, render.ErrAPIMissingIDTokenFromContext, errors.New("Missing idToken from context"))
			return
		}

		// Get the user's claims
		var claims struct {
			Email         string   `json:"email"`
			VerifiedEmail bool     `json:"email_verified"`
			Name          string   `json:"name"`
			GivenName     string   `json:"given_name"`
			FamilyName    string   `json:"family_name"`
			Roles         []string `json:"roles"`
		}

		if err := idToken.Claims(&claims); err != nil {
			zap.L().Error("OIDC failed to get User clams", zap.Error(err))
			render.Error(w, r, render.ErrAPIFailedToGetUserClaims, err)
			return
		}

		// inject the user information (Permission role etc) Connect into the context.
		// and for now assuming all are admin
		userGroups := claims.Roles
		userGroups = utils.RemoveDuplicates(userGroups)
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

		up := users.UserWithPermissions{
			User: users.User{
				Login:    claims.Email,
				LastName: claims.GivenName,
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
