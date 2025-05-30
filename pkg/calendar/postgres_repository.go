package calendar

import (
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/utils"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const table = "calendar_v1"

// PostgresRepository is a repository containing the user groups data based on a PSQL database and
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

// Get search and returns a Calendar from the repository by its id
func (r *PostgresRepository) Get(id int64) (Calendar, bool, error) {
	query := `SELECT id, name, description, timezone, period_data, enabled,
			  ARRAY(SELECT sub_calendar_id 
					FROM calendar_union_v1 
					WHERE calendar_id = :id 
					ORDER BY priority ASC) as unionCalendarIDs 
			  FROM calendar_v1 
			  WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return Calendar{}, false, err
	}
	defer rows.Close()

	if rows.Next() {
		var periodData string

		calendar := Calendar{}
		err = rows.Scan(&calendar.ID, &calendar.Name, &calendar.Description, &calendar.Timezone, &periodData, &calendar.Enabled, pq.Array(&calendar.UnionCalendarIDs))
		if err != nil {
			return Calendar{}, false, err
		}

		err = json.Unmarshal([]byte(periodData), &calendar.Periods)
		if err != nil {
			return Calendar{}, false, err
		}

		return calendar, true, nil
	}

	return Calendar{}, false, nil
}

// newStatement creates a new statement builder with Dollar format
func (r *PostgresRepository) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(r.conn.DB)
}

// Create method used to create a Calendar
func (r *PostgresRepository) Create(calendar Calendar) (int64, error) {
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)
	creationTS := time.Now().Truncate(1 * time.Millisecond).UTC()

	tx, err := r.conn.Begin()
	if err != nil {
		return -1, err
	}

	periodData, err := json.Marshal(calendar.Periods)
	if err != nil {
		return -1, err
	}

	// Create a statement builder for the transaction
	stmt := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(tx)

	// Build the insert query
	insertBuilder := stmt.Insert(table).
		Suffix("RETURNING \"id\"")

	// Handle the case where calendar.ID is provided
	if calendar.ID != 0 {
		insertBuilder = insertBuilder.
			Columns("id", "name", "description", "timezone", "period_data", "enabled", "creation_date", "last_modified").
			Values(calendar.ID, calendar.Name, calendar.Description, calendar.Timezone, string(periodData), calendar.Enabled, creationTS, creationTS)
	} else {
		insertBuilder = insertBuilder.
			Columns("name", "description", "timezone", "period_data", "enabled", "creation_date", "last_modified").
			Values(calendar.Name, calendar.Description, calendar.Timezone, string(periodData), calendar.Enabled, creationTS, creationTS)
	}

	// Execute the query
	var calendarID int64
	err = insertBuilder.QueryRow().Scan(&calendarID)
	if err != nil {
		tx.Rollback()
		return -1, err
	}

	//insert unions calendar
	for i, subCalendarID := range calendar.UnionCalendarIDs {
		priority := i + 1

		// Build the insert query for union calendar
		res, err := stmt.Insert("calendar_union_v1").
			Columns("calendar_id", "sub_calendar_id", "priority").
			Values(calendarID, subCalendarID, priority).
			Exec()

		if err != nil {
			tx.Rollback()
			return -1, err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			tx.Rollback()
			return -1, errors.New("error with the affected rows:" + err.Error())
		}
		if rowsAffected != 1 {
			tx.Rollback()
			return -1, errors.New("no row inserted (or multiple row inserted) instead of 1 row")
		}
	}

	err = tx.Commit()
	if err != nil {
		return -1, err
	}

	return calendarID, nil
}

func oldUnionIDs(r *PostgresRepository, calendarID int64) ([]int64, error) {
	unionCalendarIDs := make([]int64, 0)
	query := `SELECT ARRAY(SELECT sub_calendar_id 
						   FROM calendar_union_v1 
						   WHERE calendar_id = :id
						   ORDER BY priority ASC) as unionCalendarIDs 
			  FROM calendar_v1 
			  WHERE id = :id`
	params := map[string]interface{}{
		"id": calendarID,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(pq.Array(&unionCalendarIDs))
		if err != nil {
			return nil, err
		}
	}

	return unionCalendarIDs, nil
}

// Update method used to update a Calendar
func (r *PostgresRepository) Update(calendar Calendar) error {
	lasmodifiedTS := time.Now().Truncate(1 * time.Millisecond).UTC()

	oldUnionCalendarIDs, err := oldUnionIDs(r, calendar.ID)
	if err != nil {
		return err
	}

	tx, err := r.conn.Begin()
	if err != nil {
		return err
	}

	periodData, err := json.Marshal(calendar.Periods)
	if err != nil {
		return err
	}

	rows, err := tx.Query(`UPDATE calendar_v1 
						   SET name = $1, 
						       description = $2, 
							   timezone = $3,
							   period_data = $4, 
							   enabled = $5,
							   last_modified = $6
							WHERE id = $7  RETURNING id`, calendar.Name, calendar.Description, calendar.Timezone, string(periodData), calendar.Enabled, lasmodifiedTS, calendar.ID)

	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	var calendarID int64
	if rows.Next() {
		rows.Scan(&calendarID)
	} else {
		tx.Rollback()
		return errors.New("no id returning of update calendar action")
	}
	rows.Close()

	//insert unions calendar
	for i, subCalendarID := range calendar.UnionCalendarIDs {
		priority := i + 1
		if !contains(oldUnionCalendarIDs, subCalendarID) {
			res, err := tx.Exec(`INSERT INTO calendar_union_v1(calendar_id, sub_calendar_id, priority) 
			VALUES ($1,$2,$3)`, calendarID, subCalendarID, priority)
			if err != nil {
				tx.Rollback()
				return err
			}

			if err != nil {
				tx.Rollback()
				return err
			}

			i, err := res.RowsAffected()
			if err != nil {
				tx.Rollback()
				return errors.New("error with the affected rows:" + err.Error())
			}
			if i != 1 {
				tx.Rollback()
				return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
			}
		} else {
			_, err := tx.Exec(`UPDATE calendar_union_v1 SET priority = $1
			 					 WHERE calendar_id = $2 and sub_calendar_id = $3 `, priority, calendarID, subCalendarID)
			if err != nil {
				tx.Rollback()
				return err
			}

			if err != nil {
				tx.Rollback()
				return err
			}

			/* i, err := res.RowsAffected()
			if err != nil {
				tx.Rollback()
				return errors.New("error with the affected rows:" + err.Error())
			}
			if i != 1 {
				tx.Rollback()
				return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
			} */
		}
	}

	//delete unions calendar
	for _, subCalendarID := range oldUnionCalendarIDs {
		if !contains(calendar.UnionCalendarIDs, subCalendarID) {
			res, err := tx.Exec(`DELETE FROM calendar_union_v1 WHERE calendar_id = $1 and sub_calendar_id = $2`, calendarID, subCalendarID)
			if err != nil {
				tx.Rollback()
				return err
			}

			if err != nil {
				tx.Rollback()
				return err
			}

			i, err := res.RowsAffected()
			if err != nil {
				tx.Rollback()
				return errors.New("error with the affected rows:" + err.Error())
			}
			if i != 1 {
				tx.Rollback()
				return errors.New("no row deleted (or multiple row inserted) instead of 1 row")
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Delete method used to delete a Calendar
func (r *PostgresRepository) Delete(id int64) error {
	query := `DELETE FROM calendar_v1 WHERE id = :id`
	res, err := r.conn.NamedExec(query, map[string]interface{}{
		"id": id,
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
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)
	return nil
}

// GetAll method used to get all Calendars
func (r *PostgresRepository) GetAll() (map[int64]Calendar, error) {
	calendars := make(map[int64]Calendar, 0)

	query := `SELECT id, name, description, timezone, period_data, enabled,
			  ARRAY(SELECT sub_calendar_id 
					FROM calendar_union_v1 
					WHERE calendar_id = c.id
					ORDER BY priority ASC) as unionCalendarIDs
			  FROM calendar_V1 as c`

	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, errors.New("couldn't retrieve the calendars " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var periodData string
		var calendar Calendar

		err := rows.Scan(&calendar.ID, &calendar.Name, &calendar.Description, &calendar.Timezone, &periodData, &calendar.Enabled, pq.Array(&calendar.UnionCalendarIDs))
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(periodData), &calendar.Periods)
		if err != nil {
			zap.L().Error("Couldn't unmarshall the calendar periods:", zap.Error(err))
			return nil, err
		}
		calendars[calendar.ID] = calendar
	}

	return calendars, nil
}

// GetAllModifiedFrom returns all entities that have been modified since 'from' parameter
func (r *PostgresRepository) GetAllModifiedFrom(from time.Time) (map[int64]Calendar, error) {
	calendars := make(map[int64]Calendar, 0)

	query := `SELECT id, name, description, timezone, period_data, enabled,
			  ARRAY(SELECT sub_calendar_id 
				FROM calendar_union_v1 
				WHERE calendar_id = c.id
				ORDER BY priority ASC) as unionCalendarIDs
			  FROM calendar_V1 as c
			  WHERE last_modified >= :last_modified`

	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"last_modified": from.Truncate(1 * time.Millisecond).UTC(),
	})

	if err != nil {
		return nil, errors.New("couldn't retrieve the calendars " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var periodData string
		var calendar Calendar

		err := rows.Scan(&calendar.ID, &calendar.Name, &calendar.Description, &calendar.Timezone, &periodData, &calendar.Enabled, pq.Array(&calendar.UnionCalendarIDs))
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(periodData), &calendar.Periods)
		if err != nil {
			zap.L().Error("Couldn't unmarshall the calendar periods:", zap.Error(err))
			return nil, err
		}
		calendars[calendar.ID] = calendar
	}

	return calendars, nil
}

// GetSituationCalendar search and returns a Calendar from the repository by the situation id
func (r *PostgresRepository) GetSituationCalendar(id int64) (Calendar, bool, error) {
	query := `SELECT id, name, description, timezone, period_data, enabled,
			  ARRAY(SELECT sub_calendar_id 
					FROM calendar_union_v1 
					WHERE calendar_id = :id 
					ORDER BY priority ASC) as unionCalendarIDs 
			  FROM calendar_v1 
			  WHERE id = (SELECT s.calendar_id FROM situation_definition_v1 s WHERE s.id = :id)`
	params := map[string]interface{}{
		"id": id,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return Calendar{}, false, err
	}
	defer rows.Close()

	if rows.Next() {
		var periodData string

		calendar := Calendar{}
		err = rows.Scan(&calendar.ID, &calendar.Name, &calendar.Description, &calendar.Timezone, &periodData, &calendar.Enabled, pq.Array(&calendar.UnionCalendarIDs))
		if err != nil {
			return Calendar{}, false, err
		}

		err = json.Unmarshal([]byte(periodData), &calendar.Periods)
		if err != nil {
			return Calendar{}, false, err
		}

		return calendar, true, nil
	}

	return Calendar{}, false, nil
}
