package users

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-sdk/v4/security"
)

// PostgresRepository is a repository containing the users data based on a PSQL database and
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

//Get search and returns an User from the repository by its id
func (r *PostgresRepository) Get(id int64) (security.User, bool, error) {
	query := `SELECT id, login, role, last_name, first_name, email, phone FROM users_v1 WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return security.User{}, false, err
	}
	defer rows.Close()

	if rows.Next() {
		user := security.User{}
		err = rows.Scan(&user.ID, &user.Login, &user.Role, &user.LastName, &user.FirstName, &user.Email, &user.Phone)
		if err != nil {
			return security.User{}, false, err
		}
		return user, true, nil
	}
	return security.User{}, false, nil
}

// Create creates a new User in the repository
func (r *PostgresRepository) Create(userWithPass security.UserWithPassword) (int64, error) {
	query := `INSERT INTO users_v1 (id, login, password, role, created, last_name, first_name, email, phone)
		VALUES (DEFAULT, :login, crypt(:password ,gen_salt('md5')), :role, :created, :last_name, :first_name, :email, :phone) RETURNING id`
	params := map[string]interface{}{
		"login":      userWithPass.Login,
		"password":   userWithPass.Password,
		"role":       userWithPass.Role,
		"created":    time.Now().Truncate(1 * time.Millisecond).UTC(),
		"last_name":  userWithPass.LastName,
		"first_name": userWithPass.FirstName,
		"email":      userWithPass.Email,
		"phone":      userWithPass.Phone,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return -1, errors.New("Couldn't create the User " + err.Error())
	}
	defer rows.Close()

	var id int64
	if rows.Next() {
		rows.Scan(&id)
	} else {
		return -1, errors.New("Error creating User " + err.Error())
	}
	return id, nil
}

// Update updates an User in the repository
func (r *PostgresRepository) Update(userWithPass security.User) error {
	query := `UPDATE users_v1 
		SET login = :login, role = :role, 
		last_name = :last_name, first_name = :first_name, email = :email, phone = :phone 
		WHERE id = :id`
	params := map[string]interface{}{
		"id":         userWithPass.ID,
		"login":      userWithPass.Login,
		"role":       userWithPass.Role,
		"last_name":  userWithPass.LastName,
		"first_name": userWithPass.FirstName,
		"email":      userWithPass.Email,
		"phone":      userWithPass.Phone,
	}
	res, err := r.conn.NamedExec(query, params)
	if err != nil {
		return errors.New("Couldn't update the User:" + err.Error())
	}
	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("Error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("No row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// UpdatePassword updates an User password in the repository
func (r *PostgresRepository) UpdatePassword(userWithPass security.UserWithPassword) error {
	query := `UPDATE users_v1 SET password = crypt(:password, gen_salt('md5')) WHERE id = :id`
	params := map[string]interface{}{
		"id":       userWithPass.ID,
		"password": userWithPass.Password,
	}
	res, err := r.conn.NamedExec(query, params)
	if err != nil {
		return errors.New("Couldn't update the User:" + err.Error())
	}
	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("Error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("No row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// Delete deletes an User in the repository
func (r *PostgresRepository) Delete(id int64) error {
	query := `DELETE FROM users_v1 WHERE id = :id`
	params := map[string]interface{}{
		"id": id,
	}
	res, err := r.conn.NamedExec(query, params)
	if err != nil {
		return errors.New("Couldn't delete the User:" + err.Error())
	}
	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("Error with the affected rows:" + err.Error())
	}
	if i != 1 {
		return errors.New("No row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// GetAll returns all users in the repository
func (r *PostgresRepository) GetAll() (map[int64]security.User, error) {
	query := `SELECT id, login, role, last_name, first_name, email, phone FROM users_v1`
	rows, err := r.conn.Query(query)
	if err != nil {
		return nil, errors.New("Couldn't retrieve the Users " + err.Error())
	}
	defer rows.Close()

	users := make(map[int64]security.User, 0)
	for rows.Next() {
		user := security.User{}
		err := rows.Scan(&user.ID, &user.Login, &user.Role, &user.LastName, &user.FirstName, &user.Email, &user.Phone)
		if err != nil {
			return nil, errors.New("Couldn't scan the retrieved data: " + err.Error())
		}
		users[user.ID] = user
	}
	return users, nil
}

// GetUsersOfGroup returns all users of a Group in the repository
func (r *PostgresRepository) GetUsersOfGroup(groupID int64) (map[int64]UserOfGroup, error) {
	query := `SELECT users_v1.id, users_v1.login, users_v1.role, users_v1.created, users_v1.last_name, users_v1.first_name, users_v1.email, users_v1.phone, 
		user_memberships_v1.role AS role_in_group 
		FROM user_memberships_v1 INNER JOIN users_v1 ON user_memberships_v1.user_id = users_v1.id 
		WHERE user_memberships_v1.group_id = :group_id;`
	params := map[string]interface{}{
		"group_id": groupID,
	}
	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		return nil, errors.New("Couldn't retrieve the Users of Group " + err.Error())
	}
	defer rows.Close()

	users := make(map[int64]UserOfGroup, 0)
	for rows.Next() {
		user := UserOfGroup{}
		err := rows.Scan(&user.ID, &user.Login, &user.Role, &user.Created, &user.LastName, &user.FirstName, &user.Email, &user.Phone, &user.RoleInGroup)
		if err != nil {
			return nil, errors.New("Couldn't scan the retrieved data: " + err.Error())
		}
		users[user.ID] = user
	}
	return users, nil
}
