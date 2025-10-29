package esconfig

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/utils"
)

const table = "elasticsearch_config_v1"

// encryptionKey should be 32 bytes for AES-256
// In production, this should come from environment variable or secure key management
var encryptionKey = []byte("myrtea-es-config-key-32bytes!!")

// PostgresRepository is a repository containing the ExternalConfig definition based on a PSQL database and
// implementing the repository interface
type PostgresRepository struct {
	conn *sqlx.DB
}

// NewPostgresRepository returns a new instance of PostgresRepository
func NewPostgresRepository(dbClient *sqlx.DB) Repository {
	r := PostgresRepository{
		conn: dbClient,
	}
	var repo Repository = &r
	return repo
}

// newStatement creates a new statement builder with Dollar format
func (r *PostgresRepository) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(r.conn.DB)
}

// checkRowsAffected Check if nbRows where affected to db
func (r *PostgresRepository) checkRowsAffected(res sql.Result, nbRows int64) error {
	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("error with the affected rows:" + err.Error())
	}
	if i != nbRows {
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// encryptPassword encrypts a password using AES
func encryptPassword(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(plaintext))

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptPassword decrypts a password using AES
func decryptPassword(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

// Get use to retrieve an elasticSearchConfig by id (password masked)
func (r *PostgresRepository) Get(id int64) (model.ElasticSearchConfig, bool, error) {
	rows, err := r.newStatement().
		Select("name", "urls", `"default"`, "export_activated", "auth", "insecure", "username").
		From(table).
		Where(sq.Eq{"id": id}).
		Query()
	if err != nil {
		return model.ElasticSearchConfig{}, false, err
	}
	defer rows.Close()

	var name, urls string
	var isDefault, exportActivated, auth, insecure bool
	var username sql.NullString
	if rows.Next() {
		err := rows.Scan(&name, &urls, &isDefault, &exportActivated, &auth, &insecure, &username)
		if err != nil {
			return model.ElasticSearchConfig{}, false, fmt.Errorf("couldn't scan the elasticsearch config with id %d: %s", id, err.Error())
		}
	} else {
		return model.ElasticSearchConfig{}, false, nil
	}

	return model.ElasticSearchConfig{
		Id:              id,
		Name:            name,
		URLs:            strings.Split(urls, ","),
		Default:         isDefault,
		ExportActivated: exportActivated,
		Auth:            auth,
		Insecure:        insecure,
		Username:        username.String,
		Password:        "", // Password is masked in regular Get
	}, true, nil
}

// GetForAuth use to retrieve an elasticSearchConfig by id with cleartext password for authentication
func (r *PostgresRepository) GetForAuth(id int64) (model.ElasticSearchConfig, bool, error) {
	rows, err := r.newStatement().
		Select("name", "urls", `"default"`, "export_activated", "auth", "insecure", "username", "password").
		From(table).
		Where(sq.Eq{"id": id}).
		Query()
	if err != nil {
		return model.ElasticSearchConfig{}, false, err
	}
	defer rows.Close()

	var name, urls string
	var isDefault, exportActivated, auth, insecure bool
	var username, encryptedPassword sql.NullString
	if rows.Next() {
		err := rows.Scan(&name, &urls, &isDefault, &exportActivated, &auth, &insecure, &username, &encryptedPassword)
		if err != nil {
			return model.ElasticSearchConfig{}, false, fmt.Errorf("couldn't scan the elasticsearch config with id %d: %s", id, err.Error())
		}
	} else {
		return model.ElasticSearchConfig{}, false, nil
	}

	// Decrypt password for authentication
	password, err := decryptPassword(encryptedPassword.String)
	if err != nil {
		return model.ElasticSearchConfig{}, false, fmt.Errorf("failed to decrypt password: %w", err)
	}

	return model.ElasticSearchConfig{
		Id:              id,
		Name:            name,
		URLs:            strings.Split(urls, ","),
		Default:         isDefault,
		ExportActivated: exportActivated,
		Auth:            auth,
		Insecure:        insecure,
		Username:        username.String,
		Password:        password, // Return cleartext password
	}, true, nil
}

// GetByName use to retrieve an elasticSearchConfig by name (password masked)
func (r *PostgresRepository) GetByName(name string) (model.ElasticSearchConfig, bool, error) {
	rows, err := r.newStatement().
		Select("id", "urls", `"default"`, "export_activated", "auth", "insecure", "username").
		From(table).
		Where(sq.Eq{"name": name}).
		Query()
	if err != nil {
		return model.ElasticSearchConfig{}, false, err
	}
	defer rows.Close()

	esConfig := model.ElasticSearchConfig{
		Name: name,
	}
	var urls string
	var username sql.NullString
	if rows.Next() {
		err = rows.Scan(&esConfig.Id, &urls, &esConfig.Default, &esConfig.ExportActivated, &esConfig.Auth, &esConfig.Insecure, &username)
		if err != nil {
			return model.ElasticSearchConfig{}, false, fmt.Errorf("couldn't scan the elasticsearch config with name %s: %s", name, err.Error())
		}
	} else {
		return model.ElasticSearchConfig{}, false, nil
	}

	esConfig.URLs = strings.Split(urls, ",")
	esConfig.Username = username.String
	esConfig.Password = "" // Password is masked in regular GetByName

	return esConfig, true, nil
}

// GetByNameForAuth use to retrieve an elasticSearchConfig by name with cleartext password for authentication
func (r *PostgresRepository) GetByNameForAuth(name string) (model.ElasticSearchConfig, bool, error) {
	rows, err := r.newStatement().
		Select("id", "urls", `"default"`, "export_activated", "auth", "insecure", "username", "password").
		From(table).
		Where(sq.Eq{"name": name}).
		Query()
	if err != nil {
		return model.ElasticSearchConfig{}, false, err
	}
	defer rows.Close()

	esConfig := model.ElasticSearchConfig{
		Name: name,
	}
	var urls string
	var username, encryptedPassword sql.NullString
	if rows.Next() {
		err = rows.Scan(&esConfig.Id, &urls, &esConfig.Default, &esConfig.ExportActivated, &esConfig.Auth, &esConfig.Insecure, &username, &encryptedPassword)
		if err != nil {
			return model.ElasticSearchConfig{}, false, fmt.Errorf("couldn't scan the elasticsearch config with name %s: %s", name, err.Error())
		}
	} else {
		return model.ElasticSearchConfig{}, false, nil
	}

	esConfig.URLs = strings.Split(urls, ",")
	esConfig.Username = username.String

	// Decrypt password for authentication
	password, err := decryptPassword(encryptedPassword.String)
	if err != nil {
		return model.ElasticSearchConfig{}, false, fmt.Errorf("failed to decrypt password: %w", err)
	}
	esConfig.Password = password // Return cleartext password

	return esConfig, true, nil
}

// GetDefault use to retrieve the default elasticSearchConfig (password masked)
func (r *PostgresRepository) GetDefault() (model.ElasticSearchConfig, bool, error) {
	rows, err := r.newStatement().
		Select("id", "name", "urls", "export_activated", "auth", "insecure", "username").
		From(table).
		Where(sq.Eq{`"default"`: true}).
		Query()
	if err != nil {
		return model.ElasticSearchConfig{}, false, err
	}
	defer rows.Close()

	esConfig := model.ElasticSearchConfig{
		Default: true,
	}
	var urls string
	var username sql.NullString
	if rows.Next() {
		err = rows.Scan(&esConfig.Id, &esConfig.Name, &urls, &esConfig.ExportActivated, &esConfig.Auth, &esConfig.Insecure, &username)
		if err != nil {
			return model.ElasticSearchConfig{}, false, fmt.Errorf("couldn't scan the default elasticsearch config: %s", err.Error())
		}
	} else {
		return model.ElasticSearchConfig{}, false, nil
	}

	esConfig.URLs = strings.Split(urls, ",")
	esConfig.Username = username.String
	esConfig.Password = "" // Password is masked in regular GetDefault

	return esConfig, true, nil
}

// GetDefaultForAuth use to retrieve the default elasticSearchConfig with cleartext password for authentication
func (r *PostgresRepository) GetDefaultForAuth() (model.ElasticSearchConfig, bool, error) {
	rows, err := r.newStatement().
		Select("id", "name", "urls", "export_activated", "auth", "insecure", "username", "password").
		From(table).
		Where(sq.Eq{`"default"`: true}).
		Query()
	if err != nil {
		return model.ElasticSearchConfig{}, false, err
	}
	defer rows.Close()

	esConfig := model.ElasticSearchConfig{
		Default: true,
	}
	var urls string
	var username, encryptedPassword sql.NullString
	if rows.Next() {
		err = rows.Scan(&esConfig.Id, &esConfig.Name, &urls, &esConfig.ExportActivated, &esConfig.Auth, &esConfig.Insecure, &username, &encryptedPassword)
		if err != nil {
			return model.ElasticSearchConfig{}, false, fmt.Errorf("couldn't scan the default elasticsearch config: %s", err.Error())
		}
	} else {
		return model.ElasticSearchConfig{}, false, nil
	}

	esConfig.URLs = strings.Split(urls, ",")
	esConfig.Username = username.String

	// Decrypt password for authentication
	password, err := decryptPassword(encryptedPassword.String)
	if err != nil {
		return model.ElasticSearchConfig{}, false, fmt.Errorf("failed to decrypt password: %w", err)
	}
	esConfig.Password = password // Return cleartext password

	return esConfig, true, nil
}

// Create method used to create an elasticSearchConfig
func (r *PostgresRepository) Create(elasticSearchConfig model.ElasticSearchConfig) (int64, error) {
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)

	// Encrypt password if provided
	encryptedPassword, err := encryptPassword(elasticSearchConfig.Password)
	if err != nil {
		return -1, fmt.Errorf("failed to encrypt password: %w", err)
	}

	var id int64
	statement := r.newStatement().
		Insert(table).
		Suffix("RETURNING \"id\"")
	if elasticSearchConfig.Id != 0 {
		statement = statement.
			Columns("id", "name", "urls", `"default"`, "export_activated", "auth", "insecure", "username", "password").
			Values(elasticSearchConfig.Id, elasticSearchConfig.Name, strings.Join(elasticSearchConfig.URLs, ","),
				elasticSearchConfig.Default, elasticSearchConfig.ExportActivated, elasticSearchConfig.Auth,
				elasticSearchConfig.Insecure, elasticSearchConfig.Username, encryptedPassword)
	} else {
		statement = statement.
			Columns("name", "urls", `"default"`, "export_activated", "auth", "insecure", "username", "password").
			Values(elasticSearchConfig.Name, strings.Join(elasticSearchConfig.URLs, ","),
				elasticSearchConfig.Default, elasticSearchConfig.ExportActivated, elasticSearchConfig.Auth,
				elasticSearchConfig.Insecure, elasticSearchConfig.Username, encryptedPassword)
	}
	err = statement.QueryRow().Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

// Update method used to update un elasticSearchConfig
func (r *PostgresRepository) Update(id int64, elasticSearchConfig model.ElasticSearchConfig) error {
	updateStmt := r.newStatement().
		Update(table).
		Set("name", elasticSearchConfig.Name).
		Set("urls", strings.Join(elasticSearchConfig.URLs, ",")).
		Set(`"default"`, elasticSearchConfig.Default).
		Set("export_activated", elasticSearchConfig.ExportActivated).
		Set("auth", elasticSearchConfig.Auth).
		Set("insecure", elasticSearchConfig.Insecure).
		Set("username", elasticSearchConfig.Username).
		Where("id = ?", id)

	// Only update password if it's provided (not empty)
	if elasticSearchConfig.Password != "" {
		encryptedPassword, err := encryptPassword(elasticSearchConfig.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		updateStmt = updateStmt.Set("password", encryptedPassword)
	}

	res, err := updateStmt.Exec()
	if err != nil {
		return err
	}
	return r.checkRowsAffected(res, 1)
}

// Delete use to retrieve an elasticSearchConfig by name
func (r *PostgresRepository) Delete(id int64) error {
	res, err := r.newStatement().
		Delete(table).
		Where("id = ?", id).
		Exec()
	if err != nil {
		return err
	}
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)
	return r.checkRowsAffected(res, 1)

	// TODO: check & set default
}

// GetAll method used to get all elasticSearchConfigs
func (r *PostgresRepository) GetAll() (map[int64]model.ElasticSearchConfig, error) {
	elasticSearchConfigs := make(map[int64]model.ElasticSearchConfig)
	// Note: password is excluded from GetAll for security
	rows, err := r.newStatement().
		Select("id", "name", "urls", `"default"`, "export_activated", "auth", "insecure", "username").
		From(table).
		Query()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		esConfig := model.ElasticSearchConfig{}
		var urls string
		var username sql.NullString
		err = rows.Scan(&esConfig.Id, &esConfig.Name, &urls, &esConfig.Default, &esConfig.ExportActivated,
			&esConfig.Auth, &esConfig.Insecure, &username)
		if err != nil {
			return nil, err
		}

		esConfig.URLs = strings.Split(urls, ",")
		esConfig.Username = username.String
		// Password is intentionally not included in GetAll
		elasticSearchConfigs[esConfig.Id] = esConfig
	}
	return elasticSearchConfigs, nil
}
