package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/utils"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const table = "job_schedules_v1"

// PostgresRepository is a repository containing the rules based on a PSQL database and
// implementing the repository interface
type PostgresRepository struct {
	conn *sqlx.DB
}

// NewPostgresRepository returns a new instance of PostgresRulesRepository
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

// Create creates a new schedule in the repository
func (r *PostgresRepository) Create(schedule InternalSchedule) (int64, error) {
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)
	timestamp := time.Now().Truncate(1 * time.Millisecond).UTC()
	scheduleData, err := json.Marshal(schedule.Job)
	if err != nil {
		return -1, errors.New("failed to marshall the InternalSchedule ID:" + fmt.Sprint(schedule.ID) +
			"\nError from Marshal" + err.Error())
	}

	var id int64
	var statement sq.InsertBuilder

	statement = r.newStatement().
		Insert(table).
		Suffix("RETURNING \"id\"")

	if schedule.ID != 0 {
		statement = statement.
			Columns("id", "name", "cronexpr", "job_type", "job_data", "last_modified", "enabled").
			Values(schedule.ID, schedule.Name, schedule.CronExpr, schedule.JobType, string(scheduleData), timestamp, schedule.Enabled)
	} else {
		statement = statement.
			Columns("name", "cronexpr", "job_type", "job_data", "last_modified", "enabled").
			Values(schedule.Name, schedule.CronExpr, schedule.JobType, string(scheduleData), timestamp, schedule.Enabled)
	}

	err = statement.QueryRow().Scan(&id)
	if err != nil {
		return -1, errors.New("couldn't query the database:" + err.Error())
	}

	return id, nil
}

// Get search and returns a job schedule from the repository by its id
func (r *PostgresRepository) Get(id int64) (InternalSchedule, bool, error) {
	query := `SELECT id, name, cronexpr, job_type, job_data, enabled FROM job_schedules_v1 WHERE id = :id`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id": id,
	})

	if err != nil {
		return InternalSchedule{}, false, errors.New("couldn't retrieve the InternalSchedule with id: " + fmt.Sprint(id) + " : " + err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		var schedule InternalSchedule
		var jobData string
		err := rows.Scan(&schedule.ID, &schedule.Name, &schedule.CronExpr, &schedule.JobType, &jobData, &schedule.Enabled)
		if err != nil {
			return InternalSchedule{}, false, errors.New("couldn't scan the retrieved data: " + err.Error())
		}

		job, err := UnmarshalInternalJob(schedule.JobType, []byte(jobData), schedule.ID)
		if err != nil {
			return InternalSchedule{}, false, err
		}
		schedule.Job = job

		return schedule, true, nil
	}
	return InternalSchedule{}, false, errors.New("internalschedule not found for ID: " + fmt.Sprint(id))
}

// Update updates a schedule in the repository by its name
func (r *PostgresRepository) Update(schedule InternalSchedule) error {

	t := time.Now().Truncate(1 * time.Millisecond).UTC()
	scheduleData, err := json.Marshal(schedule.Job)
	if err != nil {
		return errors.New("failled to marshall the InternalSchedule ID:" + fmt.Sprint(schedule.ID) +
			"\nError from Marshal" + err.Error())
	}

	query := `UPDATE job_schedules_v1 SET name = :name, cronexpr = :cronexpr, 
		job_type = :job_type, job_data = :job_data, last_modified = :last_modified, enabled = :enabled WHERE id = :id`
	res, err := r.conn.NamedExec(query, map[string]interface{}{
		"id":            schedule.ID,
		"name":          schedule.Name,
		"cronexpr":      schedule.CronExpr,
		"job_type":      schedule.JobType,
		"job_data":      string(scheduleData),
		"last_modified": t,
		"enabled":       schedule.Enabled,
	})
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

// Delete deletes an entry from the repository by it's ID
func (r *PostgresRepository) Delete(id int64) error {
	query := `DELETE FROM job_schedules_v1 WHERE id = :id`

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

// GetAll returns all job schedules in the repository
func (r *PostgresRepository) GetAll() (map[int64]InternalSchedule, error) {

	query := `SELECT id, name, cronexpr, job_type, job_data, enabled FROM job_schedules_v1`
	rows, err := r.conn.Query(query)

	if err != nil {
		zap.L().Error("Couldn't retrieve the InternalSchedules", zap.Error(err))
		return nil, errors.New("couldn't retrieve the InternalSchedules " + err.Error())
	}
	defer rows.Close()

	schedules := make(map[int64]InternalSchedule, 0)
	for rows.Next() {
		var schedule InternalSchedule
		var jobData string
		err := rows.Scan(&schedule.ID, &schedule.Name, &schedule.CronExpr, &schedule.JobType, &jobData, &schedule.Enabled)
		if err != nil {
			return nil, errors.New("couldn't scan the retrieved data: " + err.Error())
		}

		job, err := UnmarshalInternalJob(schedule.JobType, []byte(jobData), schedule.ID)
		if err != nil {
			return nil, err
		}
		schedule.Job = job

		schedules[schedule.ID] = schedule
	}
	return schedules, nil
}
