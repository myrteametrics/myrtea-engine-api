package notification

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/dbutils"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
)

func dbInit(dbClient *sqlx.DB, t *testing.T) {
	dbDestroy(dbClient, t)
	_, err := dbClient.Exec(tests.NotificationHistoryTableV1)
	if err != nil {
		t.Error(err)
	}
}

func dbDestroy(dbClient *sqlx.DB, t *testing.T) {
	_, err := dbClient.Exec(tests.NotificationHistoryDropTableV1)
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
		t.Error("Notifications Repository is nil")
	}
}

func TestPostgresCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	groups := []int64{2, 3, 5}
	context := map[string]interface{}{"test": "test_context"}
	notif := NewMockNotification("level", "title", "subtitle", "description", time.Now().UTC(), groups, context)
	_, err = r.Create(groups, notif)
	if err != nil {
		t.Error(err)
	}
}

func TestPostgresGetByGroups(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)

	var err error
	groups := []int64{1, 3, 5}
	context := map[string]interface{}{"test": "test_context"}

	notif1 := NewMockNotification("level", "title", "subtitle", "description", time.Now().UTC(), groups, context)
	_, err = r.Create(groups, notif1)
	if err != nil {
		t.Error(err)
	}

	groups = []int64{1, 2}
	notif2 := NewMockNotification("level", "title", "subtitle", "description", time.Now().UTC(), groups, context)
	_, err = r.Create(groups, notif2)
	if err != nil {
		t.Error(err)
	}

	n, err := r.GetByGroups([]int64{1, 2, 3, 5, 8}, dbutils.DBQueryOptionnal{})
	if err != nil {
		t.Error(err)
	}
	if len(n) != 2 {
		t.Error("Fetching notifications by group Id not working correctly, want length of 2, have ", len(n))
	}

	n, err = r.GetByGroups([]int64{20}, dbutils.DBQueryOptionnal{})
	if err != nil {
		t.Error(err)
	}
	if len(n) != 0 {
		t.Error("Fetching notifications by group Id not working correctly, want length of 0, have ", len(n))
	}

	n, err = r.GetByGroups([]int64{3}, dbutils.DBQueryOptionnal{})
	if err != nil {
		t.Error(err)
	}
	if len(n) != 1 {
		t.Error("Fetching notifications by group Id not working correctly, want length of 3, have ", len(n))
	}
}

func TestPostgresGetByGroupsWithParams(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	r := NewPostgresRepository(db)

	var err error
	groups := []int64{1, 3, 5}
	context := map[string]interface{}{"test": "test_context"}
	notif1 := NewMockNotification("level", "title", "subtitle", "description", time.Now().UTC(), groups, context)
	id1, err := r.Create(groups, notif1)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)

	groups = []int64{1, 2}
	notif2 := NewMockNotification("level", "title", "subtitle", "description", time.Now().UTC(), groups, context)
	id2, err := r.Create(groups, notif2)
	if err != nil {
		t.Error(err)
	}

	n, err := r.GetByGroups([]int64{1, 2, 3, 5, 8}, dbutils.DBQueryOptionnal{Limit: 1})
	if err != nil {
		t.Error(err)
	}
	if len(n) != 1 {
		t.Error("Fetching notifications by group Id not working correctly, want length of 1, have ", len(n))
	}
	nn := n[0].Notification.(MockNotification)
	if nn.ID != id2 { // Most recent one (last created)
		t.Error("invalid notification ID")
	}

	n, err = r.GetByGroups([]int64{1, 2, 3, 5, 8}, dbutils.DBQueryOptionnal{Limit: 1, Offset: 1})
	if err != nil {
		t.Error(err)
	}
	if len(n) != 1 {
		t.Error("Fetching notifications by group Id not working correctly, want length of 1, have ", len(n))
	}
	nn = n[0].Notification.(MockNotification)
	if nn.ID != id1 {
		t.Error("invalid notification ID")
	}
}

func TestPostgresDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	groups := []int64{1, 3, 5}
	context := map[string]interface{}{"test": "test_context"}
	notif := NewMockNotification("level", "title", "subtitle", "description", time.Now().UTC(), groups, context)
	_, err = r.Create(groups, notif)
	if err != nil {
		t.Error(err)
	}

	err = r.Delete(1)
	if err != nil {
		t.Error(err)
	}

	n, _ := r.GetByGroups([]int64{1}, dbutils.DBQueryOptionnal{})
	if len(n) != 0 {
		t.Error("Fetching notifications by group Id not working correctly, want length of 0, have ", len(n))
	}
}

func TestUpdateRead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)
	r := NewPostgresRepository(db)

	var err error
	groups := []int64{1, 3, 5}
	context := map[string]interface{}{"test": "test_context"}
	notif := NewMockNotification("level", "title", "subtitle", "description", time.Now().UTC(), groups, context)
	_, err = r.Create(groups, notif)
	if err != nil {
		t.Error(err)
	}
	err = r.UpdateRead(1, true)
	if err != nil {
		t.Error("Couldn't perform the status update")
	}
	groups = []int64{1, 2}
	n, err := r.GetByGroups(groups, dbutils.DBQueryOptionnal{Limit: 1})
	if err != nil {
		t.Error("Fetching notifications by group Id not working correctly, want length of 0, have ", len(n))
	}
	if !n[0].IsRead {
		t.Error("the read propertie isn't update as expected, still equal to: ", n[0].IsRead)
	}

	err = r.UpdateRead(1, true)
	if err != nil {
		t.Error("property 'isread' already was in that state.")
	}

	err = r.UpdateRead(22, true)
	if err == nil {
		t.Error("Updated a non-existent notification")
	}

}
