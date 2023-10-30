package variablesconfig

import (
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
)

const table = "variables_config_v1"

// PostgresRepository is a repository containing the VariablesConfig definition based on a PSQL database and
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

// Get use to retrieve an variableConfig by id
func (r *PostgresRepository) Get(id int64) (models.VariablesConfig, bool, error) {
	rows, err := r.newStatement().
		Select("key", "value").
		From(table).
		Where(sq.Eq{"id": id}).
		Query()
	if err != nil {
		return models.VariablesConfig{}, false, err
	}
	defer rows.Close()

	var key, value string
	if rows.Next() {
		err := rows.Scan(&key, &value)
		if err != nil {
			return models.VariablesConfig{}, false, fmt.Errorf("couldn't scan the action with id %d: %s", id, err.Error())
		}
	} else {
		return models.VariablesConfig{}, false, nil
	}

	return models.VariablesConfig{
		Id:    id,
		Key:   key,
		Value: value,
	}, true, nil
}

// GetByName use to retrieve an variableConfig by name
func (r *PostgresRepository) GetByKey(key string) (models.VariablesConfig, bool, error) {
	rows, err := r.newStatement().
		Select("id", "value").
		From(table).
		Where(sq.Eq{"key": key}).
		Query()
	if err != nil {
		return models.VariablesConfig{}, false, err
	}
	defer rows.Close()

	var id int64
	var value string
	if rows.Next() {
		err := rows.Scan(&id, &value)
		if err != nil {
			return models.VariablesConfig{}, false, fmt.Errorf("couldn't scan the action with name %s: %s", key, err.Error())
		}
	} else {
		return models.VariablesConfig{}, false, nil
	}

	return models.VariablesConfig{
		Id:    id,
		Key:   key,
		Value: value,
	}, true, nil
}

// Create method used to create an Variable Config
func (r *PostgresRepository) Create(variable models.VariablesConfig) (int64, error) {
	var id int64
	err := r.newStatement().
		Insert(table).
		Columns("key", "value").
		Values(variable.Key, variable.Value).
		Suffix("RETURNING \"id\"").
		QueryRow().
		Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

// Update method used to update un Variable Config
func (r *PostgresRepository) Update(id int64, variable models.VariablesConfig) error {
	res, err := r.newStatement().
		Update(table).
		Set("key", variable.Key).
		Set("Value", variable.Value).
		Where("id = ?", id).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowsAffected(res, 1)
}

// Delete use to retrieve an Variable Config by id
func (r *PostgresRepository) Delete(id int64) error {
	res, err := r.newStatement().
		Delete(table).
		Where("id = ?", id).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowsAffected(res, 1)
}

// GetAll method used to get all Variables Config
func (r *PostgresRepository) GetAll() ([]models.VariablesConfig, error) {

	var variablesConfig []models.VariablesConfig

	rows, err := r.newStatement().
		Select("id", "key", "value").
		From(table).
		Query()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var key, value string

		err := rows.Scan(&id, &key, &value)
		if err != nil {
			return nil, err
		}

		variable := models.VariablesConfig{
			Id:    id,
			Key:   key,
			Value: value,
		}

		variablesConfig = append(variablesConfig, variable)
	}
	return variablesConfig, nil
}

func (r *PostgresRepository) GetAllAsMap() (map[string]interface{}, error) {

	variableConfigMap := make(map[string]interface{})

	rows, err := r.newStatement().
		Select("key", "value").
		From(table).
		Query()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string

		err := rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}

		variableConfigMap[key] = value
	}
	return variableConfigMap, nil
}
