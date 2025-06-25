package config_history

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

const (
	table = "config_history_v1"
)

// PostgresRepository implements the Repository interface for PostgreSQL
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

// Create method used to create a ConfigHistory entry
func (r *PostgresRepository) Create(history ConfigHistory) (int64, error) {
	// Validate the history entry
	if valid, err := history.IsValid(); !valid {
		return -1, err
	}

	// Begin a transaction to ensure atomicity
	tx, err := r.conn.Beginx()
	if err != nil {
		return -1, fmt.Errorf("couldn't begin transaction: %s", err.Error())
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Count the current number of records
	var count int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	err = tx.QueryRow(countQuery).Scan(&count)
	if err != nil {
		return -1, fmt.Errorf("couldn't count config history records: %s", err.Error())
	}

	// If we already have maxHistoryRecords or more, delete the oldest records
	if count >= maxHistoryRecords {
		// Calculate how many records to delete
		toDelete := count - maxHistoryRecords + 1 // +1 to make room for the new record

		// Delete the oldest records (those with the lowest IDs)
		deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE id IN (SELECT id FROM %s ORDER BY id ASC LIMIT $1)", table, table)
		_, err = tx.Exec(deleteQuery, toDelete)
		if err != nil {
			return -1, fmt.Errorf("couldn't delete oldest config history records: %s", err.Error())
		}
	}

	// Insert the new record
	statement := sq.StatementBuilder.
		PlaceholderFormat(sq.Dollar).
		RunWith(tx).
		Insert(table).
		Columns("id", "commentary", "update_type", "update_user", "config").
		Values(history.ID, history.Commentary, history.Type, history.User, history.Config)

	_, err = statement.Exec()
	if err != nil {
		return -1, fmt.Errorf("couldn't insert config history: %s", err.Error())
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return -1, fmt.Errorf("couldn't commit transaction: %s", err.Error())
	}

	return history.ID, nil
}

// Get retrieves a ConfigHistory entry by id
func (r *PostgresRepository) Get(id int64) (ConfigHistory, bool, error) {
	query := `SELECT id, commentary, update_type, update_user, config FROM config_history_v1 WHERE id = $1`

	var history ConfigHistory
	err := r.conn.QueryRow(query, id).Scan(&history.ID, &history.Commentary, &history.Type, &history.User, &history.Config)

	if err != nil {
		if err == sql.ErrNoRows {
			return ConfigHistory{}, false, nil
		}
		return ConfigHistory{}, false, fmt.Errorf("couldn't retrieve config history with id %d: %s", id, err.Error())
	}

	return history, true, nil
}

// GetAll returns all ConfigHistory entries
func (r *PostgresRepository) GetAll() (map[int64]ConfigHistory, error) {
	histories := make(map[int64]ConfigHistory)

	query := `SELECT id, commentary, update_type, update_user, config FROM config_history_v1 ORDER BY id DESC`
	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve config histories: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var history ConfigHistory
		err := rows.Scan(&history.ID, &history.Commentary, &history.Type, &history.User, &history.Config)
		if err != nil {
			return nil, fmt.Errorf("couldn't scan config history: %s", err.Error())
		}
		histories[history.ID] = history
	}

	return histories, nil
}

// GetAllFromInterval returns all ConfigHistory entries within a time interval
func (r *PostgresRepository) GetAllFromInterval(from time.Time, to time.Time) (map[int64]ConfigHistory, error) {
	histories := make(map[int64]ConfigHistory)

	fromMillis := from.UnixNano() / int64(time.Millisecond)
	toMillis := to.UnixNano() / int64(time.Millisecond)

	query := `SELECT id, commentary, update_type, update_user, config FROM config_history_v1 WHERE id >= $1 AND id <= $2 ORDER BY id DESC`
	rows, err := r.conn.Query(query, fromMillis, toMillis)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve config histories in interval: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var history ConfigHistory
		err := rows.Scan(&history.ID, &history.Commentary, &history.Type, &history.User, &history.Config)
		if err != nil {
			return nil, fmt.Errorf("couldn't scan config history: %s", err.Error())
		}
		histories[history.ID] = history
	}

	return histories, nil
}

// GetAllByType returns all ConfigHistory entries of a specific type
func (r *PostgresRepository) GetAllByType(historyType string) (map[int64]ConfigHistory, error) {
	histories := make(map[int64]ConfigHistory)

	query := `SELECT id, commentary, update_type, update_user, config FROM config_history_v1 WHERE update_type = $1 ORDER BY id DESC`
	rows, err := r.conn.Query(query, historyType)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve config histories by type: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var history ConfigHistory
		err := rows.Scan(&history.ID, &history.Commentary, &history.Type, &history.User, &history.Config)
		if err != nil {
			return nil, fmt.Errorf("couldn't scan config history: %s", err.Error())
		}
		histories[history.ID] = history
	}

	return histories, nil
}

// GetAllByUser returns all ConfigHistory entries created by a specific user
func (r *PostgresRepository) GetAllByUser(user string) (map[int64]ConfigHistory, error) {
	histories := make(map[int64]ConfigHistory)

	query := `SELECT id, commentary, update_type, update_user, config FROM config_history_v1 WHERE update_user = $1 ORDER BY id DESC`
	rows, err := r.conn.Query(query, user)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve config histories by user: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var history ConfigHistory
		err := rows.Scan(&history.ID, &history.Commentary, &history.Type, &history.User, &history.Config)
		if err != nil {
			return nil, fmt.Errorf("couldn't scan config history: %s", err.Error())
		}
		histories[history.ID] = history
	}

	return histories, nil
}

// Delete removes a ConfigHistory entry by id
func (r *PostgresRepository) Delete(id int64) error {
	query := `DELETE FROM config_history_v1 WHERE id = $1`

	res, err := r.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("couldn't delete config history: %s", err.Error())
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %s", err.Error())
	}

	if count == 0 {
		return errors.New("no config history found with the specified id")
	}

	return nil
}

// DeleteOldest removes the oldest ConfigHistory entry (the one with the lowest ID)
func (r *PostgresRepository) DeleteOldest() error {
	query := `DELETE FROM config_history_v1 WHERE id = (SELECT id FROM config_history_v1 ORDER BY id ASC LIMIT 1)`

	res, err := r.conn.Exec(query)
	if err != nil {
		return fmt.Errorf("couldn't delete oldest config history: %s", err.Error())
	}

	count, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %s", err.Error())
	}

	if count == 0 {
		return errors.New("no config history found to delete")
	}

	return nil
}
