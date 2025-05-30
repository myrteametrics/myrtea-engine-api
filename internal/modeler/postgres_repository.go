package modeler

import (
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/utils"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/myrteametrics/myrtea-sdk/v5/modeler"
	"go.uber.org/zap"
)

const table = "model_v1"

// PostgresRepository is a repository containing the Fact definition based on a PSQL database and
// implementing the repository interface
type PostgresRepository struct {
	conn *sqlx.DB
}

// NewPostgresRepository returns a new instance of PostgresRepository
func NewPostgresRepository(dbClient *sqlx.DB) Repository {
	r := PostgresRepository{
		conn: dbClient,
	}
	var ifm Repository = &r
	return ifm
}

// newStatement creates a new statement builder with Dollar format
func (r *PostgresRepository) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(r.conn.DB)
}

// Get search and returns a model from the repository by its name
func (r *PostgresRepository) Get(id int64) (modeler.Model, bool, error) {
	query := `SELECT definition FROM model_v1 WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return modeler.Model{}, false, err
	}
	defer rows.Close()

	var model modeler.Model
	var data string
	if rows.Next() {
		err := rows.Scan(&data)
		if err != nil {
			return modeler.Model{}, false, err
		}
	} else {
		return modeler.Model{}, false, nil
	}

	err = json.Unmarshal([]byte(data), &model)
	if err != nil {
		return modeler.Model{}, false, err
	}
	model.ID = id

	return model, true, nil
}

// GetByName search and returns a model from the repository by its name
func (r *PostgresRepository) GetByName(name string) (modeler.Model, bool, error) {
	query := `SELECT id, definition FROM model_v1 WHERE name = :name`
	params := map[string]interface{}{
		"name": name,
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return modeler.Model{}, false, err
	}
	defer rows.Close()

	var modelID int64
	var data string
	if rows.Next() {
		err := rows.Scan(&modelID, &data)
		if err != nil {
			return modeler.Model{}, false, err
		}
	} else {
		return modeler.Model{}, false, nil
	}

	var model modeler.Model
	err = json.Unmarshal([]byte(data), &model)
	if err != nil {
		return modeler.Model{}, false, err
	}
	model.ID = modelID

	return model, true, nil
}

// Create creates a new model definition in the repository
func (r *PostgresRepository) Create(model modeler.Model) (int64, error) {
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, "model_v1")
	modelData, err := json.Marshal(model)
	if err != nil {
		return -1, err
	}

	var id int64
	var statement sq.InsertBuilder

	statement = r.newStatement().
		Insert(table).
		Suffix("RETURNING \"id\"")

	if model.ID != 0 {
		statement = statement.
			Columns("id", "name", "definition").
			Values(model.ID, model.Name, string(modelData))
	} else {
		statement = statement.
			Columns("name", "definition").
			Values(model.Name, string(modelData))
	}

	err = statement.QueryRow().Scan(&id)
	if err != nil {
		return -1, errors.New("couldn't query the database:" + err.Error())
	}

	return id, nil
}

// Update updates a model in the repository by its name
func (r *PostgresRepository) Update(id int64, model modeler.Model) error {
	modeldata, err := json.Marshal(model)
	if err != nil {
		return errors.New("couldn't marshall the provided data:" + err.Error())
	}

	query := `UPDATE model_v1 SET name = :name, definition = :definition WHERE id = :id`
	params := map[string]interface{}{
		"id":         id,
		"name":       model.Name,
		"definition": modeldata,
	}
	res, err := r.conn.NamedExec(query, params)
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

// Delete deletes a model from the repository by its name
func (r *PostgresRepository) Delete(id int64) error {
	query := `DELETE FROM model_v1 WHERE id = :id`

	res, err := r.conn.NamedExec(query, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return err
	}
	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i != 1 {
		return errors.New("no row deleted (or multiple row deleted) instead of 1 row")
	}
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, "model_v1")
	return nil
}

// GetAll returns all models in the repository
func (r *PostgresRepository) GetAll() (map[int64]modeler.Model, error) {

	models := make(map[int64]modeler.Model)

	query := `SELECT id, definition FROM model_v1`
	rows, err := r.conn.Query(query)
	if err != nil {
		zap.L().Error("Couldn't retrieve the models", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var modelID int64
		var modelDef string
		var model modeler.Model
		err := rows.Scan(&modelID, &modelDef)
		if err != nil {
			zap.L().Error("Couldn't read the model rows:", zap.Error(err))
			return nil, err
		}

		err = json.Unmarshal([]byte(modelDef), &model)
		if err != nil {
			zap.L().Error("Couldn't unmarshall the model data:", zap.Error(err))
			return nil, err
		}
		model.ID = modelID

		models[modelID] = model
	}
	return models, nil

}

// GetAll returns all models in the repository
func (r *PostgresRepository) GetAllByIDs(ids []int64) (map[int64]modeler.Model, error) {

	models := make(map[int64]modeler.Model, 0)

	query := `SELECT id, definition FROM model_v1 WHERE id = ANY(:ids)`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"ids": pq.Array(ids),
	})
	if err != nil {
		zap.L().Error("Couldn't retrieve the models", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var modelID int64
		var modelDef string
		var model modeler.Model
		err := rows.Scan(&modelID, &modelDef)
		if err != nil {
			zap.L().Error("Couldn't read the model rows:", zap.Error(err))
			return nil, err
		}

		err = json.Unmarshal([]byte(modelDef), &model)
		if err != nil {
			zap.L().Error("Couldn't unmarshall the model data:", zap.Error(err))
			return nil, err
		}
		model.ID = modelID

		models[modelID] = model
	}
	return models, nil

}
