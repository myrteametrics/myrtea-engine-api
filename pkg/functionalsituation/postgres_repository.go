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

var fields = []string{"id", "name", "description", "parent_id", "color", "icon", "created_at", "updated_at", "created_by", "metadata"}

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
	var metadataJSON []byte
	err := rows.Scan(&fs.ID, &fs.Name, &fs.Description, &fs.ParentID, &fs.Color, &fs.Icon,
		&fs.CreatedAt, &fs.UpdatedAt, &fs.CreatedBy, &metadataJSON)
	if err != nil {
		return FunctionalSituation{}, err
	}

	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &fs.Metadata)
		if err != nil {
			zap.L().Error("Error unmarshaling metadata", zap.Error(err))
			fs.Metadata = make(map[string]interface{})
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

	metadataJSON, err := json.Marshal(fs.Metadata)
	if err != nil {
		return -1, fmt.Errorf("error marshaling metadata: %w", err)
	}

	statement := r.newStatement().
		Insert(table).
		Columns("name", "description", "parent_id", "color", "icon", "created_at", "updated_at", "created_by", "metadata").
		Values(fs.Name, fs.Description, fs.ParentID, fs.Color, fs.Icon, now, now, createdBy, metadataJSON).
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
	if fs.Metadata != nil {
		metadataJSON, err := json.Marshal(*fs.Metadata)
		if err != nil {
			return fmt.Errorf("error marshaling metadata: %w", err)
		}
		statement = statement.Set("metadata", metadataJSON)
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
			SELECT id, name, description, parent_id, color, icon, metadata, created_at, updated_at, created_by, 0 as depth
			FROM functional_situation_v1
			WHERE parent_id IS NULL
			UNION ALL
			SELECT fs.id, fs.name, fs.description, fs.parent_id, fs.color, fs.icon, fs.metadata, fs.created_at, fs.updated_at, fs.created_by, ft.depth + 1
			FROM functional_situation_v1 fs
			INNER JOIN fs_tree ft ON fs.parent_id = ft.id
		)
		SELECT id, name, description, parent_id, color, icon, created_at, updated_at, created_by, metadata
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
			SELECT id, name, description, parent_id, color, icon, metadata, created_at, updated_at, created_by, 0 as depth
			FROM functional_situation_v1
			WHERE id = $1
			UNION ALL
			SELECT fs.id, fs.name, fs.description, fs.parent_id, fs.color, fs.icon, fs.metadata, fs.created_at, fs.updated_at, fs.created_by, a.depth + 1
			FROM functional_situation_v1 fs
			INNER JOIN ancestors a ON fs.id = a.parent_id
		)
		SELECT id, name, description, parent_id, color, icon, created_at, updated_at, created_by, metadata
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

// MoveToParent changes the parent of a functional situation
func (r *PostgresRepository) MoveToParent(id int64, newParentID *int64) error {
	// Check for circular reference if newParentID is not nil
	if newParentID != nil {
		ancestors, err := r.GetAncestors(*newParentID)
		if err != nil {
			return err
		}
		for _, ancestor := range ancestors {
			if ancestor.ID == id {
				return fmt.Errorf("circular reference detected: cannot move to descendant")
			}
		}
	}

	_, err := r.newStatement().
		Update(table).
		Set("parent_id", newParentID).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": id}).
		Exec()
	return err
}

// AddTemplateInstance creates an association between a functional situation and a template instance
func (r *PostgresRepository) AddTemplateInstance(fsID int64, instanceID int64, addedBy string) error {
	_, err := r.newStatement().
		Insert(tableInstances).
		Columns("functional_situation_id", "template_instance_id", "added_at", "added_by").
		Values(fsID, instanceID, time.Now(), addedBy).
		Exec()
	return err
}

