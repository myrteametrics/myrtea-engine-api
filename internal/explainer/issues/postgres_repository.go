package issues

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/utils/queryutils"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"go.uber.org/zap"
)

// PostgresRepository is a repository containing the Issue definition based on a PSQL database and
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

// Get use to retrieve an issue by id
func (r *PostgresRepository) Get(id int64) (model.Issue, bool, error) {
	query := `SELECT i.id, i.key, i.name, i.level, i.situation_history_id, i.situation_id, situation_instance_id, i.situation_date,
			  i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified, i.detection_rating_avg,
			  i.assigned_at, i.assigned_to, i.closed_at, i.closed_by, i.comment
			  FROM issues_v1 as i
			  WHERE  i.id = :id`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id": id,
	})

	if err != nil {
		return model.Issue{}, false, err
	}
	defer rows.Close()

	var issue model.Issue
	if rows.Next() {
		issue, err = scanIssue(rows)
		if err != nil {
			return model.Issue{}, false, err
		}
	} else {
		return model.Issue{}, false, nil
	}

	return issue, true, nil
}

// Create method used to create an issue
func (r *PostgresRepository) Create(issue model.Issue) (int64, error) {
	creationTS := time.Now().Truncate(1 * time.Millisecond).UTC()
	lastModificationTS := creationTS

	ruleData, err := json.Marshal(issue.Rule)
	if err != nil {
		return -1, err
	}

	query := `INSERT into issues_v1 (id, key, name, level, situation_history_id, situation_id, situation_instance_id, situation_date, expiration_date, rule_data, state, created_at, last_modified, detection_rating_avg, comment)
			  values (DEFAULT, :key, :name, :level, :situation_history_id, :situation_id, :situation_instance_id, :situation_date, :expiration_date, :rule_data, :state, :created_at, :last_modified, :detection_rating_avg, :comment) RETURNING id`
	params := map[string]interface{}{
		"key":                   issue.Key,
		"name":                  issue.Name,
		"level":                 issue.Level.String(),
		"situation_history_id":  issue.SituationHistoryID,
		"situation_id":          issue.SituationID,
		"situation_instance_id": issue.TemplateInstanceID,
		"situation_date":        issue.SituationTS,
		"expiration_date":       issue.ExpirationTS,
		"rule_data":             string(ruleData),
		"state":                 issue.State.String(),
		"created_at":            creationTS,
		"last_modified":         lastModificationTS,
		"detection_rating_avg":  -1,
		"comment":               issue.Comment,
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var id int64
	if rows.Next() {
		rows.Scan(&id)
	} else {
		return -1, errors.New("no id returning of insert situation")
	}

	return id, nil
}

// UpdateComment method used to update an issue
func (r *PostgresRepository) UpdateComment(dbClient *sqlx.DB, id int64, comment string) error {
	lastModificationTS := time.Now().Truncate(1 * time.Millisecond).UTC()

	query := `UPDATE issues_v1 SET last_modified = :last_modified, comment = :comment WHERE id = :id`

	params := map[string]interface{}{
		"id":            id,
		"last_modified": lastModificationTS,
		"comment":       comment,
	}

	res, err := r.conn.NamedExec(query, params)
	if err != nil {
		return errors.New("couldn't query the database:" + err.Error())
	}

	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("error with the affected res:" + err.Error())
	}
	if i != 1 {
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}

	return nil
}

