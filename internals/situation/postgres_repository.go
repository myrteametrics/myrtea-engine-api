package situation

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"go.uber.org/zap"
)

// PostgresRepository is a repository containing the situation definition based on a PSQL database and
//implementing the repository interface
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

// Get retrieve the specified situation definition
func (r *PostgresRepository) Get(id int64, groupsIDS []int64) (Situation, bool, error) {
	query := `SELECT definition,
				ARRAY(SELECT fact_id FROM situation_facts_v1 WHERE situation_id = :id) as fact_ids
				FROM situation_definition_v1 WHERE id = :id and groups && :groups`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id":     id,
		"groups": pq.Array(groupsIDS),
	})

	if err != nil {
		return Situation{}, false, err
	}
	defer rows.Close()

	var data string
	var factIDs []int64
	if rows.Next() {
		err := rows.Scan(&data, pq.Array(&factIDs))
		if err != nil {
			return Situation{}, false, err
		}
	} else {
		return Situation{}, false, nil
	}

	var situation Situation
	err = json.Unmarshal([]byte(data), &situation)
	if err != nil {
		return Situation{}, false, err
	}
	situation.Facts = factIDs

	//This is necessary because within the definition we don't have the id
	situation.ID = id

	//Need to delete at any get of situation the universal token group
	situation.Groups = groups.DeleteTokenAllGroups(situation.Groups)

	return situation, true, nil
}

// GetByName retrieve the specified situation definition by it's name
func (r *PostgresRepository) GetByName(name string, groupsIDS []int64) (Situation, bool, error) {
	query := `SELECT id, definition,
				ARRAY(SELECT fact_id FROM situation_facts_v1 WHERE situation_id = id) as fact_ids
				FROM situation_definition_v1 WHERE name = :name and groups && :groups`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"name":   name,
		"groups": pq.Array(groupsIDS),
	})
	if err != nil {
		return Situation{}, false, err
	}
	defer rows.Close()

	var id int64
	var data string
	var factIDs []int64
	if rows.Next() {
		err := rows.Scan(&id, &data, pq.Array(&factIDs))
		if err != nil {
			return Situation{}, false, err
		}
	} else {
		return Situation{}, false, nil
	}

	var situation Situation
	err = json.Unmarshal([]byte(data), &situation)
	if err != nil {
		return Situation{}, false, err
	}
	situation.Facts = factIDs

	//This is necessary because within the definition we don't have the id
	situation.ID = id

	//Need to delete at any get of situation the universal token group
	situation.Groups = groups.DeleteTokenAllGroups(situation.Groups)

	return situation, true, nil
}

// Create creates a new situation in the database using the given situation object
func (r *PostgresRepository) Create(situation Situation) (int64, error) {
	isAllGroup := false
	for _, group := range situation.Groups {
		if group == groups.AllGroups {
			zap.L().Error("Situation shouldn't have the universal token group")
			isAllGroup = true
		}
	}

	if !isAllGroup {
		situation.Groups = append(situation.Groups, groups.AllGroups)
	}

	situationData, err := json.Marshal(situation)
	if err != nil {
		return -1, err
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()
	query := `INSERT INTO situation_definition_v1 (id, name, groups, definition, is_template, is_object, calendar_id, last_modified)
		VALUES (DEFAULT, :name, :groups, :definition, :is_template, :is_object, :calendar_id, :last_modified) RETURNING id`
	params := map[string]interface{}{
		"name":          situation.Name,
		"groups":        pq.Array(situation.Groups),
		"definition":    string(situationData),
		"is_template":   situation.IsTemplate,
		"is_object":     situation.IsObject,
		"calendar_id":   situation.CalendarID,
		"last_modified": timestamp,
	}

	if situation.CalendarID == 0 {
		params["calendar_id"] = nil
	}

	tx, err := r.conn.Beginx()
	if err != nil {
		return -1, err
	}

	rows, err := tx.NamedQuery(query, params)
	if err != nil {
		tx.Rollback()
		return -1, err
	}
	defer rows.Close()

	var id int64
	if rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			tx.Rollback()
			return -1, err
		}
		rows.Close()
	} else {
		tx.Rollback()
		return -1, errors.New("No id returning of insert situation")
	}

	err = r.updateSituationFacts(tx, id, situation.Facts)
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	return id, nil
}

