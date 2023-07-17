package handlers

import (
	"net/http"
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
