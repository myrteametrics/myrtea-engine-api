package variablesconfig

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"go.uber.org/zap"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
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
func (r *PostgresRepository) Get(id int64) (VariablesConfig, bool, error) {
	rows, err := r.newStatement().
		Select("key", "value").
		From(table).
		Where(sq.Eq{"id": id}).
		Query()
	if err != nil {
		return VariablesConfig{}, false, err
	}
	defer rows.Close()

	var key, value string
	if rows.Next() {
		err := rows.Scan(&key, &value)
		if err != nil {
			return VariablesConfig{}, false, fmt.Errorf("couldn't scan the action with id %d: %s", id, err.Error())
		}
	} else {
		return VariablesConfig{}, false, nil
	}

	return VariablesConfig{
		Id:    id,
		Key:   key,
		Value: value,
	}, true, nil
}

// GetByName use to retrieve an variableConfig by name
func (r *PostgresRepository) GetByKey(key string) (VariablesConfig, bool, error) {
	rows, err := r.newStatement().
		Select("id", "value").
		From(table).
		Where(sq.Eq{"key": key}).
		Query()
	if err != nil {
		return VariablesConfig{}, false, err
	}
	defer rows.Close()

	var id int64
	var value string
	if rows.Next() {
		err := rows.Scan(&id, &value)
		if err != nil {
			return VariablesConfig{}, false, fmt.Errorf("couldn't scan the action with name %s: %s", key, err.Error())
		}
	} else {
		return VariablesConfig{}, false, nil
	}

	return VariablesConfig{
		Id:    id,
		Key:   key,
		Value: value,
	}, true, nil
}

// Create method used to create an Variable Config
func (r *PostgresRepository) Create(variable VariablesConfig) (int64, error) {
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

	expression.G().Set(variable.Key, variable.Value)

	return id, nil
}

// Update method used to update un Variable Config
func (r *PostgresRepository) Update(id int64, variable VariablesConfig) error {
	res, err := r.newStatement().
		Update(table).
		Set("key", variable.Key).
		Set("Value", variable.Value).
		Where("id = ?", id).
		Exec()
	if err != nil {
		return err
	}

	expression.G().Set(variable.Key, variable.Value)

	return r.checkRowsAffected(res, 1)
}

func (r *PostgresRepository) Delete(id int64) error {
	builder := r.newStatement().
		Delete(table).
		Where("id = ?", id).
		Suffix("RETURNING key, value")

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
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}
		expression.G().Delete(key)
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
func (r *PostgresRepository) GetAll() ([]VariablesConfig, error) {
	variablesConfig := make([]VariablesConfig, 0)

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

		variable := VariablesConfig{
			Id:    id,
			Key:   key,
			Value: value,
		}

		variablesConfig = append(variablesConfig, variable)
	}
	return variablesConfig, nil
}

// GetAllAsMap method used to get all Variables Config as map[string]interface{}
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
