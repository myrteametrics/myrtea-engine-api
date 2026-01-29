package functionalsituation

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
)

const (
	situationTemplateInstancesTableV1 = `
		CREATE TABLE IF NOT EXISTS situation_template_instances_v1 (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		);
	`
	situationTemplateInstancesDropTableV1 = `DROP TABLE IF EXISTS situation_template_instances_v1 CASCADE;`

	situationDefinitionTableV1 = `
		CREATE TABLE IF NOT EXISTS situation_definition_v1 (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL
		);
	`
	situationDefinitionDropTableV1 = `DROP TABLE IF EXISTS situation_definition_v1 CASCADE;`

	functionalSituationTableV1 = `
		CREATE TABLE IF NOT EXISTS functional_situation_v1 (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			description TEXT DEFAULT ''::text,
			parent_id INTEGER REFERENCES functional_situation_v1 (id) ON DELETE SET NULL,
			color VARCHAR(7) DEFAULT '#0066CC',
			icon VARCHAR(50) DEFAULT 'folder',
			created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
			created_by VARCHAR(100) NOT NULL,
			metadata JSONB DEFAULT '{}'::jsonb
		);
		CREATE INDEX IF NOT EXISTS idx_functional_situation_parent ON functional_situation_v1 (parent_id);
		CREATE UNIQUE INDEX IF NOT EXISTS unq_functional_situation_name_parent ON functional_situation_v1 ((COALESCE(parent_id, 0)), name);
	`
	functionalSituationDropTableV1 = `DROP TABLE IF EXISTS functional_situation_v1 CASCADE;`

	functionalSituationInstancesTableV1 = `
		CREATE TABLE IF NOT EXISTS functional_situation_instances_v1 (
			functional_situation_id INTEGER NOT NULL REFERENCES functional_situation_v1 (id) ON DELETE CASCADE,
			template_instance_id INTEGER NOT NULL REFERENCES situation_template_instances_v1 (id) ON DELETE CASCADE,
			added_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
			added_by VARCHAR(100) NOT NULL,
			PRIMARY KEY (functional_situation_id, template_instance_id)
		);
	`
	functionalSituationInstancesDropTableV1 = `DROP TABLE IF EXISTS functional_situation_instances_v1 CASCADE;`

	functionalSituationSituationsTableV1 = `
		CREATE TABLE IF NOT EXISTS functional_situation_situations_v1 (
			functional_situation_id INTEGER NOT NULL REFERENCES functional_situation_v1 (id) ON DELETE CASCADE,
			situation_id INTEGER NOT NULL REFERENCES situation_definition_v1 (id) ON DELETE CASCADE,
			added_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
			added_by VARCHAR(100) NOT NULL,
			PRIMARY KEY (functional_situation_id, situation_id)
		);
	`
	functionalSituationSituationsDropTableV1 = `DROP TABLE IF EXISTS functional_situation_situations_v1 CASCADE;`
)

func dbInitRepo(dbClient *sqlx.DB, t *testing.T) {
	dbDestroyRepo(dbClient, t)

	tests.DBExec(dbClient, situationTemplateInstancesTableV1, t, true)
	tests.DBExec(dbClient, situationDefinitionTableV1, t, true)
	tests.DBExec(dbClient, functionalSituationTableV1, t, true)
	tests.DBExec(dbClient, functionalSituationInstancesTableV1, t, true)
	tests.DBExec(dbClient, functionalSituationSituationsTableV1, t, true)
}

func dbDestroyRepo(dbClient *sqlx.DB, t *testing.T) {
	tests.DBExec(dbClient, functionalSituationSituationsDropTableV1, t, true)
	tests.DBExec(dbClient, functionalSituationInstancesDropTableV1, t, true)
	tests.DBExec(dbClient, functionalSituationDropTableV1, t, true)
	tests.DBExec(dbClient, situationDefinitionDropTableV1, t, true)
	tests.DBExec(dbClient, situationTemplateInstancesDropTableV1, t, true)
}

func TestNewPostgresRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	if r == nil {
		t.Error("functional situation Repository is nil")
	}
}

func TestPostgresReplaceGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global functional situation repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global functional situation repository is not nil after reverse")
	}
}

func TestPostgresCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	fs := FunctionalSituation{
		Name:        "Test FS",
		Description: "Test description",
		Color:       "#FF0000",
		Icon:        "folder",
		Parameters:  map[string]interface{}{"key": "value"},
	}

	id, err := r.Create(fs, "testuser")
	if err != nil {
		t.Fatal(err)
	}
	if id <= 0 {
		t.Error("Expected positive ID")
	}
}

func TestPostgresGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	// Test not found
	_, found, err := r.Get(999)
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Error("Should not find non-existent FS")
	}

	// Create and get
	fs := FunctionalSituation{
		Name:        "Test FS",
		Description: "Test description",
		Color:       "#FF0000",
		Icon:        "folder",
	}

	id, err := r.Create(fs, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	retrieved, found, err := r.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Error("Should find created FS")
	}
	if retrieved.Name != fs.Name {
		t.Errorf("Expected name %s, got %s", fs.Name, retrieved.Name)
	}
	if retrieved.Color != fs.Color {
		t.Errorf("Expected color %s, got %s", fs.Color, retrieved.Color)
	}
}

func TestPostgresUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	// Create FS
	fs := FunctionalSituation{
		Name:  "Original Name",
		Color: "#FF0000",
	}
	id, err := r.Create(fs, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	// Update
	update := FunctionalSituation{
		Name:  "Updated Name",
		Color: "#00FF00",
	}
	err = r.Update(id, update, "updater")
	if err != nil {
		t.Fatal(err)
	}

	// Verify update
	retrieved, _, err := r.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name %s, got %s", "Updated Name", retrieved.Name)
	}
	if retrieved.Color != "#00FF00" {
		t.Errorf("Expected color %s, got %s", "#00FF00", retrieved.Color)
	}
}

func TestPostgresDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	// Create FS
	fs := FunctionalSituation{
		Name:  "To Delete",
		Color: "#FF0000",
	}
	id, err := r.Create(fs, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	// Delete
	err = r.Delete(id)
	if err != nil {
		t.Fatal(err)
	}

	// Verify deletion
	_, found, err := r.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Error("Should not find deleted FS")
	}
}

func TestPostgresGetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	// Create multiple FS
	for i := 0; i < 3; i++ {
		fs := FunctionalSituation{
			Name:  "FS " + string(rune('A'+i)),
			Color: "#FF0000",
		}
		_, err := r.Create(fs, "testuser")
		if err != nil {
			t.Fatal(err)
		}
	}

	all, err := r.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Errorf("Expected 3 FSs, got %d", len(all))
	}
}

func TestPostgresGetChildren(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	// Create parent
	parent := FunctionalSituation{
		Name:  "Parent",
		Color: "#FF0000",
	}
	parentID, err := r.Create(parent, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	// Create children
	for i := 0; i < 2; i++ {
		child := FunctionalSituation{
			Name:     "Child " + string(rune('A'+i)),
			Color:    "#00FF00",
			ParentID: &parentID,
		}
		_, err := r.Create(child, "testuser")
		if err != nil {
			t.Fatal(err)
		}
	}

	children, err := r.GetChildren(parentID)
	if err != nil {
		t.Fatal(err)
	}
	if len(children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(children))
	}
}

