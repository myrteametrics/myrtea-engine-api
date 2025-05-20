package esconfig

import (
	"database/sql"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
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
func (r *PostgresRepository) Get(id int64) (model.ElasticSearchConfig, bool, error) {
	rows, err := r.newStatement().
		Select("name", "urls", `"default"`, "export_activated").
		From(table).
		Where(sq.Eq{"id": id}).
		Query()
	if err != nil {
		return model.ElasticSearchConfig{}, false, err
	}
	defer rows.Close()

	var name, urls string
	var isDefault, exportActivated bool
	if rows.Next() {
		err := rows.Scan(&name, &urls, &isDefault, &exportActivated)
		if err != nil {
			return model.ElasticSearchConfig{}, false, fmt.Errorf("couldn't scan the elasticsearch config with id %d: %s", id, err.Error())
		}
	} else {
		return model.ElasticSearchConfig{}, false, nil
	}

	return model.ElasticSearchConfig{
		Id:              id,
		Name:            name,
		URLs:            strings.Split(urls, ","),
		Default:         isDefault,
		ExportActivated: exportActivated,
	}, true, nil
}

// GetByName use to retrieve an elasticSearchConfig by name
func (r *PostgresRepository) GetByName(name string) (model.ElasticSearchConfig, bool, error) {
	rows, err := r.newStatement().
		Select("id", "urls", `"default"`, "export_activated").
		From(table).
		Where(sq.Eq{"name": name}).
		Query()
	if err != nil {
		return model.ElasticSearchConfig{}, false, err
	}
	defer rows.Close()

	esConfig := model.ElasticSearchConfig{
		Name: name,
	}
	var urls string
	if rows.Next() {
		err = rows.Scan(&esConfig.Id, &urls, &esConfig.Default, &esConfig.ExportActivated)
		if err != nil {
			return model.ElasticSearchConfig{}, false, fmt.Errorf("couldn't scan the elasticsearch config with name %s: %s", name, err.Error())
		}
	} else {
		return model.ElasticSearchConfig{}, false, nil
	}

	esConfig.URLs = strings.Split(urls, ",")

	return esConfig, true, nil
}

// GetDefault use to retrieve the default elasticSearchConfig
func (r *PostgresRepository) GetDefault() (model.ElasticSearchConfig, bool, error) {
	rows, err := r.newStatement().
		Select("id", "name", "urls", "export_activated").
		From(table).
		Where(sq.Eq{`"default"`: true}).
		Query()
	if err != nil {
		return model.ElasticSearchConfig{}, false, err
	}
	defer rows.Close()

	esConfig := model.ElasticSearchConfig{
		Default: true,
	}
	var urls string
	if rows.Next() {
		err = rows.Scan(&esConfig.Id, &esConfig.Name, &urls, &esConfig.ExportActivated)
		if err != nil {
			return model.ElasticSearchConfig{}, false, fmt.Errorf("couldn't scan the default elasticsearch config: %s", err.Error())
		}
	} else {
		return model.ElasticSearchConfig{}, false, nil
	}

	esConfig.URLs = strings.Split(urls, ",")

	return esConfig, true, nil
}

// Create method used to create an elasticSearchConfig
func (r *PostgresRepository) Create(elasticSearchConfig model.ElasticSearchConfig) (int64, error) {
	var id int64
	err := r.newStatement().
		Insert(table).
		Columns("name", "urls", `"default"`, "export_activated").
		Values(elasticSearchConfig.Name, strings.Join(elasticSearchConfig.URLs, ","), elasticSearchConfig.Default, elasticSearchConfig.ExportActivated).
		Suffix("RETURNING \"id\"").
		QueryRow().
		Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

// Update method used to update un elasticSearchConfig
func (r *PostgresRepository) Update(id int64, elasticSearchConfig model.ElasticSearchConfig) error {
	res, err := r.newStatement().
		Update(table).
		Set("name", elasticSearchConfig.Name).
		Set("urls", strings.Join(elasticSearchConfig.URLs, ",")).
		Set(`"default"`, elasticSearchConfig.Default).
		Set("export_activated", elasticSearchConfig.ExportActivated).
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
func (r *PostgresRepository) GetAll() (map[int64]model.ElasticSearchConfig, error) {
	elasticSearchConfigs := make(map[int64]model.ElasticSearchConfig)
	rows, err := r.newStatement().
		Select("id", "name", "urls", `"default"`, "export_activated").
		From(table).
		Query()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		esConfig := model.ElasticSearchConfig{}
		var urls string
		err = rows.Scan(&esConfig.Id, &esConfig.Name, &urls, &esConfig.Default, &esConfig.ExportActivated)
		if err != nil {
			return nil, err
		}

		esConfig.URLs = strings.Split(urls, ",")
		elasticSearchConfigs[esConfig.Id] = esConfig
	}
	return elasticSearchConfigs, nil
}
