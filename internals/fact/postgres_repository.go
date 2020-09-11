package fact

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/myrteametrics/myrtea-sdk/v4/engine"
)

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

//Get search and returns an entity from the repository by its id
func (r *PostgresRepository) Get(id int64) (engine.Fact, bool, error) {
	query := `SELECT definition FROM fact_definition_v1 WHERE id = :id`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return engine.Fact{}, false, err
	}
	defer rows.Close()

	var data string
	if rows.Next() {
		err := rows.Scan(&data)
		if err != nil {
			return engine.Fact{}, false, err
		}
	} else {
		return engine.Fact{}, false, nil
	}

	// FIXME: Doesn't deserialize properly a fact (ConditionFragment has no requirement, and is converted in map[string]interface{})
	var fact engine.Fact
	err = json.Unmarshal([]byte(data), &fact)
	if err != nil {
		return engine.Fact{}, false, err
	}
	//This is necessary because within the definition we don't have the id
	fact.ID = id

	return fact, true, nil
}

//GetByName search and returns an entity from the repository by its name
func (r *PostgresRepository) GetByName(name string) (engine.Fact, bool, error) {
	query := `SELECT id, definition FROM fact_definition_v1 WHERE name = :name`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return engine.Fact{}, false, err
	}
	defer rows.Close()

	var id int64
	var data string
	if rows.Next() {
		err := rows.Scan(&id, &data)
		if err != nil {
			return engine.Fact{}, false, err
		}
	} else {
		return engine.Fact{}, false, nil
	}

	// FIXME: Doesn't deserialize properly a fact (ConditionFragment has no requirement, and is converted in map[string]interface{})
	var fact engine.Fact
	err = json.Unmarshal([]byte(data), &fact)
	if err != nil {
		return engine.Fact{}, false, err
	}
	//This is necessary because within the definition we don't have the id
	fact.ID = id

	return fact, true, nil
}

// Create creates a new Fact definition in the repository
func (r *PostgresRepository) Create(fact engine.Fact) (int64, error) {

	factdata, err := json.Marshal(fact)
	if err != nil {
		return -1, err
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()
	query := `INSERT INTO fact_definition_v1 (id, name, definition, last_modified) 
		VALUES (DEFAULT, :name, :definition, :last_modified) RETURNING id`
	params := map[string]interface{}{
		"name":          fact.Name,
		"definition":    string(factdata),
		"last_modified": timestamp,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var id int64
	if rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return -1, err
		}
	} else {
		return -1, errors.New("No id returning of insert situation")
	}

	return id, nil
}

// Update updates an entity in the repository by its name
func (r *PostgresRepository) Update(id int64, fact engine.Fact) error {
	query := `UPDATE fact_definition_v1 SET name = :name, definition = :definition, last_modified = :last_modified WHERE id = :id`

	//This is necessary because within the definition we don't have the id
	fact.ID = id

	factdata, err := json.Marshal(fact)
	if err != nil {
		return errors.New("Couldn't marshall the provided data:" + err.Error())
	}

	t := time.Now().Truncate(1 * time.Millisecond).UTC()
	res, err := r.conn.NamedExec(query, map[string]interface{}{
		"id":            id,
		"name":          fact.Name,
		"definition":    factdata,
		"last_modified": t,
	})
	if err != nil {
		return errors.New("Couldn't query the database:" + err.Error())
	}
	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("Error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("No row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// Delete deletes an entity from the repository by its name
func (r *PostgresRepository) Delete(id int64) error {
	query := `DELETE FROM fact_definition_v1 WHERE id = :id`

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
		return errors.New("No row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// GetAll returns all entities in the repository
func (r *PostgresRepository) GetAll() (map[int64]engine.Fact, error) {

	facts := make(map[int64]engine.Fact, 0)

	query := `SELECT id, definition FROM fact_definition_v1`
	rows, err := r.conn.Query(query)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var factID int64
		var factDef string
		var fact engine.Fact
		err := rows.Scan(&factID, &factDef)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(factDef), &fact)
		if err != nil {
			return nil, err
		}
		//This is necessary because within the definition we don't have the id
		fact.ID = factID

		facts[factID] = fact
	}
	return facts, nil

}
