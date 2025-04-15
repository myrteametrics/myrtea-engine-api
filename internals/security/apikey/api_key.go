package apikey

import (
	"errors"
	"github.com/google/uuid"
	sdksecurity "github.com/myrteametrics/myrtea-sdk/v5/security"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const ApiKeyLength = 48

// APIKey represents a stored API key
type APIKey struct {
	ID         uuid.UUID  `json:"id"`
	KeyHash    string     `json:"keyHash"`
	KeyPrefix  string     `json:"keyPrefix"`
	Name       string     `json:"name"`
	RoleID     uuid.UUID  `json:"roleId"`
	CreatedAt  time.Time  `json:"createdAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	IsActive   bool       `json:"isActive"`
	CreatedBy  string     `json:"createdBy"`
}

func (key *APIKey) IsValid() error {
	if key.Name == "" {
		return errors.New("missing Name")
	}

	if len(key.Name) < 3 {
		return errors.New("name is too short (less than 3 characters)")
	}

	if key.RoleID == uuid.Nil {
		return errors.New("missing RoleID")
	}

	if key.CreatedBy == "" {
		return errors.New("missing CreatedBy")
	}

	if key.ExpiresAt != nil {
		if key.ExpiresAt.Before(time.Now()) {
			return errors.New("expiration date must be in the future")
		}
	}

	return nil
}

// IsValidForCreate checks the validity for creation (with additional validations)
func (key *APIKey) IsValidForCreate() error {
	if err := key.IsValid(); err != nil {
		return err
	}

	if key.ExpiresAt == nil {
		defaultExpiration := time.Now().AddDate(0, 1, 0)
		key.ExpiresAt = &defaultExpiration
	}

	if key.CreatedAt.IsZero() {
		key.CreatedAt = time.Now()
	}

	if key.LastUsedAt == nil {
		key.LastUsedAt = &key.CreatedAt
	}

	return nil
}

// IsValidForUpdate checks the validity for update
func (key *APIKey) IsValidForUpdate() error {
	if err := key.IsValid(); err != nil {
		return err
	}

	if key.ID == uuid.Nil {
		return errors.New("missing ID for update")
	}

	return nil
}

// GenerateAPIKey creates a new API key with a prefix
func GenerateAPIKey(prefix string) string {
	randomPart := sdksecurity.RandString(ApiKeyLength - len(prefix))
	return prefix + "_" + randomPart
}

// HashAPIKey generates a bcrypt hash for an API key
func HashAPIKey(keyValue string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(keyValue), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CompareAPIKey compares an API key with a bcrypt hash
func CompareAPIKey(keyValue, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(keyValue))
	return err == nil
}
