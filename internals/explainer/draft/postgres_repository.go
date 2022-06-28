package draft

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	uuid "github.com/satori/go.uuid"
)

// PostgresRepository is a repository containing the FrontDraft definition based on a PSQL database and
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

// Get use to retrieve a draft by id
func (r *PostgresRepository) Get(issueID int64) (models.FrontDraft, bool, error) {

	query := `SELECT concurrency_uuid, data FROM issue_resolution_draft_v1 WHERE issue_id = :issue_id`
	params := map[string]interface{}{
		"issue_id": issueID,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return models.FrontDraft{}, false, err
	}
	defer rows.Close()
	var draft models.FrontDraft
	var concurrencyUUID, data string
	if rows.Next() {
		err := rows.Scan(&concurrencyUUID, &data)
		if err != nil {
			return models.FrontDraft{}, false, err
		}
	} else {
		return models.FrontDraft{}, false, nil
	}

	err = json.Unmarshal([]byte(data), &draft)
	if err != nil {
		return models.FrontDraft{}, false, err
	}

	draft.ConcurrencyUUID = concurrencyUUID

	return draft, true, nil
}

// Create method used to create a draft
func (r *PostgresRepository) Create(tx *sqlx.Tx, issueID int64, draft models.FrontDraft) error {

	data, err := json.Marshal(draft)
	if err != nil {
		return errors.New("couldn't marshall the provided data:" + err.Error())
	}

	now := time.Now().Truncate(1 * time.Millisecond).UTC()

	query := `INSERT into issue_resolution_draft_v1 (last_modified, concurrency_uuid, issue_id, data) 
		values (:last_modified, :concurrency_uuid, :issue_id, :data)`
	params := map[string]interface{}{
		"last_modified":    now,
		"concurrency_uuid": uuid.NewV4().String(),
		"issue_id":         issueID,
		"data":             data,
	}

	if tx != nil {
		_, err = tx.NamedExec(query, params)
	} else {
		_, err = r.conn.NamedExec(query, params)
	}
	if err != nil {
		return errors.New("couldn't query the database:" + err.Error())
	}
	return nil
}

// Update method used to update a draft
func (r *PostgresRepository) Update(tx *sqlx.Tx, issueID int64, draft models.FrontDraft) error {

	if draft.ConcurrencyUUID == "" {
		return errors.New("no concurrency uuid specified on this draft")
	}

	data, err := json.Marshal(draft)
	if err != nil {
		return errors.New("couldn't marshall the provided data:" + err.Error())
	}

	query := `UPDATE issue_resolution_draft_v1 SET last_modified = :last_modified, concurrency_uuid = :concurrency_uuid, data = :data 
		WHERE issue_id = :issue_id AND concurrency_uuid = :check_concurrency_uuid`
	params := map[string]interface{}{
		"last_modified":          time.Now().Truncate(1 * time.Millisecond).UTC(),
		"concurrency_uuid":       uuid.NewV4().String(),
		"issue_id":               issueID,
		"check_concurrency_uuid": draft.ConcurrencyUUID,
		"data":                   data,
	}

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

// CheckExists check if a draft already exists on an issue
func (r *PostgresRepository) CheckExists(tx *sqlx.Tx, issueID int64) (bool, error) {
	var exists bool
	var err error
	checkNameQuery := `select exists(select 1 from issue_resolution_draft_v1 where issue_id = $1) AS "exists"`
	if tx != nil {
		err = tx.QueryRow(checkNameQuery, issueID).Scan(&exists)
	} else {
		err = r.conn.QueryRow(checkNameQuery, issueID).Scan(&exists)
	}
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

// CheckExistsWithUUID double check if a draft already exists on an issue AND if the timestamp match the existing one
// In case it doesn't match, it means the draft has already been modified
func (r *PostgresRepository) CheckExistsWithUUID(tx *sqlx.Tx, issueID int64, uuid string) (bool, error) {
	var exists bool
	var err error
	checkNameQuery := `select exists(select 1 from issue_resolution_draft_v1 where issue_id = $1 AND concurrency_uuid = $2) AS "exists"`
	if tx != nil {
		err = tx.QueryRow(checkNameQuery, issueID, uuid).Scan(&exists)
	} else {
		err = r.conn.QueryRow(checkNameQuery, issueID, uuid).Scan(&exists)
	}
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}
