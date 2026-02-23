package situation

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/myrteametrics/myrtea-sdk/v5/repositories/utils"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const table = "situation_definition_v1"

// PostgresRepository is a repository containing the situation definition based on a PSQL database and
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

// scanSituation scans a row into a Situation struct
func (r *PostgresRepository) scanSituation(rows *sqlx.Rows, parseParameters bool) (Situation, error) {
	var situationID int64
	var data string
	var factIDs []int64
	err := rows.Scan(&situationID, &data, pq.Array(&factIDs))
	if err != nil {
		return Situation{}, err
	}
	var situation Situation
	err = json.Unmarshal([]byte(data), &situation)
	if err != nil {
		return Situation{}, err
	}
	situation.Facts = factIDs
	situation.ID = situationID
	if parseParameters {
		evalParameters(situation.Parameters)
	}
	// situation.Groups = groups.DeleteTokenAllGroups(situation.Groups)
	return situation, nil
}

// scanTemplateInstance scans a row into a TemplateInstance struct
func (r *PostgresRepository) scanTemplateInstance(rows *sql.Rows, parseParameters bool) (TemplateInstance, error) {
	var id int64
	var situationID int64
	var name string
	var paramsData string
	var calendarID sql.NullInt64
	var enableDependsOn bool
	var dependsOnParamsData string
	err := rows.Scan(&id, &name, &situationID, &paramsData, &calendarID, &enableDependsOn, &dependsOnParamsData)
	if err != nil {
		return TemplateInstance{}, err
	}
	templateInstance := TemplateInstance{
		ID:              id,
		SituationID:     situationID,
		Name:            name,
		CalendarID:      calendarID.Int64,
		EnableDependsOn: enableDependsOn,
	}
	err = json.Unmarshal([]byte(paramsData), &templateInstance.Parameters)
	if err != nil {
		return TemplateInstance{}, err
	}
	if parseParameters {
		evalParameters(templateInstance.Parameters)
	}
	err = json.Unmarshal([]byte(dependsOnParamsData), &templateInstance.DependsOnParameters)
	if err != nil {
		return TemplateInstance{}, err
	}
	return templateInstance, nil
}

// Get retrieve the specified situation definition
func (r *PostgresRepository) Get(id int64, parseParameters ...bool) (Situation, bool, error) {

	query := `SELECT id, definition,
				ARRAY(SELECT fact_id FROM situation_facts_v1 WHERE situation_id = :id) as fact_ids
				FROM situation_definition_v1 WHERE id = :id`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id": id,
	})

	if err != nil {
		return Situation{}, false, err
	}
	defer rows.Close()

	if rows.Next() {
		situation, err := r.scanSituation(rows, shouldParseForEvaluation(parseParameters...))
		if err != nil {
			return Situation{}, false, err
		}
		return situation, true, nil
	} else {
		return Situation{}, false, nil
	}
}

// GetByName retrieve the specified situation definition by it's name
// If parseParameters is true, evaluates situation parameters using Gval.
func (r *PostgresRepository) GetByName(name string, parseParameters ...bool) (Situation, bool, error) {

	query := `SELECT id, definition,
				ARRAY(SELECT fact_id FROM situation_facts_v1 WHERE situation_id = id) as fact_ids
				FROM situation_definition_v1 WHERE name = :name`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"name": name,
	})
	if err != nil {
		return Situation{}, false, err
	}
	defer rows.Close()

	if rows.Next() {
		situation, err := r.scanSituation(rows, shouldParseForEvaluation(parseParameters...))
		if err != nil {
			return Situation{}, false, err
		}
		return situation, true, nil
	} else {
		return Situation{}, false, nil
	}
}

func getSituationCalendarIdLinkValue(situation Situation) interface{} {
	if situation.CalendarID == 0 {
		return nil
	} else {
		return situation.CalendarID
	}
}

