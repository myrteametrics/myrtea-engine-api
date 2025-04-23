package connectorconfig

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
)

// PostgresRepository is a repository containing the ConnectorConfig definition based on a PSQL database and
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

// Get use to retrieve an ConnectorConfig by id
func (r *PostgresRepository) Get(id int64) (models.ConnectorConfig, bool, error) {
	query := `SELECT name, connector_id, current FROM connectors_config_v1 WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return models.ConnectorConfig{}, false, fmt.Errorf("couldn't retrieve the action with name %d: %s", id, err.Error())
	}
	defer rows.Close()

	var name, connectorId, current string
	if rows.Next() {
		err := rows.Scan(&name, &connectorId, &current)
		if err != nil {
			return models.ConnectorConfig{}, false, fmt.Errorf("couldn't scan the action with id %d: %s", id, err.Error())
		}
	} else {
		return models.ConnectorConfig{}, false, nil
	}

	return models.ConnectorConfig{
		Id:          id,
		Name:        name,
		ConnectorId: connectorId,
		Current:     current,
	}, true, nil
}

// Create method used to create an ConnectorConfig
func (r *PostgresRepository) Create(tx *sqlx.Tx, ConnectorConfig models.ConnectorConfig) (int64, error) {
	query := `INSERT into connectors_config_v1 (name, connector_id, current, last_modified) 
			 values (:name, :connector_id, :current, current_timestamp)`
	params := map[string]interface{}{
		"name":         ConnectorConfig.Name,
		"connector_id": ConnectorConfig.ConnectorId,
		"current":      ConnectorConfig.Current,
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

// Update method used to update un ConnectorConfig
func (r *PostgresRepository) Update(tx *sqlx.Tx, id int64, ConnectorConfig models.ConnectorConfig) error {
	query := `UPDATE connectors_config_v1 SET name = :name, connector_id = :connector_id,
				current = :current, previous = current WHERE id = :id`
	params := map[string]interface{}{
		"id":           id,
		"name":         ConnectorConfig.Name,
		"connector_id": ConnectorConfig.ConnectorId,
		"current":      ConnectorConfig.Current,
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

// Delete use to retrieve an ConnectorConfig by name
func (r *PostgresRepository) Delete(tx *sqlx.Tx, id int64) error {
	query := `DELETE FROM connectors_config_v1 WHERE id = :id`
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

// GetAll method used to get all ConnectorConfigs
func (r *PostgresRepository) GetAll() (map[int64]models.ConnectorConfig, error) {
	ConnectorConfigs := make(map[int64]models.ConnectorConfig)

	query := `SELECT id, name, connector_id, current FROM connectors_config_v1`
	rows, err := r.conn.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name, connectorId, current string

		err := rows.Scan(&id, &name, &connectorId, &current)
		if err != nil {
			return nil, err
		}

		ConnectorConfig := models.ConnectorConfig{
			Id:          id,
			Name:        name,
			ConnectorId: connectorId,
			Current:     current,
		}

		ConnectorConfigs[ConnectorConfig.Id] = ConnectorConfig
	}

	return ConnectorConfigs, nil
}
