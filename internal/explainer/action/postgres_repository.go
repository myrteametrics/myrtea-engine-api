package action

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/model"
)

// PostgresRepository is a repository containing the Action definition based on a PSQL database and
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

// Get use to retrieve an action by id
func (r *PostgresRepository) Get(id int64) (model.Action, bool, error) {
	query := `SELECT name, description, rootcause_id FROM ref_action_v1 WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return model.Action{}, false, fmt.Errorf("Couldn't retrieve the action with id %d: %s", id, err.Error())
	}
	defer rows.Close()

	var name, description string
	var rootCauseID int64

	if rows.Next() {
		err := rows.Scan(&name, &description, &rootCauseID)
		if err != nil {
			return model.Action{}, false, fmt.Errorf("Couldn't scan the action with id %d: %s", id, err.Error())
		}
	} else {
		return model.Action{}, false, nil
	}

	return model.Action{
		ID:          id,
		Name:        name,
		Description: description,
		RootCauseID: rootCauseID,
	}, true, nil
}

// Create method used to create an action
func (r *PostgresRepository) Create(tx *sqlx.Tx, action model.Action) (int64, error) {
	if !checkValidity(action) {
		return -1, errors.New("missing action data")
	}

	query := `INSERT into ref_action_v1 (id, name, description, rootcause_id) 
			 values (DEFAULT, :name, :description, :rootcause_id) RETURNING id;`
	params := map[string]interface{}{
		"name":         action.Name,
		"description":  action.Description,
		"rootcause_id": action.RootCauseID,
	}

	var err error
	var res *sqlx.Rows
	if tx != nil {
		res, err = tx.NamedQuery(query, params)
	} else {
		res, err = r.conn.NamedQuery(query, params)
	}
	if err != nil {
		return -1, err
	}
	defer res.Close()

	var id int64
	if res.Next() {
		res.Scan(&id)
	} else {
		return -1, errors.New("no id returning of insert action")
	}

	return id, nil
}

// Update method used to update un action
func (r *PostgresRepository) Update(tx *sqlx.Tx, id int64, action model.Action) error {
	if !checkValidity(action) {
		return errors.New("missing action data")
	}

	query := `UPDATE ref_action_v1 SET name = :name, description = :description WHERE id = :id`
	params := map[string]interface{}{
		"id":          id,
		"name":        action.Name,
		"description": action.Description,
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

// Delete use to retrieve an action by id
func (r *PostgresRepository) Delete(tx *sqlx.Tx, id int64) error {
	query := `DELETE FROM ref_action_v1 WHERE id = :id`
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

// GetAll method used to get all actions
func (r *PostgresRepository) GetAll() (map[int64]model.Action, error) {
	actions := make(map[int64]model.Action, 0)

	query := `SELECT id, name, description, rootcause_id FROM ref_action_v1`
	rows, err := r.conn.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, rootCauseID int64
		var name, description string

		err := rows.Scan(&id, &name, &description, &rootCauseID)
		if err != nil {
			return nil, err
		}

		action := model.Action{
			ID:          id,
			Name:        name,
			Description: description,
			RootCauseID: rootCauseID,
		}

		actions[action.ID] = action
	}
	return actions, nil
}

// GetAllBySituationID method used to get all actions for a specific situation ID
func (r *PostgresRepository) GetAllBySituationID(situationID int64) (map[int64]model.Action, error) {
	actions := make(map[int64]model.Action, 0)

	query := `SELECT a.id, a.name, a.description, a.rootcause_id 
		FROM ref_action_v1 a INNER JOIN ref_rootcause_v1 rc ON a.rootcause_id = rc.id
		where rc.situation_id = :situation_id`
	params := map[string]interface{}{
		"situation_id": situationID,
	}
	rows, err := r.conn.NamedQuery(query, params)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, rootCauseID int64
		var name, description string

		err := rows.Scan(&id, &name, &description, &rootCauseID)
		if err != nil {
			return nil, err
		}

		action := model.Action{
			ID:          id,
			Name:        name,
			Description: description,
			RootCauseID: rootCauseID,
		}

		actions[action.ID] = action
	}
	return actions, nil
}

// GetAllByRootCauseID method used to get all actions for a specific rootcause ID
func (r *PostgresRepository) GetAllByRootCauseID(rootCauseID int64) (map[int64]model.Action, error) {
	actions := make(map[int64]model.Action, 0)

	query := `SELECT id, name, description, rootcause_id FROM ref_action_v1 where rootcause_id = :rootcause_id`
	params := map[string]interface{}{
		"rootcause_id": rootCauseID,
	}
	rows, err := r.conn.NamedQuery(query, params)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, rootCauseID int64
		var name, description string

		err := rows.Scan(&id, &name, &description, &rootCauseID)
		if err != nil {
			return nil, err
		}

		action := model.Action{
			ID:          id,
			Name:        name,
			Description: description,
			RootCauseID: rootCauseID,
		}

		actions[action.ID] = action
	}
	return actions, nil
}

func checkValidity(action model.Action) bool {
	if action.Name == "" || len(action.Name) == 0 {
		return false
	}
	if action.Description == "" || len(action.Description) == 0 {
		return false
	}
	return true
}
