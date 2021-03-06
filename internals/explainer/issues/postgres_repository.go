package issues

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/queryutils"
	"go.uber.org/zap"
)

// PostgresRepository is a repository containing the Issue definition based on a PSQL database and
//implementing the repository interface
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

//Get use to retrieve an issue by id
func (r *PostgresRepository) Get(id int64, groups []int64) (models.Issue, bool, error) {
	query := `SELECT i.id, i.key, i.name, i.level, i.situation_id, situation_instance_id, i.situation_date,
			  i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified, i.detection_rating_avg,
			  i.assigned_at, i.assigned_to, i.closed_at, i.closed_by
			  FROM issues_v1 as i
	 		  inner join situation_definition_v1 on situation_definition_v1.id = i.situation_id
			  WHERE  i.id = :id and situation_definition_v1.groups && :groups`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"id":     id,
		"groups": pq.Array(groups),
	})

	if err != nil {
		return models.Issue{}, false, err
	}
	defer rows.Close()

	var issue models.Issue
	if rows.Next() {
		issue, err = scanIssue(rows)
		if err != nil {
			return models.Issue{}, false, err
		}
	} else {
		return models.Issue{}, false, nil
	}

	return issue, true, nil
}

//Create method used to create an issue
func (r *PostgresRepository) Create(issue models.Issue) (int64, error) {
	creationTS := time.Now().Truncate(1 * time.Millisecond).UTC()
	LastModificationTS := creationTS

	ruleData, err := json.Marshal(issue.Rule)
	if err != nil {
		return -1, err
	}

	query := `INSERT into issues_v1 (id, key, name, level, situation_id, situation_instance_id, situation_date, expiration_date, rule_data, state, created_at, last_modified, detection_rating_avg)
			  values (DEFAULT, :key, :name, :level, :situation_id, :situation_instance_id, :situation_date, :expiration_date, :rule_data, :state, :created_at, :last_modified, :detection_rating_avg) RETURNING id`
	params := map[string]interface{}{
		"key":                   issue.Key,
		"name":                  issue.Name,
		"level":                 issue.Level.String(),
		"situation_id":          issue.SituationID,
		"situation_instance_id": issue.TemplateInstanceID,
		"situation_date":        issue.SituationTS,
		"expiration_date":       issue.ExpirationTS,
		"rule_data":             string(ruleData),
		"state":                 issue.State.String(),
		"created_at":            creationTS,
		"last_modified":         LastModificationTS,
		"detection_rating_avg":  -1,
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
		return -1, errors.New("No id returning of insert situation")
	}

	return id, nil
}

