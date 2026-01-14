package roles

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	uuid "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const table = "roles_v4"

var fields = []string{"id", "name"}

// PostgresRepository is a repository containing the user roles data based on a PSQL database and
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

// Get search and returns an User Role from the repository by its id
func (r *PostgresRepository) Get(roleUUID uuid.UUID) (Role, bool, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where("id = ?", roleUUID).
		Query()
	if err != nil {
		return Role{}, false, err
	}
	defer rows.Close()
	return r.scanFirst(rows)
}

// GetByName search and returns an User Role from the repository by its id
func (r *PostgresRepository) GetByName(name string) (Role, bool, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where("name = ?", name).
		Query()
	if err != nil {
		return Role{}, false, err
	}
	defer rows.Close()
	return r.scanFirst(rows)
}

// Create creates a new User Role in the repository
func (r *PostgresRepository) Create(role Role) (uuid.UUID, error) {
	newUUID := uuid.New()
	_, err := r.newStatement().
		Insert(table).
		Columns(fields...).
		Values(newUUID, role.Name, role.HomePage).
		Exec()
	if err != nil {
		return uuid.UUID{}, err
	}
	return newUUID, nil

}

// Update updates an User Role in the repository
func (r *PostgresRepository) Update(role Role) error {
	result, err := r.newStatement().
		Update(table).
		Set("name", role.Name).
		Set("home_page", role.HomePage).
		Where("id = ?", role.ID).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowAffected(result, 1)
}

// Delete deletes an User Role in the repository
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

// GetAll returns all User Roles in the repository
func (r *PostgresRepository) GetAll() ([]Role, error) {
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

func (r *PostgresRepository) GetAllForUser(userUUID uuid.UUID) ([]Role, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		InnerJoin("users_roles_v4 on roles_v4.id = users_roles_v4.role_id").
		Where("users_roles_v4.user_id = ?", userUUID).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAll(rows)
}

func (r *PostgresRepository) SetRolePermissions(roleUUID uuid.UUID, permissionUUIDs []uuid.UUID) error {
	ctx := context.Background()
	tx, err := r.conn.BeginTx(ctx, nil)
	result, err := r.newTransactionStatement(tx).Delete("roles_permissions_v4").Where("role_id = ?", roleUUID).Exec()
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt := r.newTransactionStatement(tx).Insert("roles_permissions_v4").Columns("role_id", "permission_id")
	for _, permissionUUID := range permissionUUIDs {
		stmt = stmt.Values(roleUUID, permissionUUID)
	}
	result, err = stmt.Exec()
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := r.checkRowAffected(result, int64(len(permissionUUIDs))); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (r *PostgresRepository) newTransactionStatement(tx *sql.Tx) sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(tx)
}

func (r *PostgresRepository) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(r.conn.DB)
}

func (r *PostgresRepository) scanFirst(rows *sql.Rows) (Role, bool, error) {
	if rows.Next() {
		role, err := r.scan(rows)
		return role, err == nil, err
	}
	return Role{}, false, nil
}

func (r *PostgresRepository) scanAll(rows *sql.Rows) ([]Role, error) {
	roles := make([]Role, 0)
	for rows.Next() {
		role, err := r.scan(rows)
		if err != nil {
			return []Role{}, err
		}
		roles = append(roles, role)
	}
	return roles, nil
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

func (r *PostgresRepository) scan(rows *sql.Rows) (Role, error) {
	role := Role{}
	err := rows.Scan(&role.ID, &role.Name, &role.HomePage)
	if err != nil {
		return Role{}, errors.New("couldn't scan the retrieved data: " + err.Error())
	}
	return role, nil
}
