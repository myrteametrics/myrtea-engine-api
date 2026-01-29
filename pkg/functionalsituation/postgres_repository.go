package functionalsituation

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/utils/dbutils"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/utils"
	"go.uber.org/zap"
)

const table = "functional_situation_v1"
const tableInstances = "functional_situation_instances_v1"
const tableSituations = "functional_situation_situations_v1"
const tableInstanceRef = "functional_situation_instance_ref_v1"
const tableSituationRef = "functional_situation_situation_ref_v1"

var fields = []string{"id", "name", "description", "parent_id", "color", "icon", "created_at", "updated_at", "created_by", "parameters"}

// PostgresRepository is a repository containing the Functional Situation data based on a PSQL database
// and implementing the repository interface
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

// scan scans a row into a FunctionalSituation struct
func (r *PostgresRepository) scan(rows *sql.Rows) (FunctionalSituation, error) {
	var fs FunctionalSituation
	var parametersJSON []byte
	err := rows.Scan(&fs.ID, &fs.Name, &fs.Description, &fs.ParentID, &fs.Color, &fs.Icon,
		&fs.CreatedAt, &fs.UpdatedAt, &fs.CreatedBy, &parametersJSON)
	if err != nil {
		return FunctionalSituation{}, err
	}

	if len(parametersJSON) > 0 {
		err = json.Unmarshal(parametersJSON, &fs.Parameters)
		if err != nil {
			zap.L().Error("Error unmarshaling parameters", zap.Error(err))
			fs.Parameters = make(map[string]interface{})
		}
	}

	return fs, nil
}

// Create inserts a new functional situation in the database
func (r *PostgresRepository) Create(fs FunctionalSituation, createdBy string) (int64, error) {
	if ok, err := fs.IsValid(); !ok {
		return -1, err
	}

	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)

	var id int64
	now := time.Now()

	parametersJSON, err := json.Marshal(fs.Parameters)
	if err != nil {
		return -1, fmt.Errorf("error marshaling parameters: %w", err)
	}

	statement := r.newStatement().
		Insert(table).
		Columns("name", "description", "parent_id", "color", "icon", "created_at", "updated_at", "created_by", "parameters").
		Values(fs.Name, fs.Description, fs.ParentID, fs.Color, fs.Icon, now, now, createdBy, parametersJSON).
		Suffix("RETURNING \"id\"")

	err = statement.QueryRow().Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

// Get retrieves a functional situation by its ID
func (r *PostgresRepository) Get(id int64) (FunctionalSituation, bool, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where(sq.Eq{"id": id}).
		Query()
	if err != nil {
		return FunctionalSituation{}, false, err
	}
	defer rows.Close()

	return dbutils.ScanFirst(rows, r.scan)
}

// GetByName retrieves a functional situation by its name and parent ID
func (r *PostgresRepository) GetByName(name string, parentID *int64) (FunctionalSituation, bool, error) {
	// Handle NULL parent_id case (use COALESCE to match the unique constraint)
	var parentCondition sq.Sqlizer
	if parentID == nil {
		parentCondition = sq.Expr("parent_id IS NULL")
	} else {
		parentCondition = sq.Eq{"parent_id": *parentID}
	}

	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where(sq.And{sq.Eq{"name": name}, parentCondition}).
		Query()
	if err != nil {
		return FunctionalSituation{}, false, err
	}
	defer rows.Close()

	return dbutils.ScanFirst(rows, r.scan)
}

// Update updates a functional situation
func (r *PostgresRepository) Update(id int64, fs FunctionalSituationUpdate, updatedBy string) error {
	statement := r.newStatement().Update(table)

	if fs.Name != nil {
		statement = statement.Set("name", *fs.Name)
	}
	if fs.Description != nil {
		statement = statement.Set("description", *fs.Description)
	}
	if fs.ParentID != nil {
		// Use -1 as sentinel value to set parent_id to NULL
		if *fs.ParentID == -1 {
			statement = statement.Set("parent_id", nil)
		} else {
			statement = statement.Set("parent_id", *fs.ParentID)
		}
	}
	if fs.Color != nil {
		statement = statement.Set("color", *fs.Color)
	}
	if fs.Icon != nil {
		statement = statement.Set("icon", *fs.Icon)
	}
	if fs.Parameters != nil {
		parametersJSON, err := json.Marshal(*fs.Parameters)
		if err != nil {
			return fmt.Errorf("error marshaling parameters: %w", err)
		}
		statement = statement.Set("parameters", parametersJSON)
	}

	statement = statement.Set("updated_at", time.Now())
	statement = statement.Set("created_by", updatedBy) // Track who updated

	_, err := statement.Where(sq.Eq{"id": id}).Exec()
	return err
}

