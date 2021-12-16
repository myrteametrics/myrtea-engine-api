package security

import (
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/users"
)

const table = "users_v4"

var fields = []string{"id", "login", "created", "last_name", "first_name", "email", "phone"}

// DatabaseAuth is a basic Auth implementation requiring the tuple admin/admin to authenticate successfully
type DatabaseAuth struct {
	DBClient *sqlx.DB
}

// NewDatabaseAuth returns a pointer of DatabaseAuth
func NewDatabaseAuth(DBClient *sqlx.DB) *DatabaseAuth {
	return &DatabaseAuth{DBClient}
}

func (auth *DatabaseAuth) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(auth.DBClient.DB)
}

func (auth *DatabaseAuth) scan(rows *sql.Rows) (users.User, error) {
	user := users.User{}
	err := rows.Scan(&user.ID, &user.Login, &user.Created, &user.LastName, &user.FirstName, &user.Email, &user.Phone)
	if err != nil {
		return users.User{}, errors.New("Couldn't scan the retrieved data: " + err.Error())
	}
	return user, nil
}

//Get search and returns an User from the repository by its id
func (auth *DatabaseAuth) Authenticate(login string, password string) (users.User, bool, error) {
	rows, err := auth.newStatement().
		Select(fields...).
		From(table).
		Where("login = ? AND password = crypt(?, password)", login, password).
		Query()
	if err != nil {
		return users.User{}, false, err
	}
	defer rows.Close()
	if rows.Next() {
		user, err := auth.scan(rows)
		return user, err == nil, err
	}
	return users.User{}, false, errors.New("no user found, invalid credentials")
}

// // Authenticate check the input credentials and returns a User the passwords matches
// func (auth *DatabaseAuth) Authenticate(login string, password string) (bool, users.User, error) {

// 	query := `SELECT id, login, role, last_name, first_name, email, created, phone FROM users_v1
// 		WHERE login = :login AND (password =crypt(:password, password))`
// 	params := map[string]interface{}{
// 		"login":    login,
// 		"password": password,
// 	}
// 	rows, err := auth.DBClient.NamedQuery(query, params)
// 	if err != nil {
// 		return false, users.User{}, err
// 	}
// 	defer rows.Close()

// 	var user users.User
// 	// i := 0
// 	// for rows.Next() {
// 	// 	err = rows.Scan(&user.ID, &user.Login, &user.Role, &user.LastName, &user.FirstName, &user.Email, &user.Created, &user.Phone)
// 	// 	if err != nil {
// 	// 		return false, users.User{}, err
// 	// 	}
// 	// 	i++
// 	// 	break
// 	// }
// 	// if i == 0 {
// 	// 	return false, users.User{}, errors.New("Invalid credentials")
// 	// }

// 	return true, user, nil
// }