//Update method used to update an issue
func (r *PostgresRepository) Update(tx *sqlx.Tx, id int64, issue models.Issue, user groups.UserWithGroups) error {
	LastModificationTS := time.Now().Truncate(1 * time.Millisecond).UTC()

	//Here we exclude some fields that are not to be updated
	query := `UPDATE issues_v1 SET name = :name, expiration_date = :expiration_date,
	state = :state, last_modified = :last_modified`

	if issue.State == models.ClosedNoFeedback || issue.State == models.ClosedFeedback {
		query = query + `, closed_at = :ts, closed_by = :user`
	}
	if issue.State == models.Draft && issue.AssignedAt == nil {
		query = query + `, assigned_at = :ts, assigned_to = :user`
	}
	query = query + ` WHERE id = :id`

	params := map[string]interface{}{
		"id":              id,
		"name":            issue.Name,
		"expiration_date": issue.ExpirationTS,
		"state":           issue.State.String(),
		"last_modified":   LastModificationTS,
		"ts":              LastModificationTS,
		"user":            user.Login,
	}

	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.NamedExec(query, params)
	} else {
		res, err = r.conn.NamedExec(query, params)
	}

	if err != nil {
		return errors.New("Couldn't query the database:" + err.Error())
	}

	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("Error with the affected res:" + err.Error())
	}
	if i != 1 {
		return errors.New("No row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

//GetByStates method used to get all issues for an user
func (r *PostgresRepository) GetByStates(issueStates []string, groups []int64) (map[int64]models.Issue, error) {
	issues := make(map[int64]models.Issue, 0)

	query := `SELECT i.id, i.key, i.name, i.level, i.situation_id, situation_instance_id, i.situation_date,
			  i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified, i.detection_rating_avg,
			  i.assigned_at, i.assigned_to, i.closed_at, i.closed_by
			  FROM issues_v1 as i
			  inner join situation_definition_v1 on situation_definition_v1.id = i.situation_id
			  WHERE  situation_definition_v1.groups && :groups `
	params := map[string]interface{}{
		"groups": pq.Array(groups),
	}

	if len(issueStates) > 0 {
		query += ` and i.state = ANY (:states)`
		params["states"] = pq.Array(issueStates)
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		zap.L().Error("Couldn't retrieve the issues states and groups ", zap.Error(err))
		return nil, errors.New("Couldn't retrieve the issues from situation id " + err.Error())
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

//GetCloseToTimeoutByKey get all issues that belong to the same situation and their
//creation time are within the timeout duration
func (r *PostgresRepository) GetCloseToTimeoutByKey(key string, firstSituationTS time.Time) (map[int64]models.Issue, error) {
	issues := make(map[int64]models.Issue, 0)

	//TODO: list of closed states should be provided and not hardcoded !!!
	query := `SELECT i.id, i.key, i.name, i.level, i.situation_id, situation_instance_id, i.situation_date,
			  i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified, i.detection_rating_avg,
			  i.assigned_at, i.assigned_to, i.closed_at, i.closed_by
			  FROM issues_v1 as i
			  WHERE key = :key and :first_situation_date < expiration_date
			  and NOT ( i.state = ANY ( :closed_states ))`

	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"key":                  key,
		"first_situation_date": firstSituationTS,
		"closed_states":        pq.Array([]string{models.ClosedFeedback.String(), models.ClosedNoFeedback.String(), models.ClosedTimeout.String(), models.ClosedDiscard.String()}),
	})
	if err != nil {
		return nil, errors.New("Couldn't retrieve the issues with key and first situation date: " + err.Error())
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

// GetAll method used to get all issues
func (r *PostgresRepository) GetAll(groups []int64) (map[int64]models.Issue, error) {
	issues := make(map[int64]models.Issue, 0)

	query := `SELECT i.id, i.key, i.name, i.level, i.situation_id, situation_instance_id, i.situation_date,
			  i.expiration_date, i.rule_data, i.state, i.created_at, i.last_modified, i.detection_rating_avg,
			  i.assigned_at, i.assigned_to, i.closed_at, i.closed_by
			  FROM issues_v1 as i
			  inner join situation_definition_v1 on situation_definition_v1.id = i.situation_id
			  WHERE situation_definition_v1.groups && :groups`
	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"groups": pq.Array(groups),
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

//ChangeState method used to change the issues state with key and created_date between from and to
func (r *PostgresRepository) ChangeState(key string, fromStates []models.IssueState, toState models.IssueState, from time.Time, to time.Time) error {
	LastModificationTS := time.Now().Truncate(1 * time.Millisecond).UTC()

	//Here we exclude some fields that are not to be updated
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
		return errors.New("Couldn't query the database:" + err.Error())
	}

	_, err = res.RowsAffected()
	if err != nil {
		return errors.New("Error with the affected res:" + err.Error())
	}
	return nil
}

// GetByStateByPage method used to get all issues
func (r *PostgresRepository) GetByStateByPage(issueStates []string, options models.SearchOptions, groups []int64) ([]models.Issue, int, error) {
	issues := make([]models.Issue, 0)

	query := `SELECT issues_v1.id, issues_v1.key, issues_v1.name, issues_v1.level,
		issues_v1.situation_id, situation_instance_id, issues_v1.situation_date,
		issues_v1.expiration_date, issues_v1.rule_data, issues_v1.state, issues_v1.created_at, issues_v1.last_modified,
		issues_v1.detection_rating_avg, issues_v1.assigned_at, issues_v1.assigned_to, issues_v1.closed_at, issues_v1.closed_by
	FROM issues_v1
			  inner join situation_definition_v1 on situation_definition_v1.id = issues_v1.situation_id
			  WHERE situation_definition_v1.groups && :groups`
	params := map[string]interface{}{
		"groups": pq.Array(groups),
	}
	if len(issueStates) > 0 {
		query += ` and issues_v1.state = ANY (:states)`
		params["states"] = pq.Array(issueStates)
	}
	if len(options.SortBy) == 0 {
		options.SortBy = []models.SortOption{{Field: "id", Order: models.Asc}}
	}

	var err error
	query, params, err = queryutils.AppendSearchOptions(query, params, options, "issues_v1")
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

	total, err := r.CountByStateByPage(issueStates, groups)
	if err != nil {
		return nil, 0, err
	}

	return issues, total, nil
}

// CountByStateByPage method used to count all issues
func (r *PostgresRepository) CountByStateByPage(issueStates []string, groups []int64) (int, error) {

	query := `select count(*)
		FROM issues_v1 as i
		inner join situation_definition_v1 on situation_definition_v1.id = i.situation_id
		WHERE situation_definition_v1.groups && :groups`
	params := map[string]interface{}{
		"groups": pq.Array(groups),
	}
	if len(issueStates) > 0 {
		query += ` and i.state = ANY (:states)`
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

func scanIssue(rows *sqlx.Rows) (models.Issue, error) {
	var ruleData string
	var issueStateString string
	var issueLevelString string
	var issue models.Issue

	err := rows.Scan(&issue.ID, &issue.Key, &issue.Name, &issueLevelString, &issue.SituationID, &issue.TemplateInstanceID,
		&issue.SituationTS, &issue.ExpirationTS, &ruleData, &issueStateString, &issue.CreationTS,
		&issue.LastModificationTS, &issue.DetectionRatingAvg, &issue.AssignedAt, &issue.AssignedTo, &issue.ClosedAt, &issue.CloseBy)
	if err != nil {
		return models.Issue{}, err
	}

	issue.State = models.ToIssueState(issueStateString)
	issue.Level = models.ToIssueLevel(issueLevelString)

	ruleData = strings.ReplaceAll(ruleData, `"errors":[{}]`, `"errors":[]`)
	ruleData = strings.ReplaceAll(ruleData, `"errors": [{}]`, `"errors": []`)

	err = json.Unmarshal([]byte(ruleData), &issue.Rule)
	if err != nil {
		zap.L().Error("Couldn't unmarshall the issue rule:", zap.Error(err))
		return models.Issue{}, err
	}
	return issue, nil
}
