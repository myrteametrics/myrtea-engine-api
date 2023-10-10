package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	oidcAuth "github.com/myrteametrics/myrtea-engine-api/v5/internals/router/oidc"
	"github.com/spf13/viper"
)

func HandleOIDCRedirect(w http.ResponseWriter, r *http.Request) {
	// Generate a random state to prevent CSRF attacks
	expectedState, err := generateEncryptedState([]byte(viper.GetString("AUTHENTICATION_OIDC_ENCRYPTION_KEY")))
	if err != nil {
		handleError(w, r, ExpectedStateErr, err, render.ErrAPIGenerateRandomStateFailed)
		return
	}
	instanceOidc, err := oidcAuth.GetOidcInstance()
	if err != nil {
		handleError(w, r, "", err, render.ErrAPIProcessError)
		return
	}
	render.Redirect(w, r, instanceOidc.OidcConfig.AuthCodeURL(expectedState), http.StatusFound)
}

func HandleOIDCCallback(w http.ResponseWriter, r *http.Request) {

	//check if state is the state expected
	_, err := verifyEncryptedState(r.URL.Query().Get("state"), []byte(viper.GetString("AUTHENTICATION_OIDC_ENCRYPTION_KEY")))
	if err != nil {
		handleError(w, r, InvalidStateErr, errors.New(InvalidStateErr), render.ErrAPIInvalidOIDCState)
		return
	}

	instanceOidc, err := oidcAuth.GetOidcInstance()
	if err != nil {
		handleError(w, r, "", err, render.ErrAPIProcessError)
		return
	}

	oauth2Token, err := instanceOidc.OidcConfig.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		handleError(w, r, TokenExchangeErr, err, render.ErrAPIExchangeOIDCTokenFailed)
		return
	}

	// Generate the token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		handleError(w, r, NoIDTokenErr, err, render.ErrAPINoIDOIDCToken)
		return
	}

	_, err = instanceOidc.Provider.Verifier(&oidc.Config{ClientID: instanceOidc.OidcConfig.ClientID}).Verify(r.Context(), rawIDToken)

	if err != nil {
		handleError(w, r, IDTokenVerifyErr, err, render.ErrAPIVerifyIDOIDCTokenFailed)
		return
	}

	baseURL := viper.GetString("AUTHENTICATION_OIDC_FRONT_END_URL")
	redirectURL := fmt.Sprintf("%s/auth/oidc/callback?token=%s", baseURL, url.QueryEscape(rawIDToken))
	render.Redirect(w, r, redirectURL, http.StatusFound)
}
