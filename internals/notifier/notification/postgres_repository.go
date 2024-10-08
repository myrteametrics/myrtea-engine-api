package notification

import (
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/utils/dbutils"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
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
func (r *PostgresRepository) Create(notif Notification, userLogin string) (int64, error) {
	data, err := notif.ToBytes()
	if err != nil {
		return -1, err
	}

	ts := time.Now().Truncate(1 * time.Millisecond).UTC()

	insertStatement := newStatement().
		Insert("notifications_history_v1").
		Columns("id", "data", "type", "user_login", "created_at").
		Values(sq.Expr("DEFAULT"), data, getType(notif), userLogin, ts).
		Suffix("RETURNING id")

	rows, err := insertStatement.RunWith(r.conn.DB).Query()
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
func (r *PostgresRepository) Get(id int64, userLogin string) (Notification, error) {
	getStatement := newStatement().
		Select("id", "data", "isread", "type").
		Where(sq.And{sq.Eq{"id": id}, sq.Eq{"user_login": userLogin}}).
		From("notifications_history_v1")

	rows, err := getStatement.RunWith(r.conn.DB).Query()
	if err != nil {
		return nil, errors.New("couldn't retrieve any notification with this id. The query is equal to: " + err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		var id int64
		var data string
		var isRead bool
		var notifType string

		err = rows.Scan(&id, &data, &isRead, &notifType)
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

// GetAll returns all notifications from the repository
func (r *PostgresRepository) GetAll(queryOptionnal dbutils.DBQueryOptionnal, userLogin string) ([]Notification, error) {
	getStatement := newStatement().
		Select("id", "data", "isread", "type").
		Where(sq.Eq{"user_login": userLogin}).
		From("notifications_history_v1")

	if queryOptionnal.MaxAge > 0 {
		getStatement = getStatement.Where(sq.Gt{"created_at": time.Now().UTC().Add(-1 * queryOptionnal.MaxAge)})
	}

	if queryOptionnal.Limit > 0 {
		getStatement = getStatement.Limit(uint64(queryOptionnal.Limit))
	}

	if queryOptionnal.Offset > 0 {
		getStatement = getStatement.Offset(uint64(queryOptionnal.Offset))
	}

	// TODO: "ORDER BY" should be an option in dbutils.DBQueryOptionnal
	getStatement = getStatement.OrderBy("created_at DESC")

	rows, err := getStatement.RunWith(r.conn.DB).Query()
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

		err = rows.Scan(&id, &data, &isRead, &notifType)
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
func (r *PostgresRepository) Delete(id int64, userLogin string) error {
	deleteStatement := newStatement().
		Delete("notifications_history_v1").
		Where(sq.And{sq.Eq{"id": id}, sq.Eq{"user_login": userLogin}})

	res, err := deleteStatement.RunWith(r.conn.DB).Exec()
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
func (r *PostgresRepository) UpdateRead(id int64, status bool, userLogin string) error {
	update := newStatement().
		Update("notifications_history_v1").
		Set("isread", status).
		Where(sq.And{sq.Eq{"id": id}, sq.Eq{"user_login": userLogin}})

	res, err := update.RunWith(r.conn.DB).Exec()
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

// CleanExpired deletes all notifications older than the given lifetime
func (r *PostgresRepository) CleanExpired(lifetime time.Duration) (int64, error) {
	deleteStatement := newStatement().
		Delete("notifications_history_v1").
		Where(sq.Lt{"created_at": time.Now().UTC().Add(-1 * lifetime)})

	res, err := deleteStatement.RunWith(r.conn.DB).Exec()
	if err != nil {
		return 0, err
	}
	i, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return i, nil
}