// Delete removes a functional situation from the database
func (r *PostgresRepository) Delete(id int64) error {
	_, err := r.newStatement().
		Delete(table).
		Where(sq.Eq{"id": id}).
		Exec()
	if err != nil {
		return err
	}
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)
	return nil
}

// GetAll retrieves all functional situations
func (r *PostgresRepository) GetAll() ([]FunctionalSituation, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		OrderBy("name").
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dbutils.ScanAll(rows, r.scan)
}

// GetChildren retrieves all direct children of a functional situation
func (r *PostgresRepository) GetChildren(parentID int64) ([]FunctionalSituation, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where(sq.Eq{"parent_id": parentID}).
		OrderBy("name").
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dbutils.ScanAll(rows, r.scan)
}

// GetRoots retrieves all root functional situations (those without a parent)
func (r *PostgresRepository) GetRoots() ([]FunctionalSituation, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where(sq.Expr("parent_id IS NULL")).
		OrderBy("name").
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dbutils.ScanAll(rows, r.scan)
}

// GetTree retrieves the complete hierarchy as a flat list ordered by depth
func (r *PostgresRepository) GetTree() ([]FunctionalSituation, error) {
	query := `
		WITH RECURSIVE fs_tree AS (
			SELECT id, name, description, parent_id, color, icon, parameters, created_at, updated_at, created_by, 0 as depth
			FROM functional_situation_v1
			WHERE parent_id IS NULL
			UNION ALL
			SELECT fs.id, fs.name, fs.description, fs.parent_id, fs.color, fs.icon, fs.parameters, fs.created_at, fs.updated_at, fs.created_by, ft.depth + 1
			FROM functional_situation_v1 fs
			INNER JOIN fs_tree ft ON fs.parent_id = ft.id
		)
		SELECT id, name, description, parent_id, color, icon, created_at, updated_at, created_by, parameters
		FROM fs_tree
		ORDER BY depth, name
	`

	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dbutils.ScanAll(rows, r.scan)
}

// GetAncestors retrieves all ancestors of a functional situation (from parent to root)
func (r *PostgresRepository) GetAncestors(id int64) ([]FunctionalSituation, error) {
	query := `
		WITH RECURSIVE ancestors AS (
			SELECT id, name, description, parent_id, color, icon, parameters, created_at, updated_at, created_by, 0 as depth
			FROM functional_situation_v1
			WHERE id = $1
			UNION ALL
			SELECT fs.id, fs.name, fs.description, fs.parent_id, fs.color, fs.icon, fs.parameters, fs.created_at, fs.updated_at, fs.created_by, a.depth + 1
			FROM functional_situation_v1 fs
			INNER JOIN ancestors a ON fs.id = a.parent_id
		)
		SELECT id, name, description, parent_id, color, icon, created_at, updated_at, created_by, parameters
		FROM ancestors
		WHERE id != $1
		ORDER BY depth DESC
	`

	rows, err := r.conn.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dbutils.ScanAll(rows, r.scan)
}

// AddTemplateInstance creates an association between a functional situation and a template instance
// If the instance already has a reference, it links to the existing reference
// If not, it creates a new reference with the provided parameters
func (r *PostgresRepository) AddTemplateInstance(fsID int64, instanceID int64, parameters map[string]interface{}, addedBy string) error {
	// First, ensure the reference exists (create if not)
	err := r.ensureInstanceReference(instanceID, parameters, addedBy)
	if err != nil {
		return err
	}

	// Then create the link
	_, err = r.newStatement().
		Insert(tableInstances).
		Columns("functional_situation_id", "template_instance_id", "added_at", "added_by").
		Values(fsID, instanceID, time.Now(), addedBy).
		Exec()
	return err
}

