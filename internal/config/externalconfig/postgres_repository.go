package externalconfig

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
	"github.com/spf13/viper"
)

const table = "external_generic_config_v1"

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

// Get use to retrieve an externalConfig by id
func (r *PostgresRepository) Get(id int64) (models.ExternalConfig, bool, error) {
	rows, err := r.newStatement().
		Select("c.name", "v.data", "v.current_version", "v.created_at").
		From("external_generic_config_v1 AS c").
		Join("external_generic_config_versions_v1 AS v ON c.id = v.config_id").
		Where(sq.Eq{"c.id": id, "v.current_version": true}).
		Query()
	if err != nil {
		return models.ExternalConfig{}, false, err
	}
	defer rows.Close()

	var name, data string
	var created_at time.Time
	var current_version bool
	if rows.Next() {
		err := rows.Scan(&name, &data, &current_version, &created_at)
		if err != nil {
			return models.ExternalConfig{}, false, fmt.Errorf("couldn't scan the action with id %d: %s", id, err.Error())
		}
	} else {
		return models.ExternalConfig{}, false, nil
	}

	return models.ExternalConfig{
		Id:             id,
		Name:           name,
		Data:           data,
		CurrentVersion: current_version,
		CreatedAt:      created_at,
	}, true, nil
}

// GetByName use to retrieve an externalConfig by name
func (r *PostgresRepository) GetByName(name string) (models.ExternalConfig, bool, error) {
	rows, err := r.newStatement().
		Select("c.id", "v.data", "v.current_version", "v.created_at").
		From("external_generic_config_v1 AS c").
		Join("external_generic_config_versions_v1 AS v ON c.id = v.config_id").
		Where(sq.Eq{"c.name": name, "v.current_version": true}).
		Query()
	if err != nil {
		return models.ExternalConfig{}, false, err
	}
	defer rows.Close()

	var id int64
	var data string
	var created_at time.Time
	var current_version bool

	if rows.Next() {
		err := rows.Scan(&id, &data, &current_version, &created_at)
		if err != nil {
			return models.ExternalConfig{}, false, fmt.Errorf("couldn't scan the action with name %s: %s", name, err.Error())
		}
	} else {
		return models.ExternalConfig{}, false, nil
	}

	return models.ExternalConfig{
		Id:             id,
		Name:           name,
		Data:           data,
		CurrentVersion: current_version,
		CreatedAt:      created_at,
	}, true, nil
}

// Create method used to create an externalConfig
func (r *PostgresRepository) Create(externalConfig models.ExternalConfig) (int64, error) {
	tx, err := r.conn.Begin() // Start a transaction
	if err != nil {
		return -1, err
	}

	var id int64

	err = r.newStatement().
		Insert("external_generic_config_v1").
		Columns("name").
		Values(externalConfig.Name).
		Suffix("RETURNING \"id\"").
		QueryRow().
		Scan(&id)
	if err != nil {
		tx.Rollback() // Cancel transaction in case of error
		return -1, err
	}

	// Insert the new version with current_version = true
	_, err = r.newStatement().
		Insert("external_generic_config_versions_v1").
		Columns("config_id", "data", "current_version").
		Values(id, externalConfig.Data, true).
		Exec()
	if err != nil {
		tx.Rollback() // Cancel transaction in case of error
		return -1, err
	}

	// Commit the transaction if everything is successful
	err = tx.Commit()
	if err != nil {
		return -1, err
	}
	return id, nil
}