// Update method used to update an issue
func (r *PostgresRepository) Update(tx *sqlx.Tx, id int64, issue model.Issue, user users.User) error {
	lastModificationTS := time.Now().Truncate(1 * time.Millisecond).UTC()

	// Here we exclude some fields that are not to be updated
	query := `UPDATE issues_v1 SET name = :name, expiration_date = :expiration_date,
	state = :state, last_modified = :last_modified, comment = :comment`

	if issue.State == model.ClosedNoFeedback || issue.State == model.ClosedFeedbackConfirmed || issue.State == model.ClosedFeedbackRejected ||
		issue.State == model.ClosedConfirmed || issue.State == model.ClosedRejected {
		query = query + `, closed_at = :ts, closed_by = :user`
	}
	if issue.State == model.Draft && issue.AssignedAt == nil {
		query = query + `, assigned_at = :ts, assigned_to = :user`
	}
	query = query + ` WHERE id = :id`

	params := map[string]interface{}{
		"id":              id,
		"name":            issue.Name,
		"expiration_date": issue.ExpirationTS,
		"state":           issue.State.String(),
		"last_modified":   lastModificationTS,
		"ts":              lastModificationTS,
		"user":            user.Login,
		"comment":         issue.Comment,
	}

	var res sql.Result
	var err error

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
		return errors.New("error with the affected res:" + err.Error())
	}
	if i != 1 {
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// GetOpenAndDraftIssuesByKey get all issues that belong to the same situation in 'open' or 'draft' states
// if existed
func (r *PostgresRepository) GetOpenAndDraftIssuesByKey(key string) (map[int64]model.Issue, error) {
	issues := make(map[int64]model.Issue)

	query := `SELECT i.id, i.key, i.name, i.level,  i.situation_history_id, i.situation_id, situation_instance_id, i.situation_date,
			  i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified, i.detection_rating_avg,
			  i.assigned_at, i.assigned_to, i.closed_at, i.closed_by, i.comment
			  FROM issues_v1 as i
			  WHERE key = :key and (i.state = ANY ( :states ))`

	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"key":    key,
		"states": pq.Array([]string{model.Open.String(), model.Draft.String()}),
	})
	if err != nil {
		return nil, errors.New("couldn't retrieve the issues with key and first situation date: " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues[issue.ID] = issue
	}

	return issues, nil
}

// GetCloseToTimeoutByKey get all issues that belong to the same situation and their
// creation time are within the timeout duration
func (r *PostgresRepository) GetCloseToTimeoutByKey(key string, firstSituationTS time.Time) (map[int64]model.Issue, error) {
	issues := make(map[int64]model.Issue, 0)

	query := `SELECT i.id, i.key, i.name, i.level,  i.situation_history_id, i.situation_id, situation_instance_id, i.situation_date,
			  i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified, i.detection_rating_avg,
			  i.assigned_at, i.assigned_to, i.closed_at, i.closed_by, i.comment
			  FROM issues_v1 as i
			  WHERE key = :key and :first_situation_date < expiration_date
			  and NOT ( i.state = ANY ( :closed_states ))`

	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"key":                  key,
		"first_situation_date": firstSituationTS,
		"closed_states": pq.Array([]string{
			model.ClosedFeedbackConfirmed.String(),
			model.ClosedFeedbackRejected.String(),
			model.ClosedNoFeedback.String(),
			model.ClosedTimeout.String(),
			model.ClosedDiscard.String(),
			model.ClosedConfirmed.String(),
			model.ClosedRejected.String(),
		}),
	})
	if err != nil {
		return nil, errors.New("couldn't retrieve the issues with key and first situation date: " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues[issue.ID] = issue
	}

	return issues, nil
}

// Get used to get issues by key
func (r *PostgresRepository) GetByKeyByPage(key string, options model.SearchOptions) ([]model.Issue, int, error) {
	issues := make([]model.Issue, 0)

	query := `SELECT i.id, i.key, i.name, i.level, i.situation_history_id,
        i.situation_id, situation_instance_id, i.situation_date,
        i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified,
        i.detection_rating_avg, i.assigned_at, i.assigned_to, i.closed_at, i.closed_by, i.comment
    	FROM issues_v1 as i
		inner join situation_definition_v1 on situation_definition_v1.id = i.situation_id
		WHERE i.key = :key`

	params := map[string]interface{}{
		"key": key,
	}
	if len(options.SortBy) == 0 {
		options.SortBy = []model.SortOption{{Field: "id", Order: model.Asc}}
	}
	var err error
	query, params, err = queryutils.AppendSearchOptions(query, params, options, "i")
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, 0, err
		}
		issues = append(issues, issue)
	}

	total, err := r.CountByKeyByPage(key)
	if err != nil {
		return nil, 0, err
	}

	return issues, total, nil
}

