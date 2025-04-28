package handlers

import (
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/handlers/render"
)

// IsAlive godoc
//
//	@Summary		Check if alive
//	@Description	allows to check if the API is alive
//	@Tags			System
//	@Success		200	"Status OK"
//	@Router			/isalive [get]
func IsAlive(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]interface{}{"alive": true})
}

// NotImplemented returns a basic message "Not Implemented" when called, and should be use a filler for future handler
func NotImplemented(w http.ResponseWriter, r *http.Request) {
	render.NotImplemented(w, r)
}
