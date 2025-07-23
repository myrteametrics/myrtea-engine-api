package handler

import (
	routerauth "github.com/myrteametrics/myrtea-engine-api/v5/internal/router/auth"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"

	"go.uber.org/zap"
)

// LogoutHandler godoc
//
//	@Id				LogoutHandler
//
//	@Summary		Logout
//	@Description	Logs out the current user.
//	@Tags			Admin
//	@Produce		plain
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{string}	string	"Logged out successfully."
//	@Failure		500	{string}	string	"Internal Server Error"
//	@Router			/engine/logout [post]
func LogoutHandler(deleteSessionMiddleware func(http.Handler) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		successHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Logged out successfully."))
		})

		handler := deleteSessionMiddleware(successHandler)

		handler.ServeHTTP(w, r)
	})
}

// GetAuthenticationMode godoc
//
//	@Id				GetAuthenticationMode
//
//	@Summary		Get the current authentication mode
//	@Description	Retrieves the current mode used for authentication.
//	@Tags			Admin
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	routerauth.AuthenticationMode
//	@Failure		400	{string}	string	"Bad Request"
//	@Failure		500	{string}	string	"Internal Server Error"
//	@Router			/engine/authmode [get]
func GetAuthenticationMode(w http.ResponseWriter, r *http.Request) {
	mode, err := routerauth.GetMode()
	if err != nil {
		zap.L().Error("Error querying authentication mode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	httputil.JSON(w, r, mode)
}
