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

// Get use to retrieve an externalConfig by id
func (r *PostgresRepository) Get(id int64) (models.ExternalConfig, bool, error) {
	query := `SELECT name, data FROM external_generic_config_v1 WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return models.ExternalConfig{}, false, fmt.Errorf("couldn't retrieve the action with name %d: %s", id, err.Error())
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
	query := `SELECT id, data FROM external_generic_config_v1 WHERE name = :name`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"name": name,
	})
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
func (r *PostgresRepository) Create(tx *sqlx.Tx, externalConfig models.ExternalConfig) (int64, error) {
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
		return -1, errors.New("couldn't query the database:" + err.Error())
	}

	i, err := res.RowsAffected()
	if err != nil {
		return -1, errors.New("error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return -1, errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return -1, nil
}

// Update method used to update un externalConfig
func (r *PostgresRepository) Update(tx *sqlx.Tx, id int64, externalConfig models.ExternalConfig) error {
	query := `UPDATE external_generic_config_v1 SET name = :name, data = :data WHERE id = :id`
	params := map[string]interface{}{
		"id":   id,
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

// Delete use to retrieve an externalConfig by name
func (r *PostgresRepository) Delete(tx *sqlx.Tx, id int64) error {
	query := `DELETE FROM external_generic_config_v1 WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
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
func (r *PostgresRepository) GetAll() (map[int64]models.ExternalConfig, error) {
	externalConfigs := make(map[int64]models.ExternalConfig)

	query := `SELECT id, name, data FROM external_generic_config_v1`
	rows, err := r.conn.Query(query)

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
