package permissions

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	uuid "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const table = "permissions_v4"

var fields = []string{"id", "resource_type", "resource_id", "action"}

// PostgresRepository is a repository containing the user permissions data based on a PSQL database and
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

//Get search and returns an User Permission from the repository by its id
func (r *PostgresRepository) Get(permissionUUID uuid.UUID) (Permission, bool, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where("id = ?", permissionUUID).
		Query()
	if err != nil {
		return Permission{}, false, err
	}
	defer rows.Close()
	return r.scanFirst(rows)
}

// Create creates a new User Permission in the repository
func (r *PostgresRepository) Create(permission Permission) (uuid.UUID, error) {
	newUUID := uuid.New()
	_, err := r.newStatement().
		Insert(table).
		Columns(fields...).
		Values(newUUID, permission.ResourceType, permission.ResourceID, permission.Action).
		Exec()
	if err != nil {
		return uuid.UUID{}, err
	}
	return newUUID, nil
}

// Update updates an User Permission in the repository
func (r *PostgresRepository) Update(permission Permission) error {
	result, err := r.newStatement().
		Update(table).
		Set("resource_type", permission.ResourceType).
		Set("resource_id", permission.ResourceID).
		Set("action", permission.Action).
		Where("id = ?", permission.ID).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowAffected(result, 1)
}

// Delete deletes an User Permission in the repository
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

// GetAll returns all User Permissions in the repository
func (r *PostgresRepository) GetAll() ([]Permission, error) {
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

// GetAll returns all User Permissions in the repository
func (r *PostgresRepository) GetAllForRole(roleUUID uuid.UUID) ([]Permission, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		InnerJoin("roles_permissions_v4 on permissions_v4.id = roles_permissions_v4.permission_id").
		Where("roles_permissions_v4.role_id = ?", roleUUID).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAll(rows)
}

func (r *PostgresRepository) GetAllForRoles(roleUUID []uuid.UUID) ([]Permission, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		InnerJoin("roles_permissions_v4 on permissions_v4.id = roles_permissions_v4.permission_id").
		Where(sq.Eq{"roles_permissions_v4.role_id": roleUUID}).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAll(rows)
}

func (r *PostgresRepository) GetAllForUser(userUUID uuid.UUID) ([]Permission, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		InnerJoin("roles_permissions_v4 on permissions_v4.id = roles_permissions_v4.permission_id").
		InnerJoin("users_roles_v4 on roles_permissions_v4.user_id = users_roles_v4.user_id").
		Where("users_roles_v4.user_id = ?", userUUID).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return r.scanAll(rows)
}

func (r *PostgresRepository) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(r.conn.DB)
}

func (r *PostgresRepository) scanFirst(rows *sql.Rows) (Permission, bool, error) {
	if rows.Next() {
		permission, err := r.scan(rows)
		return permission, err == nil, err
	}
	return Permission{}, false, nil
}

func (r *PostgresRepository) scanAll(rows *sql.Rows) ([]Permission, error) {
	permissions := make([]Permission, 0)
	for rows.Next() {
		permission, err := r.scan(rows)
		if err != nil {
			zap.L().Warn("error", zap.Error(err))
			return []Permission{}, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
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

func (r *PostgresRepository) scan(rows *sql.Rows) (Permission, error) {
	permission := Permission{}
	err := rows.Scan(&permission.ID, &permission.ResourceType, &permission.ResourceID, &permission.Action)
	if err != nil {
		return Permission{}, errors.New("couldn't scan the retrieved data: " + err.Error())
	}
	return permission, nil
}
