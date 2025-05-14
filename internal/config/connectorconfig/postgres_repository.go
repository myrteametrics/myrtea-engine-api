package connectorconfig

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"go.uber.org/zap"
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
func (r *PostgresRepository) Get(id int64) (model.ConnectorConfig, bool, error) {
	query := `SELECT name, connector_id, current FROM connectors_config_v1 WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return model.ConnectorConfig{}, false, fmt.Errorf("couldn't retrieve the action with name %d: %s", id, err.Error())
	}
	defer rows.Close()

	var name, connectorId, current string
	if rows.Next() {
		err := rows.Scan(&name, &connectorId, &current)
		if err != nil {
			return model.ConnectorConfig{}, false, fmt.Errorf("couldn't scan the action with id %d: %s", id, err.Error())
		}
	} else {
		return model.ConnectorConfig{}, false, nil
	}

	return model.ConnectorConfig{
		Id:          id,
		Name:        name,
		ConnectorId: connectorId,
		Current:     current,
	}, true, nil
}

// Create method used to create an ConnectorConfig
func (r *PostgresRepository) Create(tx *sqlx.Tx, ConnectorConfig model.ConnectorConfig) (int64, error) {
	_, _, _ = r.refreshNextIdGen()
	query := `INSERT into connectors_config_v1 (id, name, connector_id, current, last_modified) 
			 values (DEFAULT, :name, :connector_id, :current, current_timestamp)`
	params := map[string]interface{}{
		"name":         ConnectorConfig.Name,
		"connector_id": ConnectorConfig.ConnectorId,
		"current":      ConnectorConfig.Current,
	}
	if ConnectorConfig.Id != 0 {
		query = `INSERT into connectors_config_v1 (id, name, connector_id, current, last_modified) 
			 values (:id, :name, :connector_id, :current, current_timestamp)`
		params["id"] = ConnectorConfig.Id
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
func (r *PostgresRepository) Update(tx *sqlx.Tx, id int64, ConnectorConfig model.ConnectorConfig) error {
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
		return errors.New("no row deleted (or multiple row deleted) instead of 1 row")
	}
	_, _, _ = r.refreshNextIdGen()
	return nil
}

// GetAll method used to get all ConnectorConfigs
func (r *PostgresRepository) GetAll() (map[int64]model.ConnectorConfig, error) {
	ConnectorConfigs := make(map[int64]model.ConnectorConfig)

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

		ConnectorConfig := model.ConnectorConfig{
			Id:          id,
			Name:        name,
			ConnectorId: connectorId,
			Current:     current,
		}

		ConnectorConfigs[ConnectorConfig.Id] = ConnectorConfig
	}

	return ConnectorConfigs, nil
}

func (r *PostgresRepository) refreshNextIdGen() (int64, bool, error) {
	query := `SELECT setval(pg_get_serial_sequence('connectors_config_v1', 'id'), coalesce(max(id),0) + 1, false) FROM connectors_config_v1`
	rows, err := r.conn.Query(query)

	if err != nil {
		zap.L().Error("Couldn't query the database:", zap.Error(err))
		return 0, false, err
	}
	defer rows.Close()

	var data int64
	if rows.Next() {
		err := rows.Scan(&data)
		if err != nil {
			return 0, false, err
		}
		return data, true, nil
	} else {
		return 0, false, nil
	}
}