// AddTemplateInstancesBulk creates associations between a functional situation and multiple template instances
// This is more efficient than calling AddTemplateInstance multiple times
func (r *PostgresRepository) AddTemplateInstancesBulk(fsID int64, instances []InstanceReference, addedBy string) error {
	if len(instances) == 0 {
		return nil
	}

	// Validate all instances first
	for _, instance := range instances {
		if ok, err := instance.IsValid(); !ok {
			return err
		}
	}

	// First, ensure all references exist
	for _, instance := range instances {
		err := r.ensureInstanceReference(instance.TemplateInstanceID, instance.Parameters, addedBy)
		if err != nil {
			return err
		}
	}

	// Then create all links in bulk
	insertStatement := r.newStatement().
		Insert(tableInstances).
		Columns("functional_situation_id", "template_instance_id", "added_at", "added_by")

	now := time.Now()
	for _, instance := range instances {
		insertStatement = insertStatement.Values(fsID, instance.TemplateInstanceID, now, addedBy)
	}

	_, err := insertStatement.Exec()
	return err
}

// ensureInstanceReference creates a reference for a template instance if it doesn't exist
func (r *PostgresRepository) ensureInstanceReference(instanceID int64, parameters map[string]interface{}, createdBy string) error {
	// Check if reference already exists
	var exists bool
	r.newStatement().Select()
	err := r.conn.QueryRow("SELECT EXISTS(SELECT 1 FROM "+tableInstanceRef+" WHERE template_instance_id = $1)", instanceID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Reference already exists - nothing to do (parameters are already set)
		return nil
	}

	// Create new reference
	parametersJSON, err := json.Marshal(parameters)
	if err != nil {
		return fmt.Errorf("error marshaling parameters: %w", err)
	}

	_, err = r.newStatement().
		Insert(tableInstanceRef).
		Columns("template_instance_id", "parameters", "created_at", "created_by").
		Values(instanceID, parametersJSON, time.Now(), createdBy).
		Exec()
	return err
}

// GetInstanceReference retrieves the parameters for a template instance reference
func (r *PostgresRepository) GetInstanceReference(instanceID int64) (InstanceReference, bool, error) {
	var ref InstanceReference
	var parametersJSON []byte

	rows, err := r.newStatement().Select("template_instance_id", "parameters").
		From(tableInstanceRef).
		Where(sq.Eq{"template_instance_id": instanceID}).
		Query()
	if err != nil {
		return InstanceReference{}, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return InstanceReference{}, false, nil
	}

	err = rows.Scan(&ref.TemplateInstanceID, &parametersJSON)
	if err != nil {
		return InstanceReference{}, false, err
	}

	if len(parametersJSON) > 0 {
		err = json.Unmarshal(parametersJSON, &ref.Parameters)
		if err != nil {
			zap.L().Error("Error unmarshaling instance parameters", zap.Error(err))
			ref.Parameters = make(map[string]interface{})
		}
	}

	return ref, true, nil
}

// UpdateInstanceReferenceParameters updates the parameters for an instance reference
func (r *PostgresRepository) UpdateInstanceReferenceParameters(instanceID int64, parameters map[string]interface{}) error {
	parametersJSON, err := json.Marshal(parameters)
	if err != nil {
		return fmt.Errorf("error marshaling parameters: %w", err)
	}

	_, err = r.newStatement().
		Update(tableInstanceRef).
		Set("parameters", parametersJSON).
		Where(sq.Eq{"template_instance_id": instanceID}).
		Exec()
	return err
}

