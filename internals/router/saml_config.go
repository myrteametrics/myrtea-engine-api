package router

import (
	"errors"
	"fmt"
)

// SamlSPMiddlewareConfig wraps multiple parameters for SAML authentication
type SamlSPMiddlewareConfig struct {
	MetadataMode             string
	MetadataFilePath         string
	MetadataURL              string
	AttributeUserID          string
	AttributeUserDisplayName string
	EnableMemberOfValidation bool
	AttributeUserMemberOf    string
}

// IsValid check if the config is valid
func (config SamlSPMiddlewareConfig) IsValid() (bool, error) {
	if config.MetadataMode != "FILE" && config.MetadataMode != "FETCH" {
		return false, fmt.Errorf("Invalid metadata mode : %s", config.MetadataMode)
	}
	if config.MetadataMode == "FILE" && config.MetadataFilePath == "" {
		return false, errors.New("Metadata mode FILE require a non-empty AUTHENTICATION_SAML_METADATA_FILE_PATH")
	}
	if config.MetadataMode == "FETCH" && config.MetadataURL == "" {
		return false, errors.New("Metadata mode FETCH require a non-empty AUTHENTICATION_SAML_METADATA_URL")
	}
	if config.AttributeUserID == "" {
		return false, errors.New("AttributeUserID cannot be empty")
	}
	if config.EnableMemberOfValidation && config.AttributeUserMemberOf == "" {
		return false, errors.New("MemberOf validation require a non-empty AUTHENTICATION_SAML_ATTRIBUTE_USER_MEMBEROF")
	}
	return true, nil
}