// ChangeState method used to change the issues state with key and created_date between from and to
func (r *PostgresRepository) ChangeState(key string, fromStates []model.IssueState, toState model.IssueState) error {
	LastModificationTS := time.Now().Truncate(1 * time.Millisecond).UTC()

	// Review 2023-02-10 : We are about to close all issues, there's no need to have "key" here
	// Here we exclude some fields that are not to be updated
	// query := `UPDATE issues_v1 SET state = :to_state, last_modified = :last_modified
	//		  WHERE key = :key AND state = ANY ( :from_states )`

	query := `UPDATE issues_v1 SET state = :to_state, last_modified = :last_modified
			  WHERE state = ANY ( :from_states )`

	var states []string
	for _, state := range fromStates {
		states = append(states, state.String())
	}

	params := map[string]interface{}{
		"key":           key,
		"from_states":   pq.Array(states),
		"to_state":      toState.String(),
		"last_modified": LastModificationTS,
	}

	var res sql.Result
	var err error
	res, err = r.conn.NamedExec(query, params)

	if err != nil {
		return errors.New("couldn't query the database:" + err.Error())
	}

	_, err = res.RowsAffected()
	if err != nil {
		return errors.New("error with the affected res:" + err.Error())
	}
	return nil
}

// ChangeStateBetweenDates method used to change the issues state with key and created_date between from and to
func (r *PostgresRepository) ChangeStateBetweenDates(key string, fromStates []model.IssueState, toState model.IssueState, from time.Time, to time.Time) error {
	LastModificationTS := time.Now().Truncate(1 * time.Millisecond).UTC()

	// Here we exclude some fields that are not to be updated
	query := `UPDATE issues_v1 SET state = :to_state, last_modified = :last_modified
			  WHERE key = :key AND state = ANY ( :from_states ) AND created_at >= :from AND created_at < :to`

	var states []string
	for _, state := range fromStates {
		states = append(states, state.String())
	}

	params := map[string]interface{}{
		"key":           key,
		"from_states":   pq.Array(states),
		"to_state":      toState.String(),
		"from":          from,
		"to":            to,
		"last_modified": LastModificationTS,
	}

	var res sql.Result
	var err error
	res, err = r.conn.NamedExec(query, params)

	if err != nil {
		return errors.New("couldn't query the database:" + err.Error())
	}

	_, err = res.RowsAffected()
	if err != nil {
		return errors.New("error with the affected res:" + err.Error())
	}
	return nil
}

// GetAll method used to get all issues
func (r *PostgresRepository) GetAll() (map[int64]model.Issue, error) {
	issues := make(map[int64]model.Issue, 0)

	query := `SELECT i.id, i.key, i.name, i.level, i.situation_history_id, i.situation_id, situation_instance_id, i.situation_date,
			  i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified, i.detection_rating_avg,
			  i.assigned_at, i.assigned_to, i.closed_at, i.closed_by, i.comment
			  FROM issues_v1 as i`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{})

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues[issue.ID] = issue
	}

	return issues, nil
}

func (r *PostgresRepository) GetAllBySituationIDs(situationIDs []int64) (map[int64]model.Issue, error) {
	issues := make(map[int64]model.Issue, 0)

	query := `SELECT i.id, i.key, i.name, i.level, i.situation_history_id, i.situation_id, situation_instance_id, i.situation_date,
		  i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified, i.detection_rating_avg,
		  i.assigned_at, i.assigned_to, i.closed_at, i.closed_by, i.comment
		  FROM issues_v1 as i
		  inner join situation_definition_v1 on situation_definition_v1.id = i.situation_id
		  WHERE situation_definition_v1.id = ANY(:situation_ids)`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"situation_ids": pq.Array(situationIDs),
	})

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues[issue.ID] = issue
	}

	return issues, nil
}