// Update method used to update un externalConfig
func (r *PostgresRepository) Update(id int64, externalConfig models.ExternalConfig) error {
	tx, err := r.conn.Begin() // Start a transaction
	if err != nil {
		return err
	}

	statementBuilder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(tx)

	_, err = statementBuilder.
		Update("external_generic_config_v1").
		Set("name", externalConfig.Name).
		Where("id = ?", id).
		Exec()
	if err != nil {
		tx.Rollback() // Cancel transaction in case of error
		return err
	}

	// Update old version (current_version = false)
	_, err = statementBuilder.
		Update("external_generic_config_versions_v1").
		Set("current_version", false).
		Where("config_id = ?", id).
		Where("current_version = true"). // We target the old active version
		Exec()
	if err != nil {
		tx.Rollback() // Cancel transaction in case of error
		return err
	}

	// Insert the new version with current_version = true
	_, err = statementBuilder.
		Insert("external_generic_config_versions_v1").
		Columns("config_id", "data", "current_version").
		Values(id, externalConfig.Data, true).
		Exec()
	if err != nil {
		tx.Rollback() // Annuler la transaction en cas d'erreur
		return err
	}

	// Count total versions
	var versionCount int
	err = statementBuilder.
		Select("COUNT(*)").
		From("external_generic_config_versions_v1").
		Where("config_id = ?", id).
		QueryRow().Scan(&versionCount)
	if err != nil {
		tx.Rollback() // Cancel transaction in case of error
		return err
	}

	// Delete oldest versions if more than maxVersions (5 in this case)
	maxVersions := viper.GetInt("MAX_EXTERNAL_CONFIG_VERSIONS_TO_KEEP")

	if versionCount > maxVersions {

		subQuery, _, err := statementBuilder.
			Select("created_at").
			From("external_generic_config_versions_v1").
			Where("config_id = ?", id).
			OrderBy("created_at DESC").
			Offset(uint64(maxVersions)).
			Limit(1).
			ToSql()

		if err != nil {
			tx.Rollback() // Cancel transaction in case of error
			return err
		}

		_, err = statementBuilder.
			Delete("external_generic_config_versions_v1").
			Where("config_id = ?", id).
			Where("created_at < (" + subQuery + ")").
			Exec()

		if err != nil {
			tx.Rollback() // Cancel transaction in case of error
			return err
		}
	}

	// Commit the transaction if everything is successful
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

// Delete use to retrieve an externalConfig by name
func (r *PostgresRepository) Delete(id int64) error {
	res, err := r.newStatement().
		Delete(table).
		Where("id = ?", id).
		Exec()
	if err != nil {
		return err
	}
	return r.checkRowsAffected(res, 1)
}

// GetAll method used to get all externalConfigs
func (r *PostgresRepository) GetAll() (map[int64]models.ExternalConfig, error) {
	externalConfigs := make(map[int64]models.ExternalConfig)
	rows, err := r.newStatement().
		Select("c.id", "c.name", "v.data", "v.current_version", "v.created_at").
		From("external_generic_config_v1 AS c").
		Join("external_generic_config_versions_v1 AS v ON c.id = v.config_id").
		Where(sq.Eq{"v.current_version": true}).
		Query()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name, data string
		var created_at time.Time
		var current_version bool

		err := rows.Scan(&id, &name, &data, &current_version, &created_at)
		if err != nil {
			return nil, err
		}

		externalConfig := models.ExternalConfig{
			Id:             id,
			Name:           name,
			Data:           data,
			CreatedAt:      created_at,
			CurrentVersion: current_version,
		}

		externalConfigs[externalConfig.Id] = externalConfig
	}
	return externalConfigs, nil
}

// GetAllOldVersions retrieves all non-current versions of an ExternalConfig by its id
func (r *PostgresRepository) GetAllOldVersions(id int64) ([]models.ExternalConfig, error) {
	rows, err := r.newStatement().
		Select("c.name", "v.data", "v.current_version", "v.created_at").
		From("external_generic_config_v1 AS c").
		Join("external_generic_config_versions_v1 AS v ON c.id = v.config_id").
		Where(sq.Eq{"c.id": id}).
		Where(sq.Eq{"v.current_version": false}). // Récupérer uniquement les anciennes versions
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var oldVersions []models.ExternalConfig

	for rows.Next() {
		var name, data string
		var created_at time.Time
		var current_version bool
		err := rows.Scan(&name, &data, &current_version, &created_at)
		if err != nil {
			return nil, fmt.Errorf("couldn't scan the action with id %d: %s", id, err.Error())
		}

		oldVersions = append(oldVersions, models.ExternalConfig{
			Id:             id,
			Name:           name,
			Data:           data,
			CurrentVersion: current_version,
			CreatedAt:      created_at,
		})
	}

	if len(oldVersions) == 0 {
		return nil, nil
	}

	return oldVersions, nil
}
