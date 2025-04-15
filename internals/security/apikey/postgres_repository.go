package apikey

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	uuid "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const table = "api_keys"

var fields = []string{"id", "key_hash", "key_prefix", "name", "role_id", "created_at", "expires_at", "last_used_at", "is_active", "created_by"}

type APIKeyWithValue struct {
	APIKey
	KeyValue string `json:"keyValue"`
}

// PostgresRepository is a repository containing the API keys data
// based on a PSQL database and implementing the Repository interface
type PostgresRepository struct {
	conn *sqlx.DB
}

// NewPostgresRepository returns a new instance of PostgresRepository
func NewPostgresRepository(dbClient *sqlx.DB) Repository {
	r := PostgresRepository{
		conn: dbClient,
	}
	var ifm Repository = &r
	return ifm
}

// Get retrieves and returns an APIKey from the repository by its id
func (r *PostgresRepository) Get(apiKeyUUID uuid.UUID) (APIKey, bool, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where("id = ?", apiKeyUUID).
		Query()
	if err != nil {
		return APIKey{}, false, err
	}
	defer rows.Close()
	return r.scanFirst(rows)
}

// Create creates a new APIKey in the repository
func (r *PostgresRepository) Create(apiKey APIKey) (APIKey, error) {
	newUUID := uuid.New()
	if apiKey.ID != uuid.Nil {
		newUUID = apiKey.ID
	}

	keyValue := GenerateAPIKey(apiKey.KeyPrefix)
	keyHash, err := HashAPIKey(keyValue)
	if err != nil {
		return APIKey{}, err
	}

	_, err = r.newStatement().
		Insert(table).
		Columns(fields...).
		Values(
			newUUID,
			keyHash,
			apiKey.KeyPrefix,
			apiKey.Name,
			apiKey.RoleID,
			apiKey.CreatedAt,
			apiKey.ExpiresAt,
			apiKey.LastUsedAt,
			apiKey.IsActive,
			apiKey.CreatedBy,
		).
		Exec()
	if err != nil {
		return APIKey{}, err
	}
	apiKey.ID = newUUID
	apiKey.KeyHash = keyValue
	return apiKey, nil
}

// Update updates an APIKey in the repository
func (r *PostgresRepository) Update(apiKey APIKey) error {
	result, err := r.newStatement().
		Update(table).
		Set("name", apiKey.Name).
		Set("role_id", apiKey.RoleID).
		Set("expires_at", apiKey.ExpiresAt).
		Set("is_active", apiKey.IsActive).
		Where("id = ?", apiKey.ID).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowAffected(result, 1)
}

// Delete deletes an APIKey from the repository
func (r *PostgresRepository) Delete(uuid uuid.UUID) error {
	result, err := r.newStatement().
		Delete(table).
		Where("id = ?", uuid).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowAffected(result, 1)
}

// Deactivate deactivates an APIKey without deleting it
func (r *PostgresRepository) Deactivate(uuid uuid.UUID) error {
	result, err := r.newStatement().
		Update(table).
		Set("is_active", false).
		Where("id = ?", uuid).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowAffected(result, 1)
}

// GetAll retrieves all APIKeys from the repository
func (r *PostgresRepository) GetAll() ([]APIKey, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAll(rows)
}

// GetAllForRole retrieves all APIKeys associated with a specific role
func (r *PostgresRepository) GetAllForRole(roleUUID uuid.UUID) ([]APIKey, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where("role_id = ?", roleUUID).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAll(rows)
}

// Validate checks if an API key is valid and returns the associated information
func (r *PostgresRepository) Validate(keyValue string) (APIKey, bool, error) {

	if len(keyValue) < 3 {
		return APIKey{}, false, errors.New("invalid API key format")
	}

	parts := strings.Split(keyValue, "_")
	if len(parts) < 2 {
		return APIKey{}, false, errors.New("invalid API key format: missing prefix separator")
	}
	keyPrefix := parts[0]

	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where("key_prefix = ?", keyPrefix).
		Where("is_active = true").
		Where("(expires_at IS NULL OR expires_at > NOW())").
		Query()
	if err != nil {
		return APIKey{}, false, err
	}
	defer rows.Close()

	apiKeys, err := r.scanAll(rows)
	if err != nil {
		return APIKey{}, false, err
	}

	for _, apiKey := range apiKeys {
		if CompareAPIKey(keyValue, apiKey.KeyHash) {
			_, _ = r.newStatement().
				Update(table).
				Set("last_used_at", time.Now()).
				Where("id = ?", apiKey.ID).
				Exec()

			return apiKey, true, nil
		}
	}

	return APIKey{}, true, nil
}

func (r *PostgresRepository) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(r.conn.DB)
}

func (r *PostgresRepository) scanFirst(rows *sql.Rows) (APIKey, bool, error) {
	if rows.Next() {
		apiKey, err := r.scan(rows)
		return apiKey, err == nil, err
	}
	return APIKey{}, false, nil
}

func (r *PostgresRepository) scanAll(rows *sql.Rows) ([]APIKey, error) {
	apiKeys := make([]APIKey, 0)
	for rows.Next() {
		apiKey, err := r.scan(rows)
		if err != nil {
			return []APIKey{}, err
		}
		apiKeys = append(apiKeys, apiKey)
	}
	return apiKeys, nil
}

func (r *PostgresRepository) checkRowAffected(result sql.Result, nbRows int64) error {
	i, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if i != nbRows {
		return errors.New("no row affected (or multiple rows affected) instead of 1 row")
	}
	return nil
}

func (r *PostgresRepository) scan(rows *sql.Rows) (APIKey, error) {
	apiKey := APIKey{}
	err := rows.Scan(
		&apiKey.ID,
		&apiKey.KeyHash,
		&apiKey.KeyPrefix,
		&apiKey.Name,
		&apiKey.RoleID,
		&apiKey.CreatedAt,
		&apiKey.ExpiresAt,
		&apiKey.LastUsedAt,
		&apiKey.IsActive,
		&apiKey.CreatedBy,
	)
	if err != nil {
		return APIKey{}, errors.New("couldn't scan the retrieved data: " + err.Error())
	}
	return apiKey, nil
}
