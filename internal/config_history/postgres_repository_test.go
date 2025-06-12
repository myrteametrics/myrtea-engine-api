package config_history

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
	"testing"
	"time"
)

// SQL scripts for testing
const (
	// ConfigHistoryDropTableV1 SQL statement for table drop
	ConfigHistoryDropTableV1 string = `DROP TABLE IF EXISTS ` + table + `;`
	// ConfigHistoryTableV1 SQL statement for the config_history table
	ConfigHistoryTableV1 string = `CREATE TABLE IF NOT EXISTS ` + table + `
	(
		id        bigint PRIMARY KEY NOT NULL,
		commentary text DEFAULT '',
		type      varchar(100) NOT NULL,
		user      varchar(150) NOT NULL
	);`
)

func dbInitRepo(dbClient *sqlx.DB, t *testing.T) {
	dbDestroyRepo(dbClient, t)

	_, err := dbClient.Exec(ConfigHistoryTableV1)
	if err != nil {
		t.Error(err)
	}
}

func dbDestroyRepo(dbClient *sqlx.DB, t *testing.T) {
	_, err := dbClient.Exec(ConfigHistoryDropTableV1)
	if err != nil {
		t.Error(err)
	}
}

func TestNewPostgresRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	if r == nil {
		t.Error("ConfigHistory Repository is nil")
	}
}

func TestPostgresReplaceGlobal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	r := NewPostgresRepository(tests.DBClient(t))
	reverse := ReplaceGlobals(r)
	if R() == nil {
		t.Error("Global ConfigHistory repository is nil")
	}
	reverse()
	if R() != nil {
		t.Error("Global ConfigHistory repository is not nil after reverse")
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

	var err error

	history := NewConfigHistory("test comment", "test_type", "test_user")
	id, err := r.Create(history)
	if err != nil {
		t.Error(err)
	}
	if id <= 0 {
		t.Error("invalid ConfigHistory ID")
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

	var err error

	historyGet, found, err := r.Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("found a config history from nowhere")
	}

	history := NewConfigHistory("test comment", "test_type", "test_user")
	id, err := r.Create(history)
	if err != nil {
		t.Error(err)
	}

	historyGet, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("ConfigHistory doesn't exist after creation")
		t.FailNow()
	}
	if id != historyGet.ID {
		t.Error("invalid ConfigHistory ID")
	}
	if history.Commentary != historyGet.Commentary {
		t.Error("invalid ConfigHistory Commentary")
	}
	if history.Type != historyGet.Type {
		t.Error("invalid ConfigHistory Type")
	}
	if history.User != historyGet.User {
		t.Error("invalid ConfigHistory User")
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

	var err error

	histories, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(histories) != 0 {
		t.Error("ConfigHistories should be empty")
	}

	history1 := NewConfigHistory("test comment 1", "test_type_1", "test_user_1")
	id1, err := r.Create(history1)
	if err != nil {
		t.Error(err)
	}

	history2 := NewConfigHistory("test comment 2", "test_type_2", "test_user_2")
	id2, err := r.Create(history2)
	if err != nil {
		t.Error(err)
	}

	histories, err = r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(histories) != 2 {
		t.Error("ConfigHistories should have 2 elements")
	}

	if _, ok := histories[id1]; !ok {
		t.Error("ConfigHistory 1 should be in the map")
	}
	if _, ok := histories[id2]; !ok {
		t.Error("ConfigHistory 2 should be in the map")
	}
}

func TestPostgresGetAllFromInterval(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error

	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	history1 := ConfigHistory{
		ID:         past.UnixNano() / int64(time.Millisecond),
		Commentary: "past comment",
		Type:       "test_type",
		User:       "test_user",
	}
	_, err = r.Create(history1)
	if err != nil {
		t.Error(err)
	}

	history2 := ConfigHistory{
		ID:         now.UnixNano() / int64(time.Millisecond),
		Commentary: "now comment",
		Type:       "test_type",
		User:       "test_user",
	}
	_, err = r.Create(history2)
	if err != nil {
		t.Error(err)
	}

	history3 := ConfigHistory{
		ID:         future.UnixNano() / int64(time.Millisecond),
		Commentary: "future comment",
		Type:       "test_type",
		User:       "test_user",
	}
	_, err = r.Create(history3)
	if err != nil {
		t.Error(err)
	}

	// Test getting all entries
	allHistories, err := r.GetAllFromInterval(past.Add(-1*time.Hour), future.Add(1*time.Hour))
	if err != nil {
		t.Error(err)
	}
	if len(allHistories) != 3 {
		t.Errorf("Expected 3 histories, got %d", len(allHistories))
	}

	// Test getting only past and now entries
	pastAndNowHistories, err := r.GetAllFromInterval(past.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Error(err)
	}
	if len(pastAndNowHistories) != 2 {
		t.Errorf("Expected 2 histories, got %d", len(pastAndNowHistories))
	}

	// Test getting only future entries
	futureHistories, err := r.GetAllFromInterval(now.Add(1*time.Hour), future.Add(1*time.Hour))
	if err != nil {
		t.Error(err)
	}
	if len(futureHistories) != 1 {
		t.Errorf("Expected 1 history, got %d", len(futureHistories))
	}
}

