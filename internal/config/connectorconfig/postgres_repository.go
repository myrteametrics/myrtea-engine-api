package connectorconfig

import (
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/utils"
)

const table = "connectors_config_v1"

// PostgresRepository embeds utils.PostgresRepository to implement Repository interface
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

// Get use to retrieve an ConnectorConfig by id
func (r *PostgresRepository) Get(id int64) (model.ConnectorConfig, bool, error) {
	query := r.newStatement().
		Select("name", "connector_id", "current").
		From(table).
		Where(sq.Eq{"id": id})

	rows, err := query.Query()
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
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)

	var id int64
	var statement sq.InsertBuilder

	if tx != nil {
		statement = sq.Insert(table).
			PlaceholderFormat(sq.Dollar).
			RunWith(tx)
	} else {
		statement = r.newStatement().
			Insert(table)
	}

	statement = statement.
		Columns("name", "connector_id", "current", "last_modified").
		Values(ConnectorConfig.Name, ConnectorConfig.ConnectorId, ConnectorConfig.Current, sq.Expr("current_timestamp")).
		Suffix("RETURNING \"id\"")

	if ConnectorConfig.Id != 0 {
		statement = statement.
			Columns("id", "name", "connector_id", "current", "last_modified").
			Values(ConnectorConfig.Id, ConnectorConfig.Name, ConnectorConfig.ConnectorId, ConnectorConfig.Current, sq.Expr("current_timestamp"))
	}

	err := statement.QueryRow().Scan(&id)
	if err != nil {
		return -1, errors.New("couldn't query the database:" + err.Error())
	}

	return id, nil
}

// Update method used to update un ConnectorConfig
func (r *PostgresRepository) Update(tx *sqlx.Tx, id int64, ConnectorConfig model.ConnectorConfig) error {
	var statement sq.UpdateBuilder

	if tx != nil {
		statement = sq.Update(table).
			PlaceholderFormat(sq.Dollar).
			RunWith(tx)
	} else {
		statement = r.newStatement().
			Update(table)
	}

	statement = statement.
		Set("name", ConnectorConfig.Name).
		Set("connector_id", ConnectorConfig.ConnectorId).
		Set("current", ConnectorConfig.Current).
		Set("previous", sq.Expr("current")).
		Where(sq.Eq{"id": id})

	res, err := statement.Exec()
	if err != nil {
		return errors.New("couldn't query the database:" + err.Error())
	}

	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("no row updated (or multiple row updated) instead of 1 row")
	}
	return nil
}

// Delete use to retrieve an ConnectorConfig by name
func (r *PostgresRepository) Delete(tx *sqlx.Tx, id int64) error {
	var statement sq.DeleteBuilder

	if tx != nil {
		statement = sq.Delete(table).
			PlaceholderFormat(sq.Dollar).
			RunWith(tx)
	} else {
		statement = r.newStatement().
			Delete(table)
	}

	statement = statement.Where(sq.Eq{"id": id})

	res, err := statement.Exec()
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
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)
	return nil
}

// GetAll method used to get all ConnectorConfigs
func (r *PostgresRepository) GetAll() (map[int64]model.ConnectorConfig, error) {
	ConnectorConfigs := make(map[int64]model.ConnectorConfig)

	query := r.newStatement().
		Select("id", "name", "connector_id", "current").
		From(table)

	rows, err := query.Query()
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
