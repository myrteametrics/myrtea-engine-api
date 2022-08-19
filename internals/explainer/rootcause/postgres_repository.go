package rootcause

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
)

// PostgresRepository is a repository containing the RootCause definition based on a PSQL database and
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

// Get use to retrieve an rootCause by id
func (r *PostgresRepository) Get(id int64) (models.RootCause, bool, error) {
	query := `SELECT name, description, situation_id, rule_id FROM ref_rootcause_v1 WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return models.RootCause{}, false, fmt.Errorf("Couldn't retrieve the action with id %d: %s", id, err.Error())
	}
	defer rows.Close()

	var name, description string
	var situationID, ruleID int64

	if rows.Next() {
		err := rows.Scan(&name, &description, &situationID, &ruleID)
		if err != nil {
			return models.RootCause{}, false, fmt.Errorf("Couldn't scan the action with id %d: %s", id, err.Error())
		}
	} else {
		return models.RootCause{}, false, nil
	}

	return models.RootCause{
		ID:          id,
		Name:        name,
		Description: description,
		SituationID: situationID,
		RuleID:      ruleID,
	}, true, nil
}

// Create method used to create an rootCause
func (r *PostgresRepository) Create(tx *sqlx.Tx, rootCause models.RootCause) (int64, error) {
	if !checkValidity(rootCause) {
		return -1, errors.New("missing rootcause data")
	}

	query := `INSERT into ref_rootcause_v1 (id, name, description, situation_id, rule_id) 
			 values (DEFAULT, :name, :description, :situation_id, :rule_id) RETURNING id`
	params := map[string]interface{}{
		"name":         rootCause.Name,
		"description":  rootCause.Description,
		"situation_id": rootCause.SituationID,
		"rule_id":      rootCause.RuleID,
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
		return -1, errors.New("no id returning of insert rootcause")
	}

	return id, nil
}

// Update method used to update un rootCause
func (r *PostgresRepository) Update(tx *sqlx.Tx, id int64, rootCause models.RootCause) error {
	if !checkValidity(rootCause) {
		return errors.New("missing rootcause data")
	}

	query := `UPDATE ref_rootcause_v1 SET name = :name, description = :description WHERE id = :id`
	params := map[string]interface{}{
		"id":          id,
		"name":        rootCause.Name,
		"description": rootCause.Description,
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

// Delete use to retrieve an rootCause by id
func (r *PostgresRepository) Delete(tx *sqlx.Tx, id int64) error {
	query := `DELETE FROM ref_rootcause_v1 WHERE id = :id`
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

// GetAll method used to get all rootCauses
func (r *PostgresRepository) GetAll() (map[int64]models.RootCause, error) {
	rootCauses := make(map[int64]models.RootCause, 0)

	query := `SELECT id, name, description, situation_id, rule_id FROM ref_rootcause_v1`
	rows, err := r.conn.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, situationID, ruleID int64
		var name, description string

		err := rows.Scan(&id, &name, &description, &situationID, &ruleID)
		if err != nil {
			return nil, err
		}

		rootCause := models.RootCause{
			ID:          id,
			Name:        name,
			Description: description,
			SituationID: situationID,
			RuleID:      ruleID,
		}

		rootCauses[rootCause.ID] = rootCause
	}
	return rootCauses, nil
}

// GetAllBySituationID method used to get all rootCauses for a specific situation ID
func (r *PostgresRepository) GetAllBySituationID(situationID int64) (map[int64]models.RootCause, error) {
	rootCauses := make(map[int64]models.RootCause, 0)

	query := `SELECT id, name, description, rule_id FROM ref_rootcause_v1 where situation_id = :situation_id`
	params := map[string]interface{}{
		"situation_id": situationID,
	}
	rows, err := r.conn.NamedQuery(query, params)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, ruleID int64
		var name, description string

		err := rows.Scan(&id, &name, &description, &ruleID)
		if err != nil {
			return nil, err
		}

		rootCause := models.RootCause{
			ID:          id,
			Name:        name,
			Description: description,
			SituationID: situationID,
			RuleID:      ruleID,
		}

		rootCauses[rootCause.ID] = rootCause
	}
	return rootCauses, nil
}

// GetAllBySituationIDRuleID method used to get all rootCauses for a specific situation ID and rule ID
func (r *PostgresRepository) GetAllBySituationIDRuleID(situationID int64, ruleID int64) (map[int64]models.RootCause, error) {
	rootCauses := make(map[int64]models.RootCause, 0)

	query := `SELECT id, name, description FROM ref_rootcause_v1 where situation_id = :situation_id and rule_id = :rule_id`
	params := map[string]interface{}{
		"situation_id": situationID,
		"rule_id":      ruleID,
	}
	rows, err := r.conn.NamedQuery(query, params)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name, description string

		err := rows.Scan(&id, &name, &description)
		if err != nil {
			return nil, err
		}

		rootCause := models.RootCause{
			ID:          id,
			Name:        name,
			Description: description,
			SituationID: situationID,
			RuleID:      ruleID,
		}

		rootCauses[rootCause.ID] = rootCause
	}
	return rootCauses, nil
}

func checkValidity(rootCause models.RootCause) bool {
	if rootCause.Name == "" || len(rootCause.Name) == 0 {
		return false
	}
	if rootCause.Description == "" || len(rootCause.Description) == 0 {
		return false
	}
	return true
}
