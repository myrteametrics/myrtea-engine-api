package users

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const table = "users_v4"

var fields = []string{"id", "login", "created", "last_name", "first_name", "email", "phone"}

// PostgresRepository is a repository containing the user users data based on a PSQL database and
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

// Get search and returns an User from the repository by its id
func (r *PostgresRepository) Get(userUUID uuid.UUID) (User, bool, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where("id = ?", userUUID).
		Query()
	if err != nil {
		return User{}, false, err
	}
	defer rows.Close()
	return r.scanFirst(rows)
}

// Create creates a new User in the repository
func (r *PostgresRepository) Create(user UserWithPassword) (uuid.UUID, error) {
	newUUID := uuid.New()
	_, err := r.newStatement().
		Insert(table).
		Columns(append(fields, "password")...).
		Values(newUUID,
			user.Login,
			time.Now(),
			user.LastName,
			user.FirstName,
			user.Email,
			user.Phone,
			sq.Expr("crypt(? , gen_salt('bf'))", user.Password),
		).
		Exec()
	if err != nil {
		return uuid.UUID{}, err
	}
	return newUUID, nil
}

// Update updates an User in the repository
func (r *PostgresRepository) Update(user User) error {
	result, err := r.newStatement().
		Update(table).
		Set("login", user.Login).
		Set("last_name", user.LastName).
		Set("first_name", user.FirstName).
		Set("email", user.Email).
		Set("phone", user.Phone).
		Where("id = ?", user.ID).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowAffected(result, 1)
}

// Update updates an User in the repository
func (r *PostgresRepository) UpdateWithPassword(user UserWithPassword) error {
	result, err := r.newStatement().
		Update(table).
		Set("login", user.Login).
		Set("last_name", user.LastName).
		Set("first_name", user.FirstName).
		Set("email", user.Email).
		Set("phone", user.Phone).
		Set("password", sq.Expr("crypt(? ,gen_salt('md5'))", user.Password)).
		Where("id = ?", user.ID).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowAffected(result, 1)
}

// Delete deletes an User in the repository
func (r *PostgresRepository) Delete(uuid uuid.UUID) error {
	result, err := r.newStatement().
		Delete(table).
		Where("id = ?", uuid).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowAffected(result, 1)
}

// GetAll returns all Users in the repository
func (r *PostgresRepository) GetAll() ([]User, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAll(rows)
}

// GetAll returns all Users in the repository
func (r *PostgresRepository) GetAllForRole(roleUUID uuid.UUID) ([]User, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		InnerJoin("users_roles_v4 on users_v4.id = users_roles_v4.user_id").
		Where("users_roles_v4.role_id = ?", roleUUID).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAll(rows)
}

func (r *PostgresRepository) SetUserRoles(userUUID uuid.UUID, roleUUIDs []uuid.UUID) error {
	ctx := context.Background()
	tx, err := r.conn.BeginTx(ctx, nil)
	result, err := r.newTransactionStatement(tx).Delete("users_roles_v4").Where("user_id = ?", userUUID).Exec()
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt := r.newTransactionStatement(tx).Insert("users_roles_v4").Columns("user_id", "role_id")
	for _, roleUUID := range roleUUIDs {
		stmt = stmt.Values(userUUID, roleUUID)
	}
	result, err = stmt.Exec()
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := r.checkRowAffected(result, int64(len(roleUUIDs))); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// CreateSuperUserIfNotExists creates the superuser if it does not exist
// Make sure to change the password in production
func (r *PostgresRepository) CreateSuperUserIfNotExists() error {
	rows, err := r.newStatement().
		Select("id").
		From(table).
		Where("login = ?", "admin").
		Query()
	if err != nil {
		return err
	}

	defer rows.Close()
	if rows.Next() {
		return nil
	}

	// if not, we create it
	_, err = r.Create(UserWithPassword{
		User: User{
			ID:        uuid.New(),
			Login:     "admin",
			Created:   time.Now(),
			LastName:  "admin",
			FirstName: "admin",
			Email:     "admin@myrtea.ai",
			Phone:     "",
		},
		Password: "myrtea",
	})

	return err
}

func (r *PostgresRepository) newTransactionStatement(tx *sql.Tx) sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(tx)
}

func (r *PostgresRepository) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(r.conn.DB)
}

func (r *PostgresRepository) scanFirst(rows *sql.Rows) (User, bool, error) {
	if rows.Next() {
		user, err := r.scan(rows)
		return user, err == nil, err
	}
	return User{}, false, nil
}

func (r *PostgresRepository) scanAll(rows *sql.Rows) ([]User, error) {
	users := make([]User, 0)
	for rows.Next() {
		user, err := r.scan(rows)
		if err != nil {
			return []User{}, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *PostgresRepository) checkRowAffected(result sql.Result, nbRows int64) error {
	i, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if i != nbRows {
		return errors.New("no row deleted (or multiple row deleted) instead of 1 row")
	}
	return nil
}

func (r *PostgresRepository) scan(rows *sql.Rows) (User, error) {
	user := User{}
	err := sqlx.StructScan(rows, &user)
	// todo: fix this
	if err != nil {
		return User{}, errors.New("couldn't scan the retrieved data: " + err.Error())
	}
	return user, nil
}
