package variablesconfig

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"go.uber.org/zap"

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

	listMap, err := repo.GetAllAsMap()

	if err != nil {
		zap.L().Fatal("Unable to retrieve the list of global variables", zap.Error(err))
	}

	expression.G().Load(listMap)

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
		Select("key", "value", "scope").
		From(table).
		Where(sq.Eq{"id": id}).
		Query()
	if err != nil {
		return models.VariablesConfig{}, false, err
	}
	defer rows.Close()

	var key, value, scope string
	if rows.Next() {
		err := rows.Scan(&key, &value, &scope)
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
		Scope: scope,
	}, true, nil
}

// GetByName use to retrieve an variableConfig by name
func (r *PostgresRepository) GetByKey(key string) (models.VariablesConfig, bool, error) {
	rows, err := r.newStatement().
		Select("id", "value", "scope").
		From(table).
		Where(sq.Eq{"key": key}).
		Query()
	if err != nil {
		return models.VariablesConfig{}, false, err
	}
	defer rows.Close()

	var id int64
	var value, scope string
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
		Scope: scope,
	}, true, nil
}

// Create method used to create an Variable Config
func (r *PostgresRepository) Create(variable models.VariablesConfig) (int64, error) {
	var id int64
	err := r.newStatement().
		Insert(table).
		Columns("key", "value", "scope").
		Values(variable.Key, variable.Value, variable.Scope).
		Suffix("RETURNING \"id\"").
		QueryRow().
		Scan(&id)
	if err != nil {
		return -1, err
	}

	if variable.Scope == "global" {
		expression.G().Set(variable.Key, variable.Value)
	}

	return id, nil
}

// Update method used to update un Variable Config
func (r *PostgresRepository) Update(id int64, variable models.VariablesConfig) error {
	res, err := r.newStatement().
		Update(table).
		Set("key", variable.Key).
		Set("value", variable.Value).
		Set("scope", variable.Scope).
		Where("id = ?", id).
		Exec()
	if err != nil {
		return err
	}

	if variable.Scope == "global" {
		expression.G().Set(variable.Key, variable.Value)
	}

	return r.checkRowsAffected(res, 1)
}

func (r *PostgresRepository) Delete(id int64) error {
	builder := r.newStatement().
		Delete(table).
		Where("id = ?", id).
		Suffix("RETURNING key, value, scope")

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build SQL query: %v", err)
	}

	rows, err := r.conn.Query(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	var rowCount int64

	for rows.Next() {
		var key, value, scope string
		if err := rows.Scan(&key, &value, &scope); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}
		if scope == "global" {
			expression.G().Delete(key)
		}
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %v", err)
	}

	if rowCount != 1 {
		return errors.New("unexpected number of rows affected")
	}

	return nil
}

// GetAll method used to get all Variables Config
func (r *PostgresRepository) GetAll() ([]models.VariablesConfig, error) {
	return r.GetAllByScope("")
}

// GetAll method used to get all Variables Config filtered on a specified scope value
// Scope used to filter is by default set to "global" for retro-compatibility
func (r *PostgresRepository) GetAllByScope(scope string) ([]models.VariablesConfig, error) {
	variablesConfig := make([]models.VariablesConfig, 0)

	query := r.newStatement().
		Select("id", "key", "value", "scope").
		From(table)

	if scope != "" {
		query = query.Where(sq.Eq{"scope": scope})
	}

	rows, err := query.Query()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var key, value, scope string

		err := rows.Scan(&id, &key, &value, &scope)
		if err != nil {
			return nil, err
		}

		variable := models.VariablesConfig{
			Id:    id,
			Key:   key,
			Value: value,
			Scope: scope,
		}

		variablesConfig = append(variablesConfig, variable)
	}
	return variablesConfig, nil
}

// GetAllAsMap method used to get all Variables Config as map[string]interface{}
func (r *PostgresRepository) GetAllAsMap() (map[string]interface{}, error) {
	return r.GetAllAsMapByScope("")
}

// GetAllAsMap method used to get all Variables Config as map[string]interface{} filtered on a specified scope value
// Scope used to filter is by default set to "global" for retro-compatibility
func (r *PostgresRepository) GetAllAsMapByScope(scope string) (map[string]interface{}, error) {

	variableConfigMap := make(map[string]interface{})

	query := r.newStatement().
		Select("key", "value", "scope").
		From(table)

	if scope != "" {
		query = query.Where(sq.Eq{"scope": scope})
	}

	rows, err := query.Query()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value, scope string

		err := rows.Scan(&key, &value, &scope)
		if err != nil {
			return nil, err
		}

		variableConfigMap[key] = value
	}
	return variableConfigMap, nil
}