// Update updates an entity in the repository by its name
func (r *PostgresRepository) Update(id int64, situation Situation) error {
	query := `UPDATE situation_definition_v1 SET name = :name, definition = :definition,
				is_template = :is_template, is_object = :is_object, calendar_id = :calendar_id,
				last_modified = :last_modified
	WHERE id = :id`

	//This is necessary because within the definition we don't have the id
	situation.ID = id

	situationData, err := json.Marshal(situation)
	if err != nil {
		return errors.New("Couldn't marshall the provided data" + err.Error())
	}

	t := time.Now().Truncate(1 * time.Millisecond).UTC()
	params := map[string]interface{}{
		"id":            id,
		"name":          situation.Name,
		"groups":        pq.Array(situation.Groups),
		"definition":    string(situationData),
		"is_template":   situation.IsTemplate,
		"is_object":     situation.IsObject,
		"calendar_id":   situation.CalendarID,
		"last_modified": t,
	}

	if situation.CalendarID == 0 {
		params["calendar_id"] = nil
	}

	tx, err := r.conn.Beginx()
	if err != nil {
		return err
	}

	res, err := tx.NamedExec(query, params)
	if err != nil {
		tx.Rollback()
		return errors.New("Couldn't query the database:" + err.Error())
	}
	i, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return errors.New("Error with the affected rows:" + err.Error())
	}
	if i != 1 {
		tx.Rollback()
		return errors.New("No row inserted (or multiple row inserted) instead of 1 row")
	}

	err = r.updateSituationFacts(tx, id, situation.Facts)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// Delete deletes an entity from the repository by its name