// GetByStates method used to get all issues for an user
func (r *PostgresRepository) GetByStates(issueStates []string) (map[int64]model.Issue, error) {
	issues := make(map[int64]model.Issue, 0)

	query := `SELECT i.id, i.key, i.name, i.level, i.situation_history_id, i.situation_id, situation_instance_id, i.situation_date,
			  i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified, i.detection_rating_avg,
			  i.assigned_at, i.assigned_to, i.closed_at, i.closed_by, i.comment
			  FROM issues_v1 as i`
	params := map[string]interface{}{}

	if len(issueStates) > 0 {
		query += ` WHERE i.state = ANY (:states)`
		params["states"] = pq.Array(issueStates)
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		zap.L().Error("Couldn't retrieve the issues states and groups ", zap.Error(err))
		return nil, errors.New("couldn't retrieve the issues from situation id " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues[issue.ID] = issue
	}

	return issues, nil
}

// GetByStates method used to get all issues for an user
func (r *PostgresRepository) GetByStatesBySituationIDs(issueStates []string, situationIDs []int64) (map[int64]model.Issue, error) {
	issues := make(map[int64]model.Issue, 0)

	query := `SELECT i.id, i.key, i.name, i.level, i.situation_history_id, i.situation_id, situation_instance_id, i.situation_date,
			  i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified, i.detection_rating_avg,
			  i.assigned_at, i.assigned_to, i.closed_at, i.closed_by, i.comment
			  FROM issues_v1 as i
			  inner join situation_definition_v1 on situation_definition_v1.id = i.situation_id
			  WHERE situation_definition_v1.id = ANY(:situation_ids)`
	params := map[string]interface{}{
		"situation_ids": pq.Array(situationIDs),
	}

	if len(issueStates) > 0 {
		query += ` and i.state = ANY (:states)`
		params["states"] = pq.Array(issueStates)
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		zap.L().Error("Couldn't retrieve the issues states and groups ", zap.Error(err))
		return nil, errors.New("couldn't retrieve the issues from situation id " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues[issue.ID] = issue
	}

	return issues, nil
}

// GetByStateByPage method used to get all issues
func (r *PostgresRepository) GetByStateByPage(issueStates []string, options model.SearchOptions) ([]model.Issue, int, error) {
	issues := make([]model.Issue, 0)
	query := `SELECT i.id, i.key, i.name, i.level, i.situation_history_id,
		i.situation_id, situation_instance_id, i.situation_date,
		i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified,
		i.detection_rating_avg, i.assigned_at, i.assigned_to, i.closed_at, i.closed_by, i.comment
	FROM issues_v1 as i`
	params := map[string]interface{}{}
	if len(issueStates) > 0 {
		query += ` WHERE i.state = ANY (:states)`
		params["states"] = pq.Array(issueStates)
	}
	if len(options.SortBy) == 0 {
		options.SortBy = []model.SortOption{{Field: "id", Order: model.Asc}}
	}

	var err error
	query, params, err = queryutils.AppendSearchOptions(query, params, options, "i")
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, 0, err
		}
		issues = append(issues, issue)
	}

	total, err := r.CountByStateByPage(issueStates)
	if err != nil {
		return nil, 0, err
	}

	return issues, total, nil
}

// GetByStateByPage method used to get all issues
func (r *PostgresRepository) GetByStateByPageBySituationIDs(issueStates []string, options model.SearchOptions, situationIDs []int64) ([]model.Issue, int, error) {
	issues := make([]model.Issue, 0)

	query := `SELECT i.id, i.key, i.name, i.level, i.situation_history_id,
		i.situation_id, situation_instance_id, i.situation_date,
		i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified,
		i.detection_rating_avg, i.assigned_at, i.assigned_to, i.closed_at, i.closed_by, i.comment
	FROM issues_v1 as i
	inner join situation_definition_v1 on situation_definition_v1.id = i.situation_id
	WHERE situation_definition_v1.id = ANY(:situation_ids)`
	params := map[string]interface{}{
		"situation_ids": pq.Array(situationIDs),
	}
	if len(issueStates) > 0 {
		query += ` and i.state = ANY (:states)`
		params["states"] = pq.Array(issueStates)
	}
	if len(options.SortBy) == 0 {
		options.SortBy = []model.SortOption{{Field: "id", Order: model.Asc}}
	}

	var err error
	query, params, err = queryutils.AppendSearchOptions(query, params, options, "i")
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		issue, err := scanIssue(rows)
		if err != nil {
			return nil, 0, err
		}
		issues = append(issues, issue)
	}

	total, err := r.CountByStateByPageBySituationIDs(issueStates, situationIDs)
	if err != nil {
		return nil, 0, err
	}

	return issues, total, nil
}

