package externalconfig

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
)

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

// Get use to retrieve an externalConfig by name
func (r *PostgresRepository) Get(name string) (models.ExternalConfig, bool, error) {
	query := `SELECT data FROM external_generic_config_v1 WHERE name = :name`
	params := map[string]interface{}{
		"name": name,
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return models.ExternalConfig{}, false, fmt.Errorf("couldn't retrieve the action with name %s: %s", name, err.Error())
	}
	defer rows.Close()

	var data string
	if rows.Next() {
		err := rows.Scan(&data)
		if err != nil {
			return models.ExternalConfig{}, false, fmt.Errorf("couldn't scan the action with name %s: %s", name, err.Error())
		}
	} else {
		return models.ExternalConfig{}, false, nil
	}

	return models.ExternalConfig{
		Name: name,
		Data: data,
	}, true, nil
}

// Create method used to create an externalConfig
func (r *PostgresRepository) Create(tx *sqlx.Tx, externalConfig models.ExternalConfig) error {
	query := `INSERT into external_generic_config_v1 (name, data) 
			 values (:name, :data)`
	params := map[string]interface{}{
		"name": externalConfig.Name,
		"data": externalConfig.Data,
	}

	var err error
	var res sql.Result
	if tx != nil {
		res, err = tx.NamedExec(query, params)
	} else {
		res, err = r.conn.NamedExec(query, params)
	}
	if err != nil {
		return errors.New("couldn't query the database:" + err.Error())
	}

	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// Update method used to update un externalConfig
func (r *PostgresRepository) Update(tx *sqlx.Tx, name string, externalConfig models.ExternalConfig) error {
	query := `UPDATE external_generic_config_v1 SET data = :data WHERE name = :name`
	params := map[string]interface{}{
		"name": name,
		"data": externalConfig.Data,
	}

	var err error
	var res sql.Result
	if tx != nil {
		res, err = tx.NamedExec(query, params)
	} else {
		res, err = r.conn.NamedExec(query, params)
	}
	if err != nil {
		return errors.New("couldn't query the database:" + err.Error())
	}

	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// Delete use to retrieve an externalConfig by name
func (r *PostgresRepository) Delete(tx *sqlx.Tx, name string) error {
	query := `DELETE FROM external_generic_config_v1 WHERE name = :name`
	params := map[string]interface{}{
		"name": name,
	}

	var err error
	var res sql.Result
	if tx != nil {
		res, err = tx.NamedExec(query, params)
	} else {
		res, err = r.conn.NamedExec(query, params)
	}
	if err != nil {
		return errors.New("couldn't query the database:" + err.Error())
	}

	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i != 1 {
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// GetAll method used to get all externalConfigs
func (r *PostgresRepository) GetAll() (map[string]models.ExternalConfig, error) {
	externalConfigs := make(map[string]models.ExternalConfig)

	query := `SELECT name, data FROM external_generic_config_v1`
	rows, err := r.conn.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, data string

		err := rows.Scan(&name, &name, &data)
		if err != nil {
			return nil, err
		}

		externalConfig := models.ExternalConfig{
			Name: name,
			Data: data,
		}

		externalConfigs[externalConfig.Name] = externalConfig
	}
	return externalConfigs, nil
}
