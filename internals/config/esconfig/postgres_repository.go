package esconfig

import (
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"strings"
)

const table = "elasticsearch_config_v1"

// PostgresRepository is a repository containing the ExternalConfig definition based on a PSQL database and
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

// newStatement creates a new statement builder with Dollar format
func (r *PostgresRepository) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(r.conn.DB)
}

// checkRowsAffected Check if nbRows where affected to db
func (r *PostgresRepository) checkRowsAffected(res sql.Result, nbRows int64) error {
	i, err := res.RowsAffected()
	if err != nil {
		return errors.New("error with the affected rows:" + err.Error())
	}
	if i != nbRows {
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// Get use to retrieve an elasticSearchConfig by id
func (r *PostgresRepository) Get(id int64) (models.ElasticSearchConfig, bool, error) {
	rows, err := r.newStatement().
		Select("name", "urls", "default").
		From(table).
		Where(sq.Eq{"id": id}).
		Query()
	if err != nil {
		return models.ElasticSearchConfig{}, false, err
	}
	defer rows.Close()

	var name, urls string
	var isDefault bool
	if rows.Next() {
		err := rows.Scan(&name, &urls, &isDefault)
		if err != nil {
			return models.ElasticSearchConfig{}, false, fmt.Errorf("couldn't scan the elasticsearch config with id %d: %s", id, err.Error())
		}
	} else {
		return models.ElasticSearchConfig{}, false, nil
	}

	return models.ElasticSearchConfig{
		Id:      id,
		Name:    name,
		URLs:    strings.Split(urls, ","),
		Default: isDefault,
	}, true, nil
}

// GetByName use to retrieve an elasticSearchConfig by name
func (r *PostgresRepository) GetByName(name string) (models.ElasticSearchConfig, bool, error) {
	rows, err := r.newStatement().
		Select("id", "urls", "default").
		From(table).
		Where(sq.Eq{"name": name}).
		Query()
	if err != nil {
		return models.ElasticSearchConfig{}, false, err
	}
	defer rows.Close()

	esConfig := models.ElasticSearchConfig{
		Name: name,
	}
	var urls string
	if rows.Next() {
		err = rows.Scan(&esConfig.Id, &urls, &esConfig.Default)
		if err != nil {
			return models.ElasticSearchConfig{}, false, fmt.Errorf("couldn't scan the elasticsearch config with name %s: %s", name, err.Error())
		}
	} else {
		return models.ElasticSearchConfig{}, false, nil
	}

	esConfig.URLs = strings.Split(urls, ",")

	return esConfig, true, nil
}

// Create method used to create an elasticSearchConfig
func (r *PostgresRepository) Create(elasticSearchConfig models.ElasticSearchConfig) (int64, error) {
	var id int64
	err := r.newStatement().
		Insert(table).
		Columns("name", "urls", "default").
		Values(elasticSearchConfig.Name, strings.Join(elasticSearchConfig.URLs, ","), elasticSearchConfig.Default).
		Suffix("RETURNING \"id\"").
		QueryRow().
		Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

// Update method used to update un elasticSearchConfig
func (r *PostgresRepository) Update(id int64, elasticSearchConfig models.ElasticSearchConfig) error {
	res, err := r.newStatement().
		Update(table).
		Set("name", elasticSearchConfig.Name).
		Set("urls", strings.Join(elasticSearchConfig.URLs, ",")).
		Set("default", elasticSearchConfig.Default).
		Where("id = ?", id).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowsAffected(res, 1)
}

// Delete use to retrieve an elasticSearchConfig by name
func (r *PostgresRepository) Delete(id int64) error {
	res, err := r.newStatement().
		Delete(table).
		Where("id = ?", id).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowsAffected(res, 1)

	// TODO: check & set default
}

// GetAll method used to get all elasticSearchConfigs
func (r *PostgresRepository) GetAll() (map[int64]models.ElasticSearchConfig, error) {
	elasticSearchConfigs := make(map[int64]models.ElasticSearchConfig)
	rows, err := r.newStatement().
		Select("id", "name", "urls", "default").
		From(table).
		Query()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		esConfig := models.ElasticSearchConfig{}
		var urls string
		err = rows.Scan(&esConfig.Id, &esConfig.Name, &urls, &esConfig.Default)
		if err != nil {
			return nil, err
		}

		esConfig.URLs = strings.Split(urls, ",")
		elasticSearchConfigs[esConfig.Id] = esConfig
	}
	return elasticSearchConfigs, nil
}
