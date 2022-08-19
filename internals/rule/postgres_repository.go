package rule

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/situation"
	"go.uber.org/zap"
)

// PostgresRulesRepository is a repository containing the rules based on a PSQL database and
// implementing the repository interface
type PostgresRulesRepository struct {
	conn *sqlx.DB
}

// NewPostgresRepository returns a new instance of PostgresRulesRepository
func NewPostgresRepository(dbClient *sqlx.DB) Repository {
	r := PostgresRulesRepository{
		conn: dbClient,
	}
	var ifm Repository = &r
	return ifm
}

// CheckByName returns if at least one row exists with the input rule name
func (r *PostgresRulesRepository) CheckByName(name string) (bool, error) {
	var exists bool
	checkNameQuery := `select exists(select 1 from rules_v1 where name = $1) AS "exists"`
	err := r.conn.QueryRow(checkNameQuery, name).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

// Create creates a new Rule in the repository
func (r *PostgresRulesRepository) Create(rule Rule) (int64, error) {

	t := time.Now().Truncate(1 * time.Millisecond).UTC()
	tx, err := r.conn.Begin()
	if err != nil {
		return -1, err
	}

	var rows *sql.Rows

	if rule.CalendarID == 0 {
		rows, err = tx.Query(`INSERT INTO rules_v1(name, enabled, calendar_id, last_modified)
							VALUES ($1,$2,$3,$4) RETURNING id`, rule.Name, rule.Enabled, nil, t)

	} else {
		rows, err = tx.Query(`INSERT INTO rules_v1(name, enabled, calendar_id, last_modified)
		VALUES ($1,$2,$3,$4) RETURNING id`, rule.Name, rule.Enabled, rule.CalendarID, t)
	}

	if err != nil {
		tx.Rollback()
		return -1, err
	}
	defer rows.Close()

	var ruleID int64
	if rows.Next() {
		rows.Scan(&ruleID)
	} else {
		tx.Rollback()
		return -1, errors.New("no id returning of insert rule action")
	}
	rows.Close()

	rule.ID = ruleID
	ruledata, err := json.Marshal(rule)
	if err != nil {
		return -1, errors.New("failled to marshall the rule:" + rule.Name +
			"\nError from Marshal" + err.Error())
	}

	//insert rule version
	res, err := tx.Exec(`INSERT INTO rule_versions_v1(rule_id, version_number, data, creation_datetime)
							VALUES ($1,$2,$3,$4)`, ruleID, rule.Version, string(ruledata), t)
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	i, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return -1, errors.New("error with the affected rows:" + err.Error())
	}
	if i != 1 {
		tx.Rollback()
		return -1, errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}

	err = tx.Commit()
	if err != nil {
		return -1, err
	}

	return ruleID, nil
}

//Get search and returns an entity from the repository by its id
func (r *PostgresRulesRepository) Get(id int64) (Rule, bool, error) {
	query := `select rules_v1.id, rule_versions_v1.version_number, rule_versions_v1.data 
			from rules_v1 inner join rule_versions_v1 on rules_v1.id = rule_versions_v1.rule_id 
			where rules_v1.id = :id 
			order by version_number desc LIMIT 1`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return Rule{}, false, errors.New("couldn't retrieve the Rule with id: " + fmt.Sprint(id) + " : " + err.Error())
	}
	defer rows.Close()

	var rule Rule
	var ruleID int64
	var ruleVersionNumber int64
	var data string
	if rows.Next() {
		err := rows.Scan(&ruleID, &ruleVersionNumber, &data)
		if err != nil {
			return Rule{}, false, errors.New("couldn't scan the retrieved data: " + err.Error())
		}
		err = json.Unmarshal([]byte(data), &rule)
		if err != nil {
			return Rule{}, false, errors.New("malformed Data, rule ID: " + fmt.Sprint(id) + " error: " + err.Error())
		}

		//Set the id coming from the DB
		rule.ID = ruleID
		rule.Version = ruleVersionNumber

		return rule, true, nil
	}
	return Rule{}, false, nil
}

// GetByVersion search and returns an entity from the repository by its id
func (r *PostgresRulesRepository) GetByVersion(id int64, version int64) (Rule, bool, error) {
	query := `select rules_v1.id, rule_versions_v1.version_number, rule_versions_v1.data 
			from rules_v1 inner join rule_versions_v1 on rules_v1.id = rule_versions_v1.rule_id 
			where rules_v1.id = :id and rule_versions_v1.version_number = :version`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id":      id,
		"version": version,
	})
	if err != nil {
		return Rule{}, false, errors.New("couldn't retrieve the Rule with id: " + fmt.Sprint(id) + " : " + err.Error())
	}
	defer rows.Close()

	var rule Rule
	var ruleID int64
	var ruleVersionNumber int64
	var data string
	if rows.Next() {
		err := rows.Scan(&ruleID, &ruleVersionNumber, &data)
		if err != nil {
			return Rule{}, false, errors.New("couldn't scan the retrieved data: " + err.Error())
		}
		err = json.Unmarshal([]byte(data), &rule)
		if err != nil {
			return Rule{}, false, errors.New("malformed Data, rule ID: " + fmt.Sprint(id) + " error: " + err.Error())
		}

		//Set the id coming from the DB
		rule.ID = ruleID
		rule.Version = ruleVersionNumber

		return rule, true, nil
	}
	return Rule{}, false, nil
}

