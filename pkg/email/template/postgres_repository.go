package template

import (
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/utils"
)

// PostgresRepository is a repository containing the email template definitions based on a PSQL database
// and implementing the repository interface
type PostgresRepository struct {
	conn *sqlx.DB
}

const table = "mail_templates_v1"

// NewPostgresRepository returns a new instance of PostgresRepository
func NewPostgresRepository(dbClient *sqlx.DB) Repository {
	r := PostgresRepository{
		conn: dbClient,
	}
	var ifm Repository = &r
	return ifm
}

// Create creates a new email template in the repository
func (r *PostgresRepository) Create(template Template) (int64, error) {
	if err := template.Validate(); err != nil {
		return -1, err
	}
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)

	statement := newStatement().
		Insert(table).
		Suffix("RETURNING \"id\"")
	if template.Id != 0 {
		statement = statement.
			Columns("id", "name", "description", "subject", "body_html").
			Values(template.Id, template.Name, template.Description, template.Subject, template.BodyHTML)
	} else {
		statement = statement.
			Columns("name", "description", "subject", "body_html").
			Values(template.Name, template.Description, template.Subject, template.BodyHTML)
	}

	rows, err := statement.RunWith(r.conn.DB).Query()
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var id int64
	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return -1, err
		}
	} else {
		return -1, errors.New("no id returning of insert template")
	}
	return id, nil
}

// Get returns an email template by its ID
func (r *PostgresRepository) Get(id int64) (Template, error) {
	getStatement := newStatement().
		Select("id", "name", "description", "subject", "body_html").
		From(table).
		Where("id = ?", id)

	rows, err := getStatement.RunWith(r.conn.DB).Query()
	if err != nil {
		return Template{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var template Template
		err = rows.Scan(&template.Id, &template.Name, &template.Description, &template.Subject, &template.BodyHTML)
		if err != nil {
			return Template{}, err
		}
		return template, nil
	}
	return Template{}, errors.New("template not found")
}

// GetByName returns an email template by its name
func (r *PostgresRepository) GetByName(name string) (Template, error) {
	getStatement := newStatement().
		Select("id", "name", "description", "subject", "body_html").
		From(table).
		Where("name = ?", name)

	rows, err := getStatement.RunWith(r.conn.DB).Query()
	if err != nil {
		return Template{}, err
	}
	defer rows.Close()

	if rows.Next() {
		var template Template
		err = rows.Scan(&template.Id, &template.Name, &template.Description, &template.Subject, &template.BodyHTML)
		if err != nil {
			return Template{}, err
		}
		return template, nil
	}
	return Template{}, errors.New("template not found")
}

// GetAll returns all email templates from the repository
func (r *PostgresRepository) GetAll() ([]Template, error) {
	getStatement := newStatement().
		Select("id", "name", "description", "subject", "body_html").
		From(table).
		OrderBy("name ASC")

	rows, err := getStatement.RunWith(r.conn.DB).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]Template, 0)
	for rows.Next() {
		var template Template
		err = rows.Scan(&template.Id, &template.Name, &template.Description, &template.Subject, &template.BodyHTML)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}
	return templates, nil
}

// Update updates an existing email template in the repository
func (r *PostgresRepository) Update(template Template) error {
	if err := template.Validate(); err != nil {
		return err
	}

	updateStatement := newStatement().
		Update(table).
		Set("name", template.Name).
		Set("description", template.Description).
		Set("subject", template.Subject).
		Set("body_html", template.BodyHTML).
		Where("id = ?", template.Id)

	res, err := updateStatement.RunWith(r.conn.DB).Exec()
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("no row updated (or multiple rows updated) instead of 1 row")
	}
	return nil
}

// Delete deletes an email template from the repository by its id
func (r *PostgresRepository) Delete(id int64) error {
	deleteStatement := newStatement().
		Delete(table).
		Where("id = ?", id)

	res, err := deleteStatement.RunWith(r.conn.DB).Exec()
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return errors.New("no row deleted (or multiple rows deleted) instead of 1 row")
	}
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)
	return nil
}