func TestPostgresGetAllByType(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error

	history1 := NewConfigHistory("test comment 1", "type_a", "test_user")
	_, err = r.Create(history1)
	if err != nil {
		t.Error(err)
	}

	history2 := NewConfigHistory("test comment 2", "type_a", "test_user")
	_, err = r.Create(history2)
	if err != nil {
		t.Error(err)
	}

	history3 := NewConfigHistory("test comment 3", "type_b", "test_user")
	_, err = r.Create(history3)
	if err != nil {
		t.Error(err)
	}

	typeAHistories, err := r.GetAllByType("type_a")
	if err != nil {
		t.Error(err)
	}
	if len(typeAHistories) != 2 {
		t.Errorf("Expected 2 histories of type_a, got %d", len(typeAHistories))
	}

	typeBHistories, err := r.GetAllByType("type_b")
	if err != nil {
		t.Error(err)
	}
	if len(typeBHistories) != 1 {
		t.Errorf("Expected 1 history of type_b, got %d", len(typeBHistories))
	}

	typeCHistories, err := r.GetAllByType("type_c")
	if err != nil {
		t.Error(err)
	}
	if len(typeCHistories) != 0 {
		t.Errorf("Expected 0 histories of type_c, got %d", len(typeCHistories))
	}
}

func TestPostgresGetAllByUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error

	history1 := NewConfigHistory("test comment 1", "test_type", "user_a")
	_, err = r.Create(history1)
	if err != nil {
		t.Error(err)
	}

	history2 := NewConfigHistory("test comment 2", "test_type", "user_a")
	_, err = r.Create(history2)
	if err != nil {
		t.Error(err)
	}

	history3 := NewConfigHistory("test comment 3", "test_type", "user_b")
	_, err = r.Create(history3)
	if err != nil {
		t.Error(err)
	}

	userAHistories, err := r.GetAllByUser("user_a")
	if err != nil {
		t.Error(err)
	}
	if len(userAHistories) != 2 {
		t.Errorf("Expected 2 histories from user_a, got %d", len(userAHistories))
	}

	userBHistories, err := r.GetAllByUser("user_b")
	if err != nil {
		t.Error(err)
	}
	if len(userBHistories) != 1 {
		t.Errorf("Expected 1 history from user_b, got %d", len(userBHistories))
	}

	userCHistories, err := r.GetAllByUser("user_c")
	if err != nil {
		t.Error(err)
	}
	if len(userCHistories) != 0 {
		t.Errorf("Expected 0 histories from user_c, got %d", len(userCHistories))
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

	var err error

	history := NewConfigHistory("test comment", "test_type", "test_user")
	id, err := r.Create(history)
	if err != nil {
		t.Error(err)
	}

	_, found, err := r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("ConfigHistory doesn't exist after creation")
		t.FailNow()
	}

	err = r.Delete(id)
	if err != nil {
		t.Error(err)
	}

	_, found, err = r.Get(id)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("ConfigHistory still exists after deletion")
	}
}

func TestPostgresDeleteNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	var err error

	err = r.Delete(999999)
	if err == nil {
		t.Error("Should not be able to delete a non-existing ConfigHistory")
	}
}

func TestPostgresHistoryRecordLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroyRepo(db, t)
	dbInitRepo(db, t)
	r := NewPostgresRepository(db)

	// Create maxHistoryRecords + 5 entries
	var ids []int64
	for i := 0; i < maxHistoryRecords+5; i++ {
		history := NewConfigHistory(fmt.Sprintf("test comment %d", i), "test_type", "test_user")
		// Add a small delay to ensure different timestamps
		time.Sleep(5 * time.Millisecond)
		id, err := r.Create(history)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		ids = append(ids, id)
	}

	// Get all entries
	histories, err := r.GetAll()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Check that we have exactly maxHistoryRecords entries
	if len(histories) != maxHistoryRecords {
		t.Errorf("Expected %d history records, got %d", maxHistoryRecords, len(histories))
	}

	// Check that the oldest entries were deleted
	// The first 5 entries should be gone, and only the last maxHistoryRecords should remain
	for i := 0; i < 5; i++ {
		_, found, err := r.Get(ids[i])
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		if found {
			t.Errorf("Expected entry with ID %d to be deleted, but it still exists", ids[i])
		}
	}

	// Check that the newest entries are still there
	for i := 5; i < maxHistoryRecords+5; i++ {
		_, found, err := r.Get(ids[i])
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		if !found {
			t.Errorf("Expected entry with ID %d to exist, but it was deleted", ids[i])
		}
	}
}