func TestPostgresGetRoots(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	// Create root FS
	root1 := FunctionalSituation{
		Name:  "Root 1",
		Color: "#FF0000",
	}
	id1, err := r.Create(root1, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	root2 := FunctionalSituation{
		Name:  "Root 2",
		Color: "#00FF00",
	}
	_, err = r.Create(root2, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	// Create child (should not be in roots)
	child := FunctionalSituation{
		Name:     "Child",
		Color:    "#0000FF",
		ParentID: &id1,
	}
	_, err = r.Create(child, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	roots, err := r.GetRoots()
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 2 {
		t.Errorf("Expected 2 roots, got %d", len(roots))
	}
}

func TestPostgresGetTree(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	// Create hierarchy
	root := FunctionalSituation{
		Name:  "Root",
		Color: "#FF0000",
	}
	rootID, err := r.Create(root, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	child := FunctionalSituation{
		Name:     "Child",
		Color:    "#00FF00",
		ParentID: &rootID,
	}
	childID, err := r.Create(child, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	grandchild := FunctionalSituation{
		Name:     "Grandchild",
		Color:    "#0000FF",
		ParentID: &childID,
	}
	_, err = r.Create(grandchild, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	tree, err := r.GetTree()
	if err != nil {
		t.Fatal(err)
	}
	if len(tree) != 3 {
		t.Errorf("Expected 3 nodes in tree, got %d", len(tree))
	}
	// Tree should be ordered by depth, so root should be first
	if tree[0].Name != "Root" {
		t.Errorf("Expected first element to be Root, got %s", tree[0].Name)
	}
}

func TestPostgresAddRemoveTemplateInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	// Create FS
	fs := FunctionalSituation{
		Name:  "Test FS",
		Color: "#FF0000",
	}
	fsID, err := r.Create(fs, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	// Create template instance
	_, err = db.Exec("INSERT INTO situation_template_instances_v1 (id, name) VALUES (100, 'Template 1')")
	if err != nil {
		t.Fatal(err)
	}

	// Add association
	err = r.AddTemplateInstance(fsID, 100, map[string]interface{}{}, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	// Get instances
	instances, err := r.GetTemplateInstances(fsID)
	if err != nil {
		t.Fatal(err)
	}
	if len(instances) != 1 || instances[0] != 100 {
		t.Error("Template instance not associated correctly")
	}

	// Remove association
	err = r.RemoveTemplateInstance(fsID, 100)
	if err != nil {
		t.Fatal(err)
	}

	instances, err = r.GetTemplateInstances(fsID)
	if err != nil {
		t.Fatal(err)
	}
	if len(instances) != 0 {
		t.Error("Template instance not removed correctly")
	}
}

func TestPostgresAddRemoveSituation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	// Create FS
	fs := FunctionalSituation{
		Name:  "Test FS",
		Color: "#FF0000",
	}
	fsID, err := r.Create(fs, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	// Create situation
	_, err = db.Exec("INSERT INTO situation_definition_v1 (id, name) VALUES (200, 'Situation 1')")
	if err != nil {
		t.Fatal(err)
	}

	// Add association
	err = r.AddSituation(fsID, 200, make(map[string]interface{}), "testuser")
	if err != nil {
		t.Fatal(err)
	}

	// Get situations
	situations, err := r.GetSituations(fsID)
	if err != nil {
		t.Fatal(err)
	}
	if len(situations) != 1 || situations[0] != 200 {
		t.Error("Situation not associated correctly")
	}

	// Remove association
	err = r.RemoveSituation(fsID, 200)
	if err != nil {
		t.Fatal(err)
	}

	situations, err = r.GetSituations(fsID)
	if err != nil {
		t.Fatal(err)
	}
	if len(situations) != 0 {
		t.Error("Situation not removed correctly")
	}
}

func TestPostgresUniqueConstraint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	// Create first FS
	fs1 := FunctionalSituation{
		Name:  "Duplicate Name",
		Color: "#FF0000",
	}
	_, err := r.Create(fs1, "testuser")
	if err != nil {
		t.Fatal(err)
	}

	// Try to create duplicate (should fail)
	fs2 := FunctionalSituation{
		Name:  "Duplicate Name",
		Color: "#00FF00",
	}
	_, err = r.Create(fs2, "testuser")
	if err == nil {
		t.Error("Expected error for duplicate name at same level, got nil")
	}
}