// CountByStateByPage method used to count all issues
func (r *PostgresRepository) CountByStateByPage(issueStates []string) (int, error) {

	query := `select count(*)
		FROM issues_v1`
	params := map[string]interface{}{}
	if len(issueStates) > 0 {
		query += ` WHERE issues_v1.state = ANY (:states)`
		params["states"] = pq.Array(issueStates)
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

// CountByStateByPage method used to count all issues
func (r *PostgresRepository) CountByStateByPageBySituationIDs(issueStates []string, situationIDs []int64) (int, error) {

	query := `select count(*)
		FROM issues_v1
		inner join situation_definition_v1 on situation_definition_v1.id = issues_v1.situation_id
		WHERE situation_definition_v1.id = ANY(:situation_ids)`
	params := map[string]interface{}{
		"situation_ids": pq.Array(situationIDs),
	}
	if len(issueStates) > 0 {
		query += ` and issues_v1.state = ANY (:states)`
		params["states"] = pq.Array(issueStates)
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

// CountByKeyByPage method used to count all issues
func (r *PostgresRepository) CountByKeyByPage(key string) (int, error) {

	query := `select count(*)
        FROM issues_v1 as i
        inner join situation_definition_v1 on situation_definition_v1.id = i.situation_id
        WHERE i.key = :key`
	params := map[string]interface{}{
		"key": key,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

func scanIssue(rows *sqlx.Rows) (model.Issue, error) {
	var ruleData string
	var issueStateString string
	var issueLevelString string
	var issue model.Issue

	err := rows.Scan(
		&issue.ID,
		&issue.Key,
		&issue.Name,
		&issueLevelString,
		&issue.SituationHistoryID,
		&issue.SituationID,
		&issue.TemplateInstanceID,
		&issue.SituationTS,
		&issue.ExpirationTS,
		&ruleData,
		&issueStateString,
		&issue.CreationTS,
		&issue.LastModificationTS,
		&issue.DetectionRatingAvg,
		&issue.AssignedAt,
		&issue.AssignedTo,
		&issue.ClosedAt,
		&issue.CloseBy,
		&issue.Comment)
	if err != nil {
		return model.Issue{}, err
	}

	issue.State = model.ToIssueState(issueStateString)
	issue.Level = model.ToIssueLevel(issueLevelString)

	ruleData = strings.ReplaceAll(ruleData, `"errors":[{}]`, `"errors":[]`)
	ruleData = strings.ReplaceAll(ruleData, `"errors": [{}]`, `"errors": []`)

	err = json.Unmarshal([]byte(ruleData), &issue.Rule)
	if err != nil {
		zap.L().Error("Couldn't unmarshall the issue rule:", zap.Error(err))
		return model.Issue{}, err
	}
	return issue, nil
}

// DeleteOldIssues deletes issues based on the provided timestamp
func (r *PostgresRepository) DeleteOldIssues(ts time.Time) error {
	query := `DELETE FROM issues_v1 WHERE situation_history_id IN (SELECT id FROM situation_history_v5 WHERE ts < $1)`
	result, err := r.conn.Exec(query, ts)
	if err != nil {
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	zap.L().Info("Auto purge of the table issues_v1", zap.Int64("Number of rows deleted", affectedRows))
	return nil
}

// DeleteOldIssueDetections deletes issue detections based on the provided timestamp
func (r *PostgresRepository) DeleteOldIssueDetections(ts time.Time) error {
	query := `DELETE FROM issue_detection_feedback_v3 WHERE issue_id IN (SELECT id FROM issues_v1 WHERE situation_history_id IN (SELECT id FROM situation_history_v5 WHERE ts < $1))`
	result, err := r.conn.Exec(query, ts)
	if err != nil {
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	zap.L().Info("Auto purge of the table issue_detection_feedback_v3 ", zap.Int64("Number of rows deleted", affectedRows))
	return nil
}

// DeleteOldIssueResolutions deletes issue resolutions based on the provided timestamp
func (r *PostgresRepository) DeleteOldIssueResolutions(ts time.Time) error {
	query := `DELETE FROM issue_resolution_v1 WHERE issue_id IN (SELECT id FROM issues_v1 WHERE situation_history_id IN (SELECT id FROM situation_history_v5 WHERE ts < $1))`
	result, err := r.conn.Exec(query, ts)
	if err != nil {
		return err
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	zap.L().Info("Auto purge of the table issue_resolution_v1 ", zap.Int64("Number of rows deleted", affectedRows))

	return nil
}