// RemoveTemplateInstance removes an association between a functional situation and a template instance
// If no other functional situation references this instance, the reference is also deleted
func (r *PostgresRepository) RemoveTemplateInstance(fsID int64, instanceID int64) error {
	// First, remove the link
	_, err := r.newStatement().
		Delete(tableInstances).
		Where(sq.Eq{"functional_situation_id": fsID, "template_instance_id": instanceID}).
		Exec()
	if err != nil {
		return err
	}

	// Check if any other functional situation still references this instance
	var count int
	rows, err := r.newStatement().
		Select("COUNT(*)").
		From(tableInstances).
		Where(sq.Eq{"template_instance_id": instanceID}).
		Query()
	if err != nil {
		return err
	}

	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return err
		}
	}

	// If no more references, delete the instance reference
	if count == 0 {
		_, err = r.newStatement().
			Delete(tableInstanceRef).
			Where(sq.Eq{"template_instance_id": instanceID}).
			Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveTemplateInstancesBySituation removes all template instances of a given situation from a functional situation
// This is more efficient than calling RemoveTemplateInstance multiple times
// It also handles cleanup of orphaned references
func (r *PostgresRepository) RemoveTemplateInstancesBySituation(fsID int64, situationID int64) error {
	// First, get all template instance IDs for this situation that are linked to this FS
	var instanceIDs []int64
	rows, err := r.newStatement().Select("fsi.template_instance_id").
		From(fmt.Sprintf("%s fsi", tableInstances)).
		InnerJoin("situation_template_instances_v1 sti ON fsi.template_instance_id = sti.id").
		Where(sq.And{
			sq.Eq{"fsi.functional_situation_id": fsID},
			sq.Eq{"sti.situation_id": situationID},
		}).Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var instanceID int64
		if err := rows.Scan(&instanceID); err != nil {
			return err
		}
		instanceIDs = append(instanceIDs, instanceID)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	// If no instances found, nothing to do
	if len(instanceIDs) == 0 {
		return nil
	}

	// Delete all links from functional_situation_instances_v1 in one query
	_, err = r.newStatement().
		Delete(tableInstances).
		Where(sq.And{
			sq.Eq{"functional_situation_id": fsID},
			sq.Eq{"template_instance_id": instanceIDs},
		}).
		Exec()
	if err != nil {
		return err
	}

	// Clean up orphaned references
	// Find instance IDs that are no longer referenced by any functional situation
	orphanedRows, err := r.newStatement().Select("fsi.template_instance_id").
		From(fmt.Sprintf("%s fsi", tableInstances)).
		Where(sq.And{
			sq.Eq{"fsi.template_instance_id": instanceIDs},
			r.newStatement().
				Select("1").
				From(fmt.Sprintf("%s fsi", tableInstances)).
				Where("fsi.template_instance_id = fir.template_instance_id").
				Prefix("NOT EXISTS (").
				Suffix(")"),
		}).Query()
	if err != nil {
		return err
	}
	defer orphanedRows.Close()

	var orphanedIDs []int64
	for orphanedRows.Next() {
		var orphanedID int64
		if err = orphanedRows.Scan(&orphanedID); err != nil {
			return err
		}
		orphanedIDs = append(orphanedIDs, orphanedID)
	}

	if err = orphanedRows.Err(); err != nil {
		return err
	}

	// Delete orphaned references
	if len(orphanedIDs) > 0 {
		_, err = r.newStatement().
			Delete(tableInstanceRef).
			Where(sq.Eq{"template_instance_id": orphanedIDs}).
			Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetTemplateInstances retrieves all template instance IDs associated with a functional situation
func (r *PostgresRepository) GetTemplateInstances(fsID int64) ([]int64, error) {
	rows, err := r.newStatement().
		Select("template_instance_id").
		From(tableInstances).
		Where(sq.Eq{"functional_situation_id": fsID}).
		OrderBy("template_instance_id").
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetTemplateInstancesWithParameters retrieves all template instance IDs with their parameters for a functional situation in one query
func (r *PostgresRepository) GetTemplateInstancesWithParameters(fsID int64) (map[int64]map[string]interface{}, error) {
	rows, err := r.newStatement().
		Select("fsi.template_instance_id", "COALESCE(ref.parameters, '{}'::jsonb)").
		From(tableInstances + " fsi").
		InnerJoin(tableInstanceRef + " ref ON ref.template_instance_id = fsi.template_instance_id").
		Where(sq.Eq{"fsi.functional_situation_id": fsID}).
		OrderBy("fsi.template_instance_id").
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]map[string]interface{})
	for rows.Next() {
		var instanceID int64
		var parametersJSON []byte

		if err := rows.Scan(&instanceID, &parametersJSON); err != nil {
			return nil, err
		}

		var parameters map[string]interface{}
		if len(parametersJSON) > 0 {
			err = json.Unmarshal(parametersJSON, &parameters)
			if err != nil {
				zap.L().Error("Error unmarshaling instance parameters", zap.Error(err))
				parameters = make(map[string]interface{})
			}
		}

		result[instanceID] = parameters
	}

	return result, rows.Err()
}

// AddSituation creates an association between a functional situation and a situation
// If the situation already has a reference, it links to the existing reference
// If not, it creates a new reference with the provided parameters
func (r *PostgresRepository) AddSituation(fsID int64, situationID int64, parameters map[string]interface{}, addedBy string) error {
	// First, ensure the reference exists (create if not)
	err := r.ensureSituationReference(situationID, parameters, addedBy)
	if err != nil {
		return err
	}

	// Then create the link
	_, err = r.newStatement().
		Insert(tableSituations).
		Columns("functional_situation_id", "situation_id", "added_at", "added_by").
		Values(fsID, situationID, time.Now(), addedBy).
		Exec()
	return err
}

// AddSituationsBulk creates associations between a functional situation and multiple situations
// This is more efficient than calling AddSituation multiple times
func (r *PostgresRepository) AddSituationsBulk(fsID int64, situations []SituationReference, addedBy string) error {
	if len(situations) == 0 {
		return nil
	}

	// Validate all situations first
	for _, situation := range situations {
		if ok, err := situation.IsValid(); !ok {
			return err
		}
	}

	// First, ensure all references exist
	for _, situation := range situations {
		err := r.ensureSituationReference(situation.SituationID, situation.Parameters, addedBy)
		if err != nil {
			return err
		}
	}

	// Then create all links in bulk
	insertStatement := r.newStatement().
		Insert(tableSituations).
		Columns("functional_situation_id", "situation_id", "added_at", "added_by")

	now := time.Now()
	for _, situation := range situations {
		insertStatement = insertStatement.Values(fsID, situation.SituationID, now, addedBy)
	}

	_, err := insertStatement.Exec()
	return err
}

// ensureSituationReference creates a reference for a situation if it doesn't exist
func (r *PostgresRepository) ensureSituationReference(situationID int64, parameters map[string]interface{}, createdBy string) error {
	// Check if reference already exists
	var exists bool
	err := r.conn.QueryRow("SELECT EXISTS(SELECT 1 FROM "+tableSituationRef+" WHERE situation_id = $1)", situationID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		// Reference already exists - nothing to do (parameters are already set)
		return nil
	}

	// Create new reference
	parametersJSON, err := json.Marshal(parameters)
	if err != nil {
		return fmt.Errorf("error marshaling parameters: %w", err)
	}

	_, err = r.newStatement().
		Insert(tableSituationRef).
		Columns("situation_id", "parameters", "created_at", "created_by").
		Values(situationID, parametersJSON, time.Now(), createdBy).
		Exec()
	return err
}

// GetSituationReference retrieves the parameters for a situation reference
func (r *PostgresRepository) GetSituationReference(situationID int64) (SituationReference, bool, error) {
	var ref SituationReference
	var parametersJSON []byte

	err := r.conn.QueryRow("SELECT situation_id, parameters FROM "+tableSituationRef+" WHERE situation_id = $1", situationID).
		Scan(&ref.SituationID, &parametersJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return SituationReference{}, false, nil
	}
	if err != nil {
		return SituationReference{}, false, err
	}

	if len(parametersJSON) > 0 {
		err = json.Unmarshal(parametersJSON, &ref.Parameters)
		if err != nil {
			zap.L().Error("Error unmarshaling situation parameters", zap.Error(err))
			ref.Parameters = make(map[string]interface{})
		}
	}

	return ref, true, nil
}

// UpdateSituationReferenceParameters updates the parameters for a situation reference
func (r *PostgresRepository) UpdateSituationReferenceParameters(situationID int64, parameters map[string]interface{}) error {
	parametersJSON, err := json.Marshal(parameters)
	if err != nil {
		return fmt.Errorf("error marshaling parameters: %w", err)
	}

	_, err = r.newStatement().
		Update(tableSituationRef).
		Set("parameters", parametersJSON).
		Where(sq.Eq{"situation_id": situationID}).
		Exec()
	return err
}

// RemoveSituation removes an association between a functional situation and a situation
// If no other functional situation references this situation, the reference is also deleted
func (r *PostgresRepository) RemoveSituation(fsID int64, situationID int64) error {
	// First, remove the link
	_, err := r.newStatement().
		Delete(tableSituations).
		Where(sq.Eq{"functional_situation_id": fsID, "situation_id": situationID}).
		Exec()
	if err != nil {
		return err
	}

	// Check if any other functional situation still references this situation
	var count int
	err = r.conn.QueryRow("SELECT COUNT(*) FROM "+tableSituations+" WHERE situation_id = $1", situationID).Scan(&count)
	if err != nil {
		return err
	}

	// If no more references, delete the situation reference
	if count == 0 {
		_, err = r.newStatement().
			Delete(tableSituationRef).
			Where(sq.Eq{"situation_id": situationID}).
			Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// GetSituations retrieves all situation IDs associated with a functional situation
func (r *PostgresRepository) GetSituations(fsID int64) ([]int64, error) {
	rows, err := r.newStatement().
		Select("situation_id").
		From(tableSituations).
		Where(sq.Eq{"functional_situation_id": fsID}).
		OrderBy("situation_id").
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetSituationsWithParameters retrieves all situation IDs with their parameters for a functional situation in one query
func (r *PostgresRepository) GetSituationsWithParameters(fsID int64) (map[int64]map[string]interface{}, error) {
	rows, err := r.newStatement().
		Select("fss.situation_id", "COALESCE(ref.parameters, '{}'::jsonb)").
		From(tableSituations + " fss").
		InnerJoin(tableSituationRef + " ref ON ref.situation_id = fss.situation_id").
		Where(sq.Eq{"fss.functional_situation_id": fsID}).
		OrderBy("fss.situation_id").
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]map[string]interface{})
	for rows.Next() {
		var situationID int64
		var parametersJSON []byte

		if err := rows.Scan(&situationID, &parametersJSON); err != nil {
			return nil, err
		}

		var parameters map[string]interface{}
		if len(parametersJSON) > 0 {
			err = json.Unmarshal(parametersJSON, &parameters)
			if err != nil {
				zap.L().Error("Error unmarshaling situation parameters", zap.Error(err))
				parameters = make(map[string]interface{})
			}
		}

		result[situationID] = parameters
	}

	return result, rows.Err()
}

// GetEnrichedTree retrieves the complete hierarchy with all template instances and situations
func (r *PostgresRepository) GetEnrichedTree() ([]FunctionalSituationTreeNode, error) {
	// Step 1: Get all functional situations ordered by hierarchy
	query := `
		WITH RECURSIVE fs_tree AS (
			SELECT id, name, description, parent_id, color, icon, parameters, created_at, updated_at, created_by, 0 as depth
			FROM functional_situation_v1
			WHERE parent_id IS NULL
			UNION ALL
			SELECT fs.id, fs.name, fs.description, fs.parent_id, fs.color, fs.icon, fs.parameters, fs.created_at, fs.updated_at, fs.created_by, ft.depth + 1
			FROM functional_situation_v1 fs
			INNER JOIN fs_tree ft ON fs.parent_id = ft.id
		)
		SELECT id, name, description, parent_id, color, icon, created_at, updated_at, created_by, parameters
		FROM fs_tree
		ORDER BY depth, name
	`

	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fsList, err := dbutils.ScanAll(rows, r.scan)
	if err != nil {
		return nil, err
	}

	if len(fsList) == 0 {
		return []FunctionalSituationTreeNode{}, nil
	}

	// Collect all FS IDs
	fsIDs := make([]int64, len(fsList))
	for i, fs := range fsList {
		fsIDs[i] = fs.ID
	}

	// Step 2: Get all template instance associations with parameters in one query
	instanceAssocQuery := `
		SELECT fsi.functional_situation_id, ti.id, ti.name, ti.situation_id, s.name, COALESCE(ref.parameters, '{}'::jsonb)
		FROM functional_situation_instances_v1 fsi
		INNER JOIN functional_situation_instance_ref_v1 ref ON ref.template_instance_id = fsi.template_instance_id
		INNER JOIN situation_template_instances_v1 ti ON ti.id = fsi.template_instance_id
		INNER JOIN situation_definition_v1 s ON s.id = ti.situation_id
		WHERE fsi.functional_situation_id = ANY($1)
		ORDER BY fsi.functional_situation_id, ti.name
	`

	instanceRows, err := r.conn.Query(instanceAssocQuery, pq.Array(fsIDs))
	if err != nil {
		return nil, fmt.Errorf("error fetching template instances: %w", err)
	}
	defer instanceRows.Close()

	instancesByFS := make(map[int64][]TreeTemplateInstance)
	for instanceRows.Next() {
		var fsID int64
		var ti TreeTemplateInstance
		var parametersJSON []byte
		if err := instanceRows.Scan(&fsID, &ti.ID, &ti.Name, &ti.SituationID, &ti.SituationName, &parametersJSON); err != nil {
			return nil, err
		}
		if len(parametersJSON) > 0 {
			_ = json.Unmarshal(parametersJSON, &ti.Parameters)
		}
		instancesByFS[fsID] = append(instancesByFS[fsID], ti)
	}
	if err := instanceRows.Err(); err != nil {
		return nil, err
	}

	// Step 3: Get all situation associations with parameters in one query
	situationAssocQuery := `
		SELECT fss.functional_situation_id, s.id, s.name, s.is_template, COALESCE(ref.parameters, '{}'::jsonb)
		FROM functional_situation_situations_v1 fss
		INNER JOIN functional_situation_situation_ref_v1 ref ON ref.situation_id = fss.situation_id
		INNER JOIN situation_definition_v1 s ON s.id = fss.situation_id
		WHERE fss.functional_situation_id = ANY($1)
		ORDER BY fss.functional_situation_id, s.name
	`

	situationRows, err := r.conn.Query(situationAssocQuery, pq.Array(fsIDs))
	if err != nil {
		return nil, fmt.Errorf("error fetching situations: %w", err)
	}
	defer situationRows.Close()

	situationsByFS := make(map[int64][]TreeSituation)
	for situationRows.Next() {
		var fsID int64
		var s TreeSituation
		var parametersJSON []byte
		if err := situationRows.Scan(&fsID, &s.ID, &s.Name, &s.IsTemplate, &parametersJSON); err != nil {
			return nil, err
		}
		if len(parametersJSON) > 0 {
			_ = json.Unmarshal(parametersJSON, &s.Parameters)
		}
		situationsByFS[fsID] = append(situationsByFS[fsID], s)
	}
	if err := situationRows.Err(); err != nil {
		return nil, err
	}

	// Step 4: Build flat nodes map and parent-to-children map
	nodeDataMap := make(map[int64]FunctionalSituation)
	childrenMap := make(map[int64][]int64)
	var rootIDs []int64

	for _, fs := range fsList {
		nodeDataMap[fs.ID] = fs
		if fs.ParentID == nil {
			rootIDs = append(rootIDs, fs.ID)
		} else {
			childrenMap[*fs.ParentID] = append(childrenMap[*fs.ParentID], fs.ID)
		}
	}

	// Step 5: Recursive function to build tree node with children
	var buildNode func(id int64) FunctionalSituationTreeNode
	buildNode = func(id int64) FunctionalSituationTreeNode {
		fs := nodeDataMap[id]

		instances := instancesByFS[fs.ID]
		if instances == nil {
			instances = []TreeTemplateInstance{}
		}
		situations := situationsByFS[fs.ID]
		if situations == nil {
			situations = []TreeSituation{}
		}

		node := FunctionalSituationTreeNode{
			ID:                fs.ID,
			Name:              fs.Name,
			Description:       fs.Description,
			ParentID:          fs.ParentID,
			Color:             fs.Color,
			Icon:              fs.Icon,
			Parameters:        fs.Parameters,
			CreatedAt:         fs.CreatedAt,
			UpdatedAt:         fs.UpdatedAt,
			CreatedBy:         fs.CreatedBy,
			TemplateInstances: instances,
			Situations:        situations,
			Children:          []FunctionalSituationTreeNode{},
		}

		// Recursively build children
		for _, childID := range childrenMap[id] {
			node.Children = append(node.Children, buildNode(childID))
		}

		return node
	}

	// Step 6: Build the tree from roots
	result := make([]FunctionalSituationTreeNode, 0, len(rootIDs))
	for _, rootID := range rootIDs {
		result = append(result, buildNode(rootID))
	}

	if result == nil {
		return []FunctionalSituationTreeNode{}, nil
	}

	return result, nil
}
