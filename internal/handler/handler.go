package handler

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"
)

// IsAlive godoc
//
//	@Id				IsAlive
//
//	@Summary		Check if alive
//	@Description	allows to check if the API is alive
//	@Tags			System
//	@Success		200	"Status OK"
//	@Router			/isalive [get]
func IsAlive(w http.ResponseWriter, r *http.Request) {
	httputil.JSON(w, r, map[string]interface{}{"alive": true})
}

// NotImplemented returns a basic message "Not Implemented" when called, and should be use a filler for future handler
func NotImplemented(w http.ResponseWriter, r *http.Request) {
	httputil.NotImplemented(w, r)
}
