package externalconfig

import (
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
)

const table = "external_generic_config_v1"

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

// Get use to retrieve an externalConfig by id
func (r *PostgresRepository) Get(id int64) (models.ExternalConfig, bool, error) {
	rows, err := r.newStatement().
		Select("name", "data").
		From(table).
		Where(sq.Eq{"id": id}).
		Query()
	if err != nil {
		return models.ExternalConfig{}, false, err
	}
	defer rows.Close()

	var name, data string
	if rows.Next() {
		err := rows.Scan(&name, &data)
		if err != nil {
			return models.ExternalConfig{}, false, fmt.Errorf("couldn't scan the action with id %d: %s", id, err.Error())
		}
	} else {
		return models.ExternalConfig{}, false, nil
	}

	return models.ExternalConfig{
		Id:   id,
		Name: name,
		Data: data,
	}, true, nil
}

// GetByName use to retrieve an externalConfig by name
func (r *PostgresRepository) GetByName(name string) (models.ExternalConfig, bool, error) {
	rows, err := r.newStatement().
		Select("id", "data").
		From(table).
		Where(sq.Eq{"name": name}).
		Query()
	if err != nil {
		return models.ExternalConfig{}, false, err
	}
	defer rows.Close()

	var id int64
	var data string
	if rows.Next() {
		err := rows.Scan(&id, &data)
		if err != nil {
			return models.ExternalConfig{}, false, fmt.Errorf("couldn't scan the action with name %s: %s", name, err.Error())
		}
	} else {
		return models.ExternalConfig{}, false, nil
	}

	return models.ExternalConfig{
		Id:   id,
		Name: name,
		Data: data,
	}, true, nil
}

// Create method used to create an externalConfig
func (r *PostgresRepository) Create(externalConfig models.ExternalConfig) (int64, error) {
	var id int64
	err := r.newStatement().
		Insert(table).
		Columns("name", "data").
		Values(externalConfig.Name, externalConfig.Data).
		Suffix("RETURNING \"id\"").
		QueryRow().
		Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

// Update method used to update un externalConfig
func (r *PostgresRepository) Update(id int64, externalConfig models.ExternalConfig) error {
	res, err := r.newStatement().
		Update(table).
		Set("name", externalConfig.Name).
		Set("data", externalConfig.Data).
		Where("id = ?", id).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowsAffected(res, 1)
}

// Delete use to retrieve an externalConfig by name
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

// GetAll method used to get all externalConfigs
func (r *PostgresRepository) GetAll() (map[int64]models.ExternalConfig, error) {
	externalConfigs := make(map[int64]models.ExternalConfig)
	rows, err := r.newStatement().
		Select("id", "name", "data").
		From(table).
		Query()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name, data string

		err := rows.Scan(&id, &name, &data)
		if err != nil {
			return nil, err
		}

		externalConfig := models.ExternalConfig{
			Id:   id,
			Name: name,
			Data: data,
		}

		externalConfigs[externalConfig.Id] = externalConfig
	}
	return externalConfigs, nil
}
