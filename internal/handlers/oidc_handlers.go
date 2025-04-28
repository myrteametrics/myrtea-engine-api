package handlers

import (
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	oidcAuth "github.com/myrteametrics/myrtea-engine-api/v5/internal/router/oidc"
	"github.com/spf13/viper"
)

func HandleOIDCRedirect(w http.ResponseWriter, r *http.Request) {
	// Generate a random state to prevent CSRF attacks
	expectedState, err := generateEncryptedState([]byte(viper.GetString("AUTHENTICATION_OIDC_ENCRYPTION_KEY")))
	if err != nil {
		handleError(w, r, ExpectedStateErr, err, httputil.ErrAPIGenerateRandomStateFailed)
		return
	}
	instanceOidc, err := oidcAuth.GetOidcInstance()
	if err != nil {
		handleError(w, r, "", err, httputil.ErrAPIProcessError)
		return
	}
	http.Redirect(w, r, instanceOidc.OidcConfig.AuthCodeURL(expectedState), http.StatusFound)
}

func HandleOIDCCallback(w http.ResponseWriter, r *http.Request) {

	//check if state is the state expected
	_, err := verifyEncryptedState(r.URL.Query().Get("state"), []byte(viper.GetString("AUTHENTICATION_OIDC_ENCRYPTION_KEY")))
	if err != nil {
		handleError(w, r, InvalidStateErr, errors.New(InvalidStateErr), httputil.ErrAPIInvalidOIDCState)
		return
	}

	instanceOidc, err := oidcAuth.GetOidcInstance()
	if err != nil {
		handleError(w, r, "", err, httputil.ErrAPIProcessError)
		return
	}

	oauth2Token, err := instanceOidc.OidcConfig.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		handleError(w, r, TokenExchangeErr, err, httputil.ErrAPIExchangeOIDCTokenFailed)
		return
	}

	// Generate the token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		handleError(w, r, NoIDTokenErr, err, httputil.ErrAPINoIDOIDCToken)
		return
	}

	_, err = instanceOidc.Provider.Verifier(&oidc.Config{ClientID: instanceOidc.OidcConfig.ClientID}).Verify(r.Context(), rawIDToken)

	if err != nil {
		handleError(w, r, IDTokenVerifyErr, err, httputil.ErrAPIVerifyIDOIDCTokenFailed)
		return
	}

	baseURL := viper.GetString("AUTHENTICATION_OIDC_FRONT_END_URL")
	redirectURL := fmt.Sprintf("%s/auth/oidc/callback?token=%s", baseURL, url.QueryEscape(rawIDToken))
	http.Redirect(w, r, redirectURL, http.StatusFound)
}
