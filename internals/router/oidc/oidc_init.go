package oidcAuth

import (
	"log"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var (
	OidcConfig oauth2.Config
	Provider   *oidc.Provider
)

func InitOidc() {
	ctx := context.Background()
	var err error
	Provider, err = oidc.NewProvider(ctx, viper.GetString("AUTHENTICATION_OIDC_ISSUER_URL"))
	if err != nil {
		log.Fatal(err)
	}

	OidcConfig = oauth2.Config{
		ClientID:     viper.GetString("AUTHENTICATION_OIDC_CLIENT_ID"),
		ClientSecret: viper.GetString("AUTHENTICATION_OIDC_CLIENT_SECRET"),
		RedirectURL:  viper.GetString("AUTHENTICATION_OIDC_REDIRECT_URL"),
		Endpoint:     Provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
}
