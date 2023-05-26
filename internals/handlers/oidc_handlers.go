package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	oidcAuth "github.com/myrteametrics/myrtea-engine-api/v5/internals/router/oidc"
	"github.com/spf13/viper"
)

func HandleOIDCRedirect(w http.ResponseWriter, r *http.Request) {
	// Generate a random state to prevent CSRF attacks
	expectedState, err := generateEncryptedState([]byte(viper.GetString("OIDC_ENCRYPTION_KEY")))
	if err != nil {
		handleError(w, r, ExpectedStateErr, err, render.ErrAPIGenerateRandomStateFailed)
		return
	}
	render.Redirect(w, r, oidcAuth.OidcConfig.AuthCodeURL(expectedState), http.StatusFound)
}

func HandleOIDCCallback(w http.ResponseWriter, r *http.Request) {

	//check if state is the state expected
	_, err := verifyEncryptedState(r.URL.Query().Get("state"), []byte(viper.GetString("OIDC_ENCRYPTION_KEY")))
	if err != nil {
		handleError(w, r, InvalidStateErr, errors.New(InvalidStateErr), render.ErrAPIInvalidOIDCState)
		return
	}

	oauth2Token, err := oidcAuth.OidcConfig.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		handleError(w, r, TokenExchangeErr, err, render.ErrAPIExchangeOIDCTokenFailed)
		return
	}

	// Generate the token and add it to a cookie
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		handleError(w, r, NoIDTokenErr, err, render.ErrAPINoIDOIDCToken)
		return
	}

	_, err = oidcAuth.Provider.Verifier(&oidc.Config{ClientID: oidcAuth.OidcConfig.ClientID}).Verify(r.Context(), rawIDToken)

	if err != nil {
		handleError(w, r, IDTokenVerifyErr, err, render.ErrAPIVerifyIDOIDCTokenFailed)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    rawIDToken,
		HttpOnly: true,
		Secure:   true,
		Domain:   viper.GetString("FRONT_END_DOMAIN"),
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteNoneMode,
	})

	render.Redirect(w, r, viper.GetString("FRONT_END_URL"), http.StatusFound)
}