//GetByName search and returns an entity from the repository by its name
func (r *PostgresRulesRepository) GetByName(name string) (Rule, bool, error) {
	query := `select rules_v1.id, rule_versions_v1.version_number, rule_versions_v1.data 
		from rules_v1 inner join rule_versions_v1 on rules_v1.id = rule_versions_v1.rule_id 
		where rules_v1.name = :name 
		order by version_number desc LIMIT 1`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return Rule{}, false, errors.New("couldn't retrieve the Rule with name: " + name + " : " + err.Error())
	}
	defer rows.Close()

	var rule Rule
	var ruleID int64
	var ruleVersionNumber int64
	var data string
	if rows.Next() {
		err := rows.Scan(&ruleID, &ruleVersionNumber, &data)
		if err != nil {
			return Rule{}, false, errors.New("couldn't scan the retrieved data: " + err.Error())
		}
		err = json.Unmarshal([]byte(data), &rule)
		if err != nil {
			return Rule{}, false, errors.New("malformed Data, rule Name: " + name + " error: " + err.Error())
		}

		//Set the id coming from the DB
		rule.ID = ruleID
		rule.Version = ruleVersionNumber

		return rule, true, nil
	}
	return Rule{}, false, nil
}

// Update updates an entity in the repository by its name
func (r *PostgresRulesRepository) Update(rule Rule) error {

	var newVersion bool
	existing, found, err := r.Get(rule.ID)
	if err != nil {
		return err
	}
	if found {
		if rule.SameCasesAs(existing) {
			rule.Version = existing.Version
			newVersion = false
		} else {
			rule.Version = existing.Version + 1
			newVersion = true
		}
	} else {
		return fmt.Errorf("The rule with ID %d was not found", rule.ID)
	}

	t := time.Now().Truncate(1 * time.Millisecond).UTC()
	ruledata, err := json.Marshal(rule)
	if err != nil {
		return errors.New("failled to marshall the rule:" + rule.Name +
			"\nError from Marshal" + err.Error())
	}

	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	var res sql.Result
	if newVersion {
		//Insert new Version
		res, err = tx.Exec(`INSERT INTO rule_versions_v1(rule_id, version_number, data, creation_datetime)
							VALUES ($1,$2,$3,$4)`, rule.ID, rule.Version, string(ruledata), t)
	} else {
		//Update Version
		res, err = tx.Exec(`UPDATE rule_versions_v1 SET data = $1
							WHERE rule_id = $2 and version_number = $3`, string(ruledata), rule.ID, rule.Version)
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	i, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return errors.New("error with the affected rows:" + err.Error())
	}
	if i != 1 {
		tx.Rollback()
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}

	//Update Rule
	if rule.CalendarID == 0 {
		res, err = tx.Exec(`UPDATE rules_v1 SET name = $1, enabled = $2, calendar_id = $3, last_modified = $4 WHERE id = $5`,
			rule.Name, rule.Enabled, nil, t, rule.ID)
	} else {
		res, err = tx.Exec(`UPDATE rules_v1 SET name = $1, enabled = $2, calendar_id = $3, last_modified = $4 WHERE id = $5`,
			rule.Name, rule.Enabled, rule.CalendarID, t, rule.ID)
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	i, err = res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return errors.New("error with the affected rows:" + err.Error())
	}
	if i != 1 {
		tx.Rollback()
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes an entry from the repository by it's ID
func (r *PostgresRulesRepository) Delete(id int64) error {
	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	//Delete rule_versions
	_, err = tx.Exec(`DELETE FROM rule_versions_v1 WHERE rule_id = $1`, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	//Delete situation rules
	_, err = tx.Exec(`DELETE FROM situation_rules_v1 WHERE rule_id = $1`, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	//Delete rule
	res, err := tx.Exec(`DELETE FROM rules_v1 WHERE id = $1`, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	i, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if i <= 0 {
		tx.Rollback()
		return errors.New("no row modified during delete")
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// GetAll returns all entities in the repository
func (r *PostgresRulesRepository) GetAll() (map[int64]Rule, error) {

	query := `SELECT rules_v1.id, (SELECT data FROM rule_versions_v1 WHERE rule_id = id ORDER BY version_number DESC LIMIT 1) FROM rules_v1`
	rows, err := r.conn.Query(query)
	if err != nil {
		zap.L().Error("Couldn't retrieve the Rules", zap.Error(err))
		return nil, errors.New("couldn't retrieve the Rules " + err.Error())
	}
	defer rows.Close()

	rules := make(map[int64]Rule, 0)
	for rows.Next() {
		var ruledata string
		var rule Rule
		var ruleID int64

		err := rows.Scan(&ruleID, &ruledata)
		if err != nil {
			zap.L().Error("Couldn't read the rows:", zap.Error(err))
			return nil, errors.New("couldn't read the rows: " + err.Error())
		}
		err = json.Unmarshal([]byte(ruledata), &rule)
		if err != nil {
			zap.L().Error("Couldn't unmarshall the rule data:", zap.Error(err))
			return nil, errors.New("couldn't unmarshall the rule data: " + err.Error())
		}

		//Set the id coming from the DB
		rule.ID = ruleID

		rules[ruleID] = rule
	}
	return rules, nil
}

// GetAllEnabled returns all entities in the repository that are enabled (column enabled = true)
func (r *PostgresRulesRepository) GetAllEnabled() (map[int64]Rule, error) {

	query := `SELECT rules_v1.id, (SELECT data FROM rule_versions_v1 WHERE rule_id = id ORDER BY version_number DESC LIMIT 1)
		FROM rules_v1 WHERE enabled = true`
	rows, err := r.conn.Query(query)
	if err != nil {
		zap.L().Error("Couldn't retrieve the Rules", zap.Error(err))
		return nil, errors.New("couldn't retrieve the Rules " + err.Error())
	}
	defer rows.Close()

	rules := make(map[int64]Rule, 0)
	for rows.Next() {
		var ruledata string
		var rule Rule
		var ruleID int64

		err := rows.Scan(&ruleID, &ruledata)
		if err != nil {
			zap.L().Error("Couldn't read the rows:", zap.Error(err))
			return nil, errors.New("couldn't read the rows: " + err.Error())
		}
		err = json.Unmarshal([]byte(ruledata), &rule)
		if err != nil {
			zap.L().Error("Couldn't unmarshall the rule data:", zap.Error(err))
			return nil, errors.New("couldn't unmarshall the rule data: " + err.Error())
		}

		//Set the id coming from the DB
		rule.ID = ruleID

		rules[ruleID] = rule
	}
	return rules, nil
}

// GetAllModifiedFrom returns all entities that have been modified since 'from' parameter
func (r *PostgresRulesRepository) GetAllModifiedFrom(from time.Time) (map[int64]Rule, error) {

	query := `SELECT rules_v1.id, (SELECT data FROM rule_versions_v1 WHERE rule_id = id ORDER BY version_number DESC LIMIT 1)
				FROM rules_v1
				WHERE last_modified >= :last_modified ORDER BY last_modified ASC`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"last_modified": from.Truncate(1 * time.Millisecond).UTC(),
	})
	if err != nil {
		zap.L().Error("Couldn't retrieve the Rules", zap.Error(err))
		return nil, errors.New("couldn't retrieve the Rules " + err.Error())
	}
	defer rows.Close()

	rules := make(map[int64]Rule, 0)
	for rows.Next() {
		var ruledata string
		var rule Rule
		var ruleID int64

		err := rows.Scan(&ruleID, &ruledata)
		if err != nil {
			zap.L().Error("Couldn't read the rows:", zap.Error(err))
			return nil, errors.New("couldn't read the rows: " + err.Error())
		}
		err = json.Unmarshal([]byte(ruledata), &rule)
		if err != nil {
			zap.L().Error("Couldn't unmarshall the rule data:", zap.Error(err))
			return nil, errors.New("couldn't unmarshall the rule data: " + err.Error())
		}

		//Set the id coming from the DB
		rule.ID = ruleID

		rules[ruleID] = rule
	}
	return rules, nil
}

func (r *PostgresRulesRepository) GetEnabledRuleIDs(situationID int64, ts time.Time) ([]int64, error) {

	ruleIDs, err := situation.R().GetRules(situationID)
	if err != nil {
		return nil, fmt.Errorf("error geting rules for situation instance (%d): %s", situationID, err.Error())
	}

	ruleIDsInt := make([]int64, 0)
	for _, id := range ruleIDs {
		r, found, err := r.Get(id)
		if err != nil {
			zap.L().Error("Get Rule", zap.Int64("id", id), zap.Error(err))
			continue
		}
		if !found {
			zap.L().Warn("Rule is missing", zap.Int64("id", id))
			continue
		}

		cfound, valid, _ := calendar.CBase().InPeriodFromCalendarID(int64(r.CalendarID), ts)
		if !cfound || valid {
			ruleIDsInt = append(ruleIDsInt, id)
		}
	}

	return ruleIDsInt, nil
}