// Create creates a new situation in the database using the given situation object
func (r *PostgresRepository) Create(situation Situation) (int64, error) {

	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)
	situationData, err := json.Marshal(situation)
	if err != nil {
		return -1, err
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()

	tx, err := r.conn.Beginx()
	if err != nil {
		return -1, err
	}

	defer func() { _ = tx.Rollback() }()

	// Create a new statement builder
	statement := r.newStatement().
		Insert(table).
		Suffix("RETURNING \"id\"")

	// If situation.ID is provided, include it in the query
	if situation.ID != 0 {
		statement = statement.
			Columns("id", "name", "definition", "is_template", "is_object", "calendar_id", "last_modified").
			Values(situation.ID, situation.Name, string(situationData), situation.IsTemplate, situation.IsObject, getSituationCalendarIdLinkValue(situation), timestamp)
	} else {
		statement = statement.
			Columns("name", "definition", "is_template", "is_object", "calendar_id", "last_modified").
			Values(situation.Name, string(situationData), situation.IsTemplate, situation.IsObject, getSituationCalendarIdLinkValue(situation), timestamp)
	}

	// Execute the query and get the returned ID
	var id int64
	err = statement.RunWith(tx).QueryRow().Scan(&id)
	if err != nil {
		return -1, errors.New("couldn't query the database: " + err.Error())
	}

	err = r.updateSituationFacts(tx, id, situation.Facts)
	if err != nil {
		return -1, err
	}

	err = tx.Commit()
	if err != nil {
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
		return errors.New("couldn't marshall the provided data" + err.Error())
	}

	t := time.Now().Truncate(1 * time.Millisecond).UTC()
	params := map[string]interface{}{
		"id":            id,
		"name":          situation.Name,
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

	defer func() { _ = tx.Rollback() }()

	res, err := tx.NamedExec(query, params)
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

	err = r.updateSituationFacts(tx, id, situation.Facts)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
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

	defer func() { _ = tx.Rollback() }()

	params := map[string]interface{}{"id": id}
	//Delete situations_facts
	_, err = tx.NamedExec(`DELETE FROM situation_facts_v1 WHERE situation_id = :id`, params)
	if err != nil {
		return err
	}
	//Delete situation rules
	_, err = tx.NamedExec(`DELETE FROM situation_rules_v1 WHERE situation_id = :id`, params)
	if err != nil {
		return err
	}
	//Delete situation definition
	res, err := tx.NamedExec(`DELETE FROM situation_definition_v1 WHERE id = :id`, params)
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
	err = tx.Commit()
	if err != nil {
		return err
	}
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)
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
// If parseParameters is true, evaluates situation parameters using Gval.
func (r *PostgresRepository) GetSituationsByFactID(factID int64, ignoreIsObject bool, parseParameters ...bool) ([]Situation, error) {

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
		situation, err := r.scanSituation(rows, shouldParseForEvaluation(parseParameters...))
		if err != nil {
			return nil, err
		}

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
// If parseParameters is true, evaluates situation parameters using Gval.
func (r *PostgresRepository) GetAll(parseParameters ...bool) (map[int64]Situation, error) {

	query := `SELECT id, definition,
			ARRAY(SELECT fact_id FROM situation_facts_v1 WHERE situation_id = situation_definition_v1.id) as fact_ids
			FROM situation_definition_v1`
	params := map[string]interface{}{}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.parseAllRows(rows, shouldParseForEvaluation(parseParameters...))
}

// GetAllByIDs returns all entities filtered by IDs in the repository
// If parseParameters is true, evaluates situation parameters using Gval.
func (r *PostgresRepository) GetAllByIDs(ids []int64, parseParameters ...bool) (map[int64]Situation, error) {

	query := `SELECT id, definition,
			ARRAY(SELECT fact_id FROM situation_facts_v1 WHERE situation_id = situation_definition_v1.id) as fact_ids
			FROM situation_definition_v1 WHERE id = ANY(:ids)`
	params := map[string]interface{}{
		"ids": pq.Array(ids),
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.parseAllRows(rows, shouldParseForEvaluation(parseParameters...))
}

// GetAllByRuleID returns all entities in the repository based on a rule ID
// If parseParameters is true, evaluates situation parameters using Gval.
func (r *PostgresRepository) GetAllByRuleID(ruleID int64, parseParameters ...bool) (map[int64]Situation, error) {

	query := `SELECT id, definition, ARRAY(SELECT fact_id FROM situation_facts_v1 WHERE situation_id = situation_definition_v1.id) as fact_ids
		FROM situation_definition_v1 INNER JOIN situation_rules_v1 ON situation_definition_v1.id = situation_rules_v1.situation_id
		WHERE situation_rules_v1.rule_id = :rule_id`
	params := map[string]interface{}{
		"rule_id": ruleID,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.parseAllRows(rows, shouldParseForEvaluation(parseParameters...))
}

func (r *PostgresRepository) parseAllRows(rows *sqlx.Rows, parseParameters bool) (map[int64]Situation, error) {
	situations := make(map[int64]Situation)
	for rows.Next() {
		situation, err := r.scanSituation(rows, parseParameters)
		if err != nil {
			return nil, err
		}

		situations[situation.ID] = situation
	}
	return situations, nil
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
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, "situation_template_instances_v1")
	isTemplate, err := r.isTemplate(situationID)

	if err != nil {
		return -1, err
	}
	if !isTemplate {
		return -1, errors.New("the Situation does not exists or it is not a template situation")
	}

	parametersData, err := json.Marshal(instance.Parameters)
	if err != nil {
		return -1, err
	}

	dependsOnParametresData, err := json.Marshal(instance.DependsOnParameters)
	if err != nil {
		return -1, err
	}

	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()
	query := `INSERT INTO situation_template_instances_v1 (situation_id, id, name, parameters, calendar_id, last_modified, enable_depends_on, depends_on_parameters)
		VALUES (:situation_id, DEFAULT, :name, :parameters, :calendar_id, :last_modified, :enable_depends_on, :depends_on_parameters) RETURNING id`
	if instance.ID != 0 {
		query = `INSERT INTO situation_template_instances_v1 (situation_id, id, name, parameters, calendar_id, last_modified, enable_depends_on, depends_on_parameters)
		VALUES (:situation_id, :id, :name, :parameters, :calendar_id, :last_modified, :enable_depends_on, :depends_on_parameters) RETURNING id`
	}

	params := map[string]interface{}{
		"situation_id":          situationID,
		"id":                    instance.ID,
		"name":                  instance.Name,
		"parameters":            string(parametersData),
		"calendar_id":           instance.CalendarID,
		"last_modified":         timestamp,
		"enable_depends_on":     instance.EnableDependsOn,
		"depends_on_parameters": string(dependsOnParametresData),
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
		return -1, errors.New("no id returning of insert situation")
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
		return errors.New("the Situation does not exists or it is not a template situation")
	}

	query := `UPDATE situation_template_instances_v1 SET situation_id = :situation_id,
					name = :name, parameters = :parameters,
					calendar_id = :calendar_id,
					last_modified = :last_modified,
					enable_depends_on = :enable_depends_on,
					depends_on_parameters = :depends_on_parameters WHERE id = :id`

	//This is necessary because within the definition we don't have the id
	instance.ID = instanceID

	parametersData, err := json.Marshal(instance.Parameters)
	if err != nil {
		return errors.New("couldn't marshall the provided situation parameters" + err.Error())
	}

	dependsOnParametresData, err := json.Marshal(instance.DependsOnParameters)
	if err != nil {
		return errors.New("couldn't marshall the provided situation dependsOn Parameters" + err.Error())
	}

	t := time.Now().Truncate(1 * time.Millisecond).UTC()

	params := map[string]interface{}{
		"id":                    instanceID,
		"situation_id":          instance.SituationID,
		"name":                  instance.Name,
		"parameters":            string(parametersData),
		"calendar_id":           instance.CalendarID,
		"last_modified":         t,
		"enable_depends_on":     instance.EnableDependsOn,
		"depends_on_parameters": string(dependsOnParametresData),
	}

	if instance.CalendarID == 0 {
		params["calendar_id"] = nil
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
		return errors.New("no row deleted (or multiple row deleted) instead of 1 row")
	}
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, "situation_template_instances_v1")
	return nil
}

// GetTemplateInstance returns the situation template instance
// If parseParameters is true, evaluates situation parameters using Gval.
func (r *PostgresRepository) GetTemplateInstance(instanceID int64, parseParameters ...bool) (TemplateInstance, bool, error) {

	rows, err := r.newStatement().
		Select("id", "name", "situation_id", "parameters", "calendar_id", "enable_depends_on", "depends_on_parameters").
		From("situation_template_instances_v1").
		Where(sq.Eq{"id": instanceID}).
		Query()
	if err != nil {
		return TemplateInstance{}, false, err
	}
	defer rows.Close()

	if rows.Next() {
		templateInstance, err := r.scanTemplateInstance(rows, shouldParseForEvaluation(parseParameters...))
		if err != nil {
			return TemplateInstance{}, false, err
		}

		return templateInstance, true, nil
	}

	return TemplateInstance{}, false, nil
}

// GetAllTemplateInstances returns the list of template instances of the situation
// If parseParameters is true, the situation or situation instance parameters are evaluated using Gval.
func (r *PostgresRepository) GetAllTemplateInstances(situationID int64, parseParameters ...bool) (map[int64]TemplateInstance, error) {
	rows, err := r.newStatement().
		Select(
			"id",
			"name",
			"situation_id",
			"parameters",
			"calendar_id",
			"enable_depends_on",
			"depends_on_parameters",
		).
		From("situation_template_instances_v1").
		Where(sq.Eq{"situation_id": situationID}).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templateInstances := make(map[int64]TemplateInstance)
	for rows.Next() {
		templateInstance, err := r.scanTemplateInstance(rows, shouldParseForEvaluation(parseParameters...))
		if err != nil {
			return nil, err
		}
		templateInstances[templateInstance.ID] = templateInstance
	}

	return templateInstances, nil
}

// GetAllTemplateInstancesByIDs returns a map of template instances by their IDs
// If parseParameters is true, the situation instance parameters are evaluated using Gval.
func (r *PostgresRepository) GetAllTemplateInstancesByIDs(ids []int64, parseParameters ...bool) (map[int64]TemplateInstance, error) {
	if len(ids) == 0 {
		return make(map[int64]TemplateInstance), nil
	}

	rows, err := r.newStatement().
		Select(
			"id",
			"name",
			"situation_id",
			"parameters",
			"calendar_id",
			"enable_depends_on",
			"depends_on_parameters",
		).
		From("situation_template_instances_v1").
		Where(sq.Eq{"id": ids}).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templateInstances := make(map[int64]TemplateInstance)
	for rows.Next() {
		templateInstance, err := r.scanTemplateInstance(rows, shouldParseForEvaluation(parseParameters...))
		if err != nil {
			return nil, err
		}
		templateInstances[templateInstance.ID] = templateInstance
	}

	return templateInstances, nil
}

// GetSituationOverview returns an overview of situations + their template instances
func (r *PostgresRepository) GetSituationOverview() ([]SituationOverview, error) {
	// Build query using Squirrel
	query := r.newStatement().
		Select(
			"sd.id as situation_id",
			"sd.name as situation_name",
			"sd.is_template",
			"COALESCE(ARRAY_AGG(DISTINCT ts.tag_id) FILTER (WHERE ts.tag_id IS NOT NULL), '{}') as situation_tags",
			"sti.id as instance_id",
			"sti.name as instance_name",
			"COALESCE(ARRAY_AGG(DISTINCT tsti.tag_id) FILTER (WHERE tsti.tag_id IS NOT NULL), '{}') as instance_tags",
		).
		From("situation_definition_v1 sd").
		LeftJoin("tags_situations_v1 ts ON sd.id = ts.situation_id").
		LeftJoin("situation_template_instances_v1 sti ON sd.id = sti.situation_id").
		LeftJoin("tags_situation_template_instances_v1 tsti ON sti.id = tsti.situation_template_instance_id").
		GroupBy("sd.id", "sd.name", "sd.is_template", "sti.id", "sti.name").
		OrderBy("sd.id", "sti.id")

	rows, err := query.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map to build the hierarchical structure
	situationMap := make(map[int64]*SituationOverview)
	var result []SituationOverview

	for rows.Next() {
		var situationID int64
		var situationName string
		var isTemplate bool
		var situationTags pq.Int64Array
		var instanceID sql.NullInt64
		var instanceName sql.NullString
		var instanceTags pq.Int64Array

		err := rows.Scan(&situationID, &situationName, &isTemplate, &situationTags,
			&instanceID, &instanceName, &instanceTags)
		if err != nil {
			return nil, err
		}

		// Get or create the situation overview
		situation, exists := situationMap[situationID]
		if !exists {
			situation = &SituationOverview{
				ID:                situationID,
				Name:              situationName,
				IsTemplate:        isTemplate,
				Tags:              situationTags,
				InstanceTemplates: []SituationOverview{},
			}
			situationMap[situationID] = situation
		}

		// Add template instance if it exists
		if instanceID.Valid {
			instance := SituationOverview{
				ID:         instanceID.Int64,
				Name:       instanceName.String,
				IsTemplate: false,
				Tags:       []int64(instanceTags),
			}
			situation.InstanceTemplates = append(situation.InstanceTemplates, instance)
		}
	}

	// Convert map to slice
	for _, situation := range situationMap {
		result = append(result, *situation)
	}

	return result, nil
}

// GetAllTemplateInstancesByRuleID returns all template instances associated with situations that use the specified rule ID
func (r *PostgresRepository) GetAllTemplateInstancesByRuleID(ruleID int64, parseParameters ...bool) (map[int64]TemplateInstance, error) {
	rows, err := r.newStatement().
		Select(
			"sti.id",
			"sti.situation_id",
			"sti.name",
			"sti.parameters",
			"sti.calendar_id",
			"sti.enable_depends_on",
			"sti.depends_on_parameters",
		).
		From("situation_template_instances_v1 sti").
		Join("situation_rules_v1 sr ON sti.situation_id = sr.situation_id").
		Where(sq.Eq{"sr.rule_id": ruleID}).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templateInstances := make(map[int64]TemplateInstance)
	for rows.Next() {
		templateInstance, err := r.scanTemplateInstance(rows, shouldParseForEvaluation(parseParameters...))
		if err != nil {
			return nil, err
		}
		templateInstances[templateInstance.ID] = templateInstance
	}

	return templateInstances, nil
}

func (r *PostgresRepository) GetSituationsWithInstances(parseParameters ...bool) ([]SituationWithInstances, error) {
	parse := shouldParseForEvaluation(parseParameters...)

	rows, err := r.newStatement().
		Select("id", "definition").
		From("situation_definition_v1").
		OrderBy("id").
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	situationIDs := make([]int64, 0)
	situationMap := make(map[int64]*SituationWithInstances)

	for rows.Next() {
		var id int64
		var def string

		if err := rows.Scan(&id, &def); err != nil {
			return nil, err
		}

		var s Situation
		if err := json.Unmarshal([]byte(def), &s); err != nil {
			return nil, err
		}

		s.ID = id
		s.Facts = nil

		if parse {
			evalParameters(s.Parameters)
		}

		situationIDs = append(situationIDs, s.ID)
		situationMap[s.ID] = &SituationWithInstances{
			Situation: s,
			Instances: []TemplateInstance{},
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(situationIDs) == 0 {
		return []SituationWithInstances{}, nil
	}

	instRows, err := r.newStatement().
		Select("id", "name", "situation_id", "parameters", "calendar_id", "enable_depends_on", "depends_on_parameters").
		From("situation_template_instances_v1").
		Where(sq.Eq{"situation_id": situationIDs}).
		OrderBy("situation_id", "id").
		Query()
	if err != nil {
		return nil, err
	}
	defer instRows.Close()

	for instRows.Next() {
		inst, err := r.scanTemplateInstance(instRows, parse)
		if err != nil {
			return nil, err
		}
		if parent, ok := situationMap[inst.SituationID]; ok {
			parent.Instances = append(parent.Instances, inst)
		}
	}
	if err := instRows.Err(); err != nil {
		return nil, err
	}

	result := make([]SituationWithInstances, 0, len(situationMap))
	for _, s := range situationMap {
		result = append(result, *s)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Situation.ID < result[j].Situation.ID })

	return result, nil
}
