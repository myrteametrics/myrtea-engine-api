package router

import (
	"net/url"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	dsig "github.com/russellhaering/goxmldsig"
)

// CustomServiceProvider returns a custom saml.ServiceProvider for the provided
// options.
func CustomServiceProvider(opts samlsp.Options) saml.ServiceProvider {
	metadataURL := opts.URL.ResolveReference(&url.URL{Path: "saml/metadata"})
	acsURL := opts.URL.ResolveReference(&url.URL{Path: "saml/acs"})
	sloURL := opts.URL.ResolveReference(&url.URL{Path: "saml/slo"})

	var forceAuthn *bool
	if opts.ForceAuthn {
		forceAuthn = &opts.ForceAuthn
	}

	// signatureMethod := dsig.RSASHA1SignatureMethod
	signatureMethod := dsig.RSASHA256SignatureMethod
	if !opts.SignRequest {
		signatureMethod = ""
	}

	return saml.ServiceProvider{
		EntityID:          opts.EntityID,
		Key:               opts.Key,
		Certificate:       opts.Certificate,
		Intermediates:     opts.Intermediates,
		MetadataURL:       *metadataURL,
		AcsURL:            *acsURL,
		SloURL:            *sloURL,
		IDPMetadata:       opts.IDPMetadata,
		ForceAuthn:        forceAuthn,
		SignatureMethod:   signatureMethod,
		AllowIDPInitiated: opts.AllowIDPInitiated,
	}
}