// RemoveTemplateInstance removes an association between a functional situation and a template instance
func (r *PostgresRepository) RemoveTemplateInstance(fsID int64, instanceID int64) error {
	_, err := r.newStatement().
		Delete(tableInstances).
		Where(sq.Eq{"functional_situation_id": fsID, "template_instance_id": instanceID}).
		Exec()
	return err
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

// GetFunctionalSituationsByInstance retrieves all functional situations associated with a template instance
func (r *PostgresRepository) GetFunctionalSituationsByInstance(instanceID int64) ([]FunctionalSituation, error) {
	fieldsWithPrefix := make([]string, len(fields))
	for i, field := range fields {
		fieldsWithPrefix[i] = "fs." + field
	}

	rows, err := r.newStatement().
		Select(fieldsWithPrefix...).
		From(table + " fs").
		InnerJoin(tableInstances + " fsi ON fs.id = fsi.functional_situation_id").
		Where(sq.Eq{"fsi.template_instance_id": instanceID}).
		OrderBy("fs.name").
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dbutils.ScanAll(rows, r.scan)
}

// AddSituation creates an association between a functional situation and a situation
func (r *PostgresRepository) AddSituation(fsID int64, situationID int64, addedBy string) error {
	_, err := r.newStatement().
		Insert(tableSituations).
		Columns("functional_situation_id", "situation_id", "added_at", "added_by").
		Values(fsID, situationID, time.Now(), addedBy).
		Exec()
	return err
}

// RemoveSituation removes an association between a functional situation and a situation
func (r *PostgresRepository) RemoveSituation(fsID int64, situationID int64) error {
	_, err := r.newStatement().
		Delete(tableSituations).
		Where(sq.Eq{"functional_situation_id": fsID, "situation_id": situationID}).
		Exec()
	return err
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

// GetFunctionalSituationsBySituation retrieves all functional situations associated with a situation
func (r *PostgresRepository) GetFunctionalSituationsBySituation(situationID int64) ([]FunctionalSituation, error) {
	fieldsWithPrefix := make([]string, len(fields))
	for i, field := range fields {
		fieldsWithPrefix[i] = "fs." + field
	}

	rows, err := r.newStatement().
		Select(fieldsWithPrefix...).
		From(table + " fs").
		InnerJoin(tableSituations + " fss ON fs.id = fss.functional_situation_id").
		Where(sq.Eq{"fss.situation_id": situationID}).
		OrderBy("fs.name").
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dbutils.ScanAll(rows, r.scan)
}

// GetOverview retrieves an overview of all functional situations with aggregated counts
func (r *PostgresRepository) GetOverview() ([]FunctionalSituationOverview, error) {
	query := `
		WITH RECURSIVE fs_tree AS (
			SELECT 
				fs.id, 
				fs.name, 
				fs.description, 
				fs.parent_id, 
				fs.color, 
				fs.icon,
				0 as depth
			FROM functional_situation_v1 fs
			WHERE fs.parent_id IS NULL
			UNION ALL
			SELECT 
				fs.id, 
				fs.name, 
				fs.description, 
				fs.parent_id, 
				fs.color, 
				fs.icon,
				ft.depth + 1
			FROM functional_situation_v1 fs
			INNER JOIN fs_tree ft ON fs.parent_id = ft.id
		),
		counts AS (
			SELECT 
				ft.id,
				ft.name,
				ft.description,
				ft.parent_id,
				ft.color,
				ft.icon,
				ft.depth,
				COUNT(DISTINCT fsi.template_instance_id) as instance_count,
				COUNT(DISTINCT fss.situation_id) as situation_count,
				COUNT(DISTINCT children.id) as children_count
			FROM fs_tree ft
			LEFT JOIN functional_situation_instances_v1 fsi ON ft.id = fsi.functional_situation_id
			LEFT JOIN functional_situation_situations_v1 fss ON ft.id = fss.functional_situation_id
			LEFT JOIN functional_situation_v1 children ON children.parent_id = ft.id
			GROUP BY ft.id, ft.name, ft.description, ft.parent_id, ft.color, ft.icon, ft.depth
		)
		SELECT 
			id, 
			name, 
			description, 
			parent_id, 
			color, 
			icon,
			instance_count::int,
			situation_count::int,
			children_count::int
		FROM counts
		ORDER BY depth, name
	`

	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overviews []FunctionalSituationOverview
	for rows.Next() {
		var o FunctionalSituationOverview
		err := rows.Scan(&o.ID, &o.Name, &o.Description, &o.ParentID, &o.Color, &o.Icon,
			&o.InstanceCount, &o.SituationCount, &o.ChildrenCount)
		if err != nil {
			return nil, err
		}
		// Set aggregated status to "unknown" for now (will be implemented with issue integration)
		o.AggregatedStatus = "unknown"
		overviews = append(overviews, o)
	}

	return overviews, rows.Err()
}

// GetOverviewByID retrieves an overview for a specific functional situation
func (r *PostgresRepository) GetOverviewByID(id int64) (FunctionalSituationOverview, bool, error) {
	query := `
		SELECT 
			fs.id, 
			fs.name, 
			fs.description, 
			fs.parent_id, 
			fs.color, 
			fs.icon,
			COUNT(DISTINCT fsi.template_instance_id)::int as instance_count,
			COUNT(DISTINCT fss.situation_id)::int as situation_count,
			COUNT(DISTINCT children.id)::int as children_count
		FROM functional_situation_v1 fs
		LEFT JOIN functional_situation_instances_v1 fsi ON fs.id = fsi.functional_situation_id
		LEFT JOIN functional_situation_situations_v1 fss ON fs.id = fss.functional_situation_id
		LEFT JOIN functional_situation_v1 children ON children.parent_id = fs.id
		WHERE fs.id = $1
		GROUP BY fs.id, fs.name, fs.description, fs.parent_id, fs.color, fs.icon
	`

	var o FunctionalSituationOverview
	err := r.conn.QueryRow(query, id).Scan(&o.ID, &o.Name, &o.Description, &o.ParentID, &o.Color, &o.Icon,
		&o.InstanceCount, &o.SituationCount, &o.ChildrenCount)
	if errors.Is(err, sql.ErrNoRows) {
		return FunctionalSituationOverview{}, false, nil
	}
	if err != nil {
		return FunctionalSituationOverview{}, false, err
	}

	// Set aggregated status to "unknown" for now
	o.AggregatedStatus = "unknown"

	return o, true, nil
}

// GetEnrichedTree retrieves the complete hierarchy with all template instances and situations
func (r *PostgresRepository) GetEnrichedTree() ([]FunctionalSituationTreeNode, error) {
	// Step 1: Get all functional situations ordered by hierarchy
	query := `
		WITH RECURSIVE fs_tree AS (
			SELECT id, name, description, parent_id, color, icon, metadata, created_at, updated_at, created_by, 0 as depth
			FROM functional_situation_v1
			WHERE parent_id IS NULL
			UNION ALL
			SELECT fs.id, fs.name, fs.description, fs.parent_id, fs.color, fs.icon, fs.metadata, fs.created_at, fs.updated_at, fs.created_by, ft.depth + 1
			FROM functional_situation_v1 fs
			INNER JOIN fs_tree ft ON fs.parent_id = ft.id
		)
		SELECT id, name, description, parent_id, color, icon, created_at, updated_at, created_by, metadata
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

	// Step 2: Get all template instance associations in one query
	instanceAssocQuery := `
		SELECT fsi.functional_situation_id, ti.id, ti.name, ti.situation_id
		FROM functional_situation_instances_v1 fsi
		INNER JOIN situation_template_instances_v1 ti ON ti.id = fsi.template_instance_id
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
		if err := instanceRows.Scan(&fsID, &ti.ID, &ti.Name, &ti.SituationID); err != nil {
			return nil, err
		}
		instancesByFS[fsID] = append(instancesByFS[fsID], ti)
	}
	if err := instanceRows.Err(); err != nil {
		return nil, err
	}

	// Step 3: Get all situation associations in one query
	situationAssocQuery := `
		SELECT fss.functional_situation_id, s.id, s.name, s.is_template
		FROM functional_situation_situations_v1 fss
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
		if err := situationRows.Scan(&fsID, &s.ID, &s.Name, &s.IsTemplate); err != nil {
			return nil, err
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
			Metadata:          fs.Metadata,
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
