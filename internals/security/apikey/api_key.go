package apikey

import (
	"errors"
	"github.com/google/uuid"
	sdksecurity "github.com/myrteametrics/myrtea-sdk/v5/security"
	"time"
)

const ApiKeyLength = 24

// APIKey represents a stored API key
type APIKey struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	KeyHash    string     `json:"keyHash" db:"key_hash"`
	KeyPrefix  string     `json:"keyPrefix" db:"key_prefix"`
	Name       string     `json:"name" db:"name"`
	RoleID     uuid.UUID  `json:"roleId" db:"role_id"`
	CreatedAt  time.Time  `json:"createdAt" db:"created_at"`
	ExpiresAt  *time.Time `json:"expiresAt" db:"expires_at"`
	LastUsedAt *time.Time `json:"lastUsedAt" db:"last_used_at"`
	IsActive   bool       `json:"isActive" db:"is_active"`
	CreatedBy  string     `json:"createdBy" db:"created_by"`
}

type CreateResponse struct {
	Id  uuid.UUID `json:"id"`
	Key string    `json:"key"`
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
