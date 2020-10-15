package router

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

// GetIDPMetadata returns the IDP metadata descriptor from a local XML file or a remote URL
func GetIDPMetadata(mode string, filePath string, fetchURL string) (*saml.EntityDescriptor, error) {
	if mode == "FILE" {
		return parseLocalIDPMetadata(filePath)
	}
	if mode == "FETCH" {
		return fetchRemoteIDPMetadata(fetchURL)
	}
	return nil, fmt.Errorf("Invalid SAML Metadata mode : %s", mode)
}

func fetchRemoteIDPMetadata(samlIDPMetadataURL string) (*saml.EntityDescriptor, error) {
	idpMetadataURL, err := url.Parse(samlIDPMetadataURL)
	if err != nil {
		return nil, err
	}
	metadata, err := samlsp.FetchMetadata(context.TODO(), http.DefaultClient, *idpMetadataURL)
	return metadata, err
}

func parseLocalIDPMetadata(samlIDPMetadataFile string) (*saml.EntityDescriptor, error) {
	data, err := ioutil.ReadFile(samlIDPMetadataFile)
	if err != nil {
		return nil, err
	}
	metadata, err := samlsp.ParseMetadata(data)
	return metadata, err
}