func (r *PostgresRepository) Delete(id int64) error {

	tx, err := r.conn.Beginx()
	if err != nil {
		return err
	}

	params := map[string]interface{}{"id": id}
	//Delete situations_facts
	_, err = tx.NamedExec(`DELETE FROM situation_facts_v1 WHERE situation_id = :id`, params)
	if err != nil {
		tx.Rollback()
		return err
	}
	//Delete situation rules
	_, err = tx.NamedExec(`DELETE FROM situation_rules_v1 WHERE situation_id = :id`, params)
	if err != nil {
		tx.Rollback()
		return err
	}
	//Delete situation definition
	res, err := tx.NamedExec(`DELETE FROM situation_definition_v1 WHERE id = :id`, params)
	if err != nil {
		tx.Rollback()
		return err
	}
	i, err := res.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	if i != 1 {
		tx.Rollback()
		return errors.New("No row deleted (or multiple row deleted) instead of 1 row")
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) updateSituationFacts(tx *sqlx.Tx, id int64, factIDs []int64) error {
	query := `DELETE FROM situation_facts_v1 WHERE situation_id = :id`
	var err error
	if tx != nil {
		_, err = tx.NamedExec(query, map[string]interface{}{"id": id})
	} else {
		_, err = r.conn.NamedExec(query, map[string]interface{}{"id": id})
	}
	if err != nil {
		zap.L().Error("Couldn't query the database:", zap.Error(err))
		return err
	}

	query = `INSERT INTO situation_facts_v1 (situation_id,fact_id) VALUES(:situationID, :factID)`
	for _, factID := range factIDs {
		if tx != nil {
			_, err = tx.NamedExec(query, map[string]interface{}{"situationID": id, "factID": factID})
		} else {
			_, err = r.conn.NamedExec(query, map[string]interface{}{"situationID": id, "factID": factID})
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) deleteSituationFacts(id int64) error {
	query := `DELETE FROM situation_facts_v1 WHERE situation_id = :id`
	_, err := r.conn.NamedExec(query, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		zap.L().Error("Couldn't query the database:", zap.Error(err))
		return err
	}

	return nil
}

// GetSituationsByFactID returns the situations in which the fact is required
func (r *PostgresRepository) GetSituationsByFactID(factID int64, ignoreIsObject bool) ([]Situation, error) {
	query := `SELECT situation_definition_v1.id, situation_definition_v1.definition,
	ARRAY(SELECT fact_id FROM situation_facts_v1 WHERE situation_id = situation_definition_v1.id) as fact_ids
	FROM situation_facts_v1 INNER JOIN situation_definition_v1
	ON situation_facts_v1.situation_id = situation_definition_v1.id
	WHERE situation_facts_v1.fact_id = :factID AND (:ignore_is_object = false OR is_object = false)`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"factID":           factID,
		"ignore_is_object": ignoreIsObject,
	})
	if err != nil {
		zap.L().Error("Couldn't query the database:", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	situations := make([]Situation, 0)
	for rows.Next() {
		var data string
		var situationID int64
		var situation Situation
		var factIDs []int64
		err := rows.Scan(&situationID, &data, pq.Array(&factIDs))
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(data), &situation)
		if err != nil {
			return nil, err
		}
		situation.Facts = factIDs
		//This is necessary because within the definition we don't have the id
		situation.ID = situationID

		//Need to delete at any get of situation the universal token group
		situation.Groups = groups.DeleteTokenAllGroups(situation.Groups)

		situations = append(situations, situation)
	}

	return situations, nil
}

// GetFacts returns the list of facts
func (r *PostgresRepository) GetFacts(id int64) ([]int64, error) {
	query := `SELECT fact_id FROM situation_facts_v1 WHERE situation_id = :id`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	factIDs := make([]int64, 0)
	for rows.Next() {
		var factID int64
		err := rows.Scan(&factID)
		if err != nil {
			return nil, err
		}
		factIDs = append(factIDs, factID)
	}

	return factIDs, nil
}

// GetAll returns all entities in the repository
func (r *PostgresRepository) GetAll(groupsIDS []int64) (map[int64]Situation, error) {
	query := `SELECT id, definition,
			ARRAY(SELECT fact_id FROM situation_facts_v1 WHERE situation_id = situation_definition_v1.id) as fact_ids
			FROM situation_definition_v1 WHERE groups && :groups`
	params := map[string]interface{}{
		"groups": pq.Array(groupsIDS),
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return parseAllRows(rows)
}

// GetAllByRuleID returns all entities in the repository based on a rule ID
func (r *PostgresRepository) GetAllByRuleID(groupsIDS []int64, ruleID int64) (map[int64]Situation, error) {

	query := `SELECT id, definition, ARRAY(SELECT fact_id FROM situation_facts_v1 WHERE situation_id = situation_definition_v1.id) as fact_ids
		FROM situation_definition_v1 INNER JOIN situation_rules_v1 ON situation_definition_v1.id = situation_rules_v1.situation_id
		WHERE groups && :groups AND situation_rules_v1.rule_id = :rule_id`
	params := map[string]interface{}{
		"groups":  pq.Array(groupsIDS),
		"rule_id": ruleID,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return parseAllRows(rows)
}

func parseAllRows(rows *sqlx.Rows) (map[int64]Situation, error) {
	situations := make(map[int64]Situation, 0)
	for rows.Next() {
		var data string
		var situationID int64
		var situation Situation
		var factIDs []int64
		err := rows.Scan(&situationID, &data, pq.Array(&factIDs))
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal([]byte(data), &situation)
		if err != nil {
			return nil, err
		}
		situation.Facts = factIDs
		//This is necessary because within the definition we don't have the id
		situation.ID = situationID

		//Need to delete at any get of situation the universal token group
		situation.Groups = groups.DeleteTokenAllGroups(situation.Groups)

		situations[situation.ID] = situation
	}
	return situations, nil
}

// IsInGroups returns true ins the situation is in one of the groups
func (r *PostgresRepository) IsInGroups(id int64, groups []int64) (bool, error) {
	var inGroup bool
	checkNameQuery := `SELECT groups && $1 FROM situation_definition_v1 WHERE id = $2;`
	err := r.conn.QueryRow(checkNameQuery, pq.Array(groups), id).Scan(&inGroup)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return inGroup, nil
}

// GetRules returns the list of rules used in to evaluate the situation
func (r *PostgresRepository) GetRules(id int64) ([]int64, error) {
	query := `SELECT rule_id FROM situation_rules_v1 WHERE situation_id = :id ORDER BY execution_order`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ruleIDs := make([]int64, 0)
	for rows.Next() {
		var ruleID int64
		err := rows.Scan(&ruleID)
		if err != nil {
			return nil, err
		}
		ruleIDs = append(ruleIDs, ruleID)
	}

	return ruleIDs, nil
}

// SetRules sets the list of rules for the situation evaluation
func (r *PostgresRepository) SetRules(id int64, rules []int64) error {
	query := `DELETE FROM situation_rules_v1 WHERE situation_id = :id`
	_, err := r.conn.NamedExec(query, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		zap.L().Error("Couldn't query the database:", zap.Error(err))
		return err
	}

	query = `INSERT INTO situation_rules_v1 (situation_id,rule_id, execution_order)
		VALUES(:situationID, :ruleID, :executionOrder)`
	for index, ruleID := range rules {
		_, err := r.conn.NamedExec(query, map[string]interface{}{
			"situationID":    id,
			"ruleID":         ruleID,
			"executionOrder": index,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// AddRule adds a rule ad the end of the situation rule list
func (r *PostgresRepository) AddRule(tx *sqlx.Tx, id int64, ruleID int64) error {
	query := `INSERT INTO situation_rules_v1 (situation_id , rule_id, execution_order)
				values (:situation_id, :rule_id, (SELECT COALESCE(MAX(execution_order) + 1 , 0) from situation_rules_v1 WHERE situation_id = :situation_id))`
	params := map[string]interface{}{
		"situation_id": id,
		"rule_id":      ruleID,
	}
	var err error
	if tx != nil {
		_, err = tx.NamedExec(query, params)
	} else {
		_, err = r.conn.NamedExec(query, params)
	}
	if err != nil {
		zap.L().Error("Couldn't query the database:", zap.Error(err))
		return err
	}

	return nil
}

// RemoveRule removes a rule ad the end of the situation rule list
func (r *PostgresRepository) RemoveRule(tx *sqlx.Tx, id int64, ruleID int64) error {
	query := `DELETE FROM situation_rules_v1 WHERE situation_id = :situation_id AND rule_id = :rule_id`
	params := map[string]interface{}{
		"situation_id": id,
		"rule_id":      ruleID,
	}
	var err error
	if tx != nil {
		_, err = tx.NamedExec(query, params)
	} else {
		_, err = r.conn.NamedExec(query, params)
	}
	if err != nil {
		zap.L().Error("Couldn't query the database:", zap.Error(err))
		return err
	}

	return nil
}

// CreateTemplateInstance creates a situation template instance
func (r *PostgresRepository) CreateTemplateInstance(situationID int64, instance TemplateInstance) (int64, error) {

	isTemplate, err := r.isTemplate(situationID)

	if err != nil {
		return -1, err
	}
	if !isTemplate {
		return -1, errors.New("The Situation does not exists or it is not a template situation")
	}

	parametersData, err := json.Marshal(instance.Parameters)
	if err != nil {
		return -1, err
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()
	query := `INSERT INTO situation_template_instances_v1 (situation_id, id, name, parameters, calendar_id, last_modified)
		VALUES (:situation_id, DEFAULT, :name, :parameters, :calendar_id, :last_modified) RETURNING id`
	params := map[string]interface{}{
		"situation_id":  situationID,
		"name":          instance.Name,
		"parameters":    string(parametersData),
		"calendar_id":   instance.CalendarID,
		"last_modified": timestamp,
	}

	if instance.CalendarID == 0 {
		params["calendar_id"] = nil
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

// UpdateTemplateInstance updates a situation template instance
func (r *PostgresRepository) UpdateTemplateInstance(instanceID int64, instance TemplateInstance) error {

	isTemplate, err := r.isTemplate(instance.SituationID)
	if err != nil {
		return err
	}
	if !isTemplate {
		return errors.New("The Situation does not exists or it is not a template situation")
	}

	query := `UPDATE situation_template_instances_v1 SET situation_id = :situation_id,
					name = :name, parameters = :parameters,
					calendar_id = :calendar_id,
					last_modified = :last_modified WHERE id = :id`

	//This is necessary because within the definition we don't have the id
	instance.ID = instanceID

	parametersData, err := json.Marshal(instance.Parameters)
	if err != nil {
		return errors.New("Couldn't marshall the provided situation parameters" + err.Error())
	}

	t := time.Now().Truncate(1 * time.Millisecond).UTC()

	params := map[string]interface{}{
		"id":            instanceID,
		"situation_id":  instance.SituationID,
		"name":          instance.Name,
		"parameters":    string(parametersData),
		"calendar_id":   instance.CalendarID,
		"last_modified": t,
	}

	if instance.CalendarID == 0 {
		params["calendar_id"] = nil
	}

	res, err := r.conn.NamedExec(query, params)

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

func (r *PostgresRepository) isTemplate(id int64) (bool, error) {

	query := `SELECT is_template FROM situation_definition_v1 WHERE id = :id`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id": id,
	})

	if err != nil {
		return false, err
	}
	defer rows.Close()

	var isTemplate bool
	if rows.Next() {
		err := rows.Scan(&isTemplate)
		if err != nil {
			return false, err
		}

		return isTemplate, nil
	}
	return false, nil
}

// DeleteTemplateInstance deletes a situation template instance
func (r *PostgresRepository) DeleteTemplateInstance(instanceID int64) error {
	query := `DELETE FROM situation_template_instances_v1 WHERE id = :id`

	res, err := r.conn.NamedExec(query, map[string]interface{}{
		"id": instanceID,
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

// GetTemplateInstance returns the situation template instance
func (r *PostgresRepository) GetTemplateInstance(instanceID int64) (TemplateInstance, bool, error) {
	query := `SELECT name, situation_id, parameters, calendar_id FROM situation_template_instances_v1 WHERE id = :id`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id": instanceID,
	})
	if err != nil {
		return TemplateInstance{}, false, err
	}
	defer rows.Close()

	if rows.Next() {
		var situationID int64
		var calendarID sql.NullInt64
		var name string
		var paramsData string
		err := rows.Scan(&name, &situationID, &paramsData, &calendarID)
		if err != nil {
			return TemplateInstance{}, false, err
		}
		templateInstance := TemplateInstance{
			ID:          instanceID,
			SituationID: situationID,
			Name:        name,
			CalendarID:  calendarID.Int64,
		}
		err = json.Unmarshal([]byte(paramsData), &templateInstance.Parameters)
		if err != nil {
			return TemplateInstance{}, false, err
		}

		return templateInstance, true, nil
	}

	return TemplateInstance{}, false, nil
}

// GetAllTemplateInstances returns the list of template instances of the situation
func (r *PostgresRepository) GetAllTemplateInstances(situationID int64) (map[int64]TemplateInstance, error) {
	query := `SELECT id, name, parameters, calendar_id FROM situation_template_instances_v1 WHERE situation_id = :situation_id`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"situation_id": situationID,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templateInstances := make(map[int64]TemplateInstance, 0)
	for rows.Next() {
		var id int64
		var name string
		var paramsData string
		var calendarID sql.NullInt64
		err := rows.Scan(&id, &name, &paramsData, &calendarID)
		if err != nil {
			return nil, err
		}
		templateInstance := TemplateInstance{
			ID:          id,
			SituationID: situationID,
			Name:        name,
			CalendarID:  calendarID.Int64,
		}
		err = json.Unmarshal([]byte(paramsData), &templateInstance.Parameters)
		if err != nil {
			return nil, err
		}
		templateInstances[templateInstance.ID] = templateInstance
	}

	return templateInstances, nil
}
