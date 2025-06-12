package config_history

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

const table = "config_history_v1"

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

	statement := r.newStatement().
		Insert(table).
		Columns("id", "commentary", "type", "user").
		Values(history.ID, history.Commentary, history.Type, history.User)

	_, err := statement.Exec()
	if err != nil {
		return -1, fmt.Errorf("couldn't insert config history: %s", err.Error())
	}

	return history.ID, nil
}

// Get retrieves a ConfigHistory entry by id
func (r *PostgresRepository) Get(id int64) (ConfigHistory, bool, error) {
	query := `SELECT id, commentary, type, user FROM config_history_v1 WHERE id = $1`

	var history ConfigHistory
	err := r.conn.QueryRow(query, id).Scan(&history.ID, &history.Commentary, &history.Type, &history.User)

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

	query := `SELECT id, commentary, type, user FROM config_history_v1 ORDER BY id DESC`
	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve config histories: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var history ConfigHistory
		err := rows.Scan(&history.ID, &history.Commentary, &history.Type, &history.User)
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

	query := `SELECT id, commentary, type, user FROM config_history_v1 WHERE id >= $1 AND id <= $2 ORDER BY id DESC`
	rows, err := r.conn.Query(query, fromMillis, toMillis)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve config histories in interval: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var history ConfigHistory
		err := rows.Scan(&history.ID, &history.Commentary, &history.Type, &history.User)
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

	query := `SELECT id, commentary, type, user FROM config_history_v1 WHERE type = $1 ORDER BY id DESC`
	rows, err := r.conn.Query(query, historyType)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve config histories by type: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var history ConfigHistory
		err := rows.Scan(&history.ID, &history.Commentary, &history.Type, &history.User)
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

	query := `SELECT id, commentary, type, user FROM config_history_v1 WHERE user = $1 ORDER BY id DESC`
	rows, err := r.conn.Query(query, user)
	if err != nil {
		return nil, fmt.Errorf("couldn't retrieve config histories by user: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var history ConfigHistory
		err := rows.Scan(&history.ID, &history.Commentary, &history.Type, &history.User)
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
