package routeroidc

import (
	"errors"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

type OidcInstance struct {
	OidcConfig oauth2.Config
	Provider   *oidc.Provider
}

var (
	mu       sync.Mutex
	instance *OidcInstance
)

func InitOidc(oidcIssuerURL string, oidcClientID string, oidcClientSecret string, oidcRedirectURL string, oidcScopes []string) error {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		return nil
	}

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, oidcIssuerURL)
	if err != nil {
		zap.L().Error("create instance  oidc provider failled ", zap.Error(err))
		return err
	}

	//Scopes
	scopes := []string{oidc.ScopeOpenID}
	scopes = append(scopes, oidcScopes...)

	oidcConfig := oauth2.Config{
		ClientID:     oidcClientID,
		ClientSecret: oidcClientSecret,
		RedirectURL:  oidcRedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}

	instance = &OidcInstance{
		OidcConfig: oidcConfig,
		Provider:   provider,
	}

	return nil
}

func GetOidcInstance() (*OidcInstance, error) {
	mu.Lock()
	defer mu.Unlock()

	if instance == nil {
		zap.L().Error("OIDC instance is not initialized. Call InitOidc first.")
		return nil, errors.New("OIDC instance is not initialized. Call InitOidc first.")
	}

	return instance, nil
}
