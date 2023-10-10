package handlers

import (
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/router/authmanagement"
	"go.uber.org/zap"
)

// LogoutHandler prend une fonction middleware comme argument
// Cette fonction middleware est responsable de supprimer la session
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
// @Summary Get the current authentication mode
// @Description Retrieves the current mode used for authentication.
// @Tags Admin
// @Produce json
// @Security Bearer
// @Success 200 {object} AuthenticationMode
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /engine/authmode [get]
func GetAuthenticationMode(w http.ResponseWriter, r *http.Request) {
	mode, err := authmanagement.GetMode()
	if err != nil {
		zap.L().Error("Error querying authentication mode", zap.Error(err))
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	render.JSON(w, r, mode)
}
