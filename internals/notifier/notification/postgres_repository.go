package notification

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/dbutils"
)

// PostgresRepository is a repository containing the Fact definition based on a PSQL database and
// implementing the repository interface
type PostgresRepository struct {
	conn *sqlx.DB
}

// NewPostgresRepository returns a new instance of PostgresRepository
func NewPostgresRepository(dbClient *sqlx.DB) Repository {
	r := PostgresRepository{
		conn: dbClient,
	}
	var ifm Repository = &r
	return ifm
}

// Create creates a new Notification definition in the repository
func (r *PostgresRepository) Create(notif Notification) (int64, error) {
	data, err := json.Marshal(notif)
	if err != nil {
		return -1, err
	}

	ts := time.Now().Truncate(1 * time.Millisecond).UTC()
	query := `INSERT INTO notifications_history_v1 (id, data, type, created_at) VALUES (DEFAULT, :data, :type :created_at) RETURNING id`
	params := map[string]interface{}{
		"data":       data,
		"type":       getType(notif),
		"created_at": ts,
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

// Get returns a notification by its ID
func (r *PostgresRepository) Get(id int64) (Notification, error) {

	// TODO: "ORDER BY" should be an option in dbutils.DBQueryOptionnal
	query := `SELECT id, data, isread, type FROM notifications_history_v1 WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, errors.New("couldn't retrieve any notification with this id. The query is equal to: " + err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		var id int64
		var data string
		var isRead bool
		var notifType string

		err := rows.Scan(&id, &data, &isRead, &notifType)
		if err != nil {
			return nil, errors.New("couldn't retrieve any notification. The query is equal to: " + err.Error())
		}

		t, ok := H().notificationTypes[notifType]

		if !ok {
			return nil, errors.New("notification type does not exist")
		}

		instance, err := t.NewInstance(id, []byte(data), isRead)

		if err != nil {
			return nil, errors.New("notification couldn't be instanced")
		}

		return instance, nil
	}
	return nil, errors.New("no notification found with this id")
}

// GetByRoles returns all notifications related to a certain list of roles
func (r *PostgresRepository) GetAll(queryOptionnal dbutils.DBQueryOptionnal) ([]Notification, error) {

	// TODO: "ORDER BY" should be an option in dbutils.DBQueryOptionnal
	query := `SELECT id, data, isread, type FROM notifications_history_v1`
	params := map[string]interface{}{}
	if queryOptionnal.MaxAge > 0 {
		query += ` WHERE created_at > :created_at`
		params["created_at"] = time.Now().UTC().Add(-1 * queryOptionnal.MaxAge)
	}
	query += ` ORDER BY created_at DESC`
	if queryOptionnal.Limit > 0 {
		query += ` LIMIT :limit`
		params["limit"] = queryOptionnal.Limit
	}
	if queryOptionnal.Offset > 0 {
		query += ` OFFSET :offset`
		params["offset"] = queryOptionnal.Offset
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, errors.New("couldn't retrieve any notification with these roles. The query is equal to: " + err.Error())
	}
	defer rows.Close()

	notifications := make([]Notification, 0)
	for rows.Next() {

		var id int64
		var data string
		var isRead bool
		var notifType string

		err := rows.Scan(&id, &data, &isRead, &notifType)
		if err != nil {
			return nil, errors.New("couldn't scan the notification data:" + err.Error())
		}

		t, ok := H().notificationTypes[notifType]

		if !ok {
			return nil, errors.New("notification type does not exist")
		}

		instance, err := t.NewInstance(id, []byte(data), isRead)

		if err != nil {
			return nil, errors.New("notification couldn't be instanced")
		}

		notifications = append(notifications, instance)
	}
	if err != nil {
		return nil, errors.New("deformed Data " + err.Error())
	}
	return notifications, nil
}

// Delete deletes a  notification from the repository by its id
func (r *PostgresRepository) Delete(id int64) error {
	query := `DELETE FROM notifications_history_v1 WHERE id = :id`

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
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// UpdateRead updates a notification status by changing the isRead state to true once it has been read
func (r *PostgresRepository) UpdateRead(id int64, status bool) error {
	query := `UPDATE notifications_history_v1 SET isread = :status WHERE id = :id`

	res, err := r.conn.NamedExec(query, map[string]interface{}{
		"status": status,
		"id":     id,
	})
	if err != nil {
		return err
	}
	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i != 1 {
		return errors.New("no row updated (or multiple row updated) instead of 1 row")
	}
	return nil
}
