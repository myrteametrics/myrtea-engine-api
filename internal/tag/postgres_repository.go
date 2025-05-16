package tag

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/utils/dbutils"
	"go.uber.org/zap"
	"time"
)

const table = "tags_v1"
const tableSituations = "tags_situations_v1"
const tableTemplateInstances = "tags_situation_template_instances_v1"

var fields = []string{"id", "name", "description", "color", "created_at", "updated_at"}
var fieldsPrefix = []string{"t.id", "t.name", "t.description", "t.color", "t.created_at", "t.updated_at"}

type PostgresRepository struct {
	conn *sqlx.DB
}

// NewPostgresRepository returns a new instance of PostgresRepository
func NewPostgresRepository(conn *sqlx.DB) Repository {
	r := PostgresRepository{
		conn: conn,
	}
	var ifm Repository = &r
	return ifm
}

func (r *PostgresRepository) Create(tag Tag) (int64, error) {
	_, _, _ = r.refreshNextIdGen()
	var id int64
	now := time.Now()
	statement := r.newStatement().
		Insert(table).
		Columns("name", "description", "color", "created_at", "updated_at").
		Values(tag.Name, tag.Description, tag.Color, now, now).
		Suffix("RETURNING \"id\"")
	if tag.Id != 0 {
		statement = statement.
			Columns("id", "name", "description", "color", "created_at", "updated_at").
			Values(tag.Id, tag.Name, tag.Description, tag.Color, now, now)
	}
	err := statement.QueryRow().Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (r *PostgresRepository) Get(id int64) (Tag, bool, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Where(sq.Eq{"id": id}).
		Query()
	if err != nil {
		return Tag{}, false, err
	}
	defer rows.Close()
	return dbutils.ScanFirst(rows, r.scan)
}

func (r *PostgresRepository) Update(tag Tag) error {
	_, err := r.newStatement().
		Update(table).
		Set("name", tag.Name).
		Set("description", tag.Description).
		Set("color", tag.Color).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": tag.Id}).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) Delete(id int64) error {
	_, err := r.newStatement().
		Delete(table).
		Where(sq.Eq{"id": id}).
		Exec()
	if err != nil {
		return err
	}
	_, _, _ = r.refreshNextIdGen()
	return nil
}

func (r *PostgresRepository) GetAll() ([]Tag, error) {
	rows, err := r.newStatement().
		Select(fields...).
		From(table).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dbutils.ScanAll(rows, r.scan)
}

func (r *PostgresRepository) CreateLinkWithSituation(tagID int64, situationID int64) error {
	_, err := r.newStatement().
		Insert(tableSituations).
		Columns("tag_id", "situation_id").
		Values(tagID, situationID).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) DeleteLinkWithSituation(tagID int64, situationID int64) error {
	_, err := r.newStatement().
		Delete(tableSituations).
		Where(sq.Eq{"tag_id": tagID, "situation_id": situationID}).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) GetTagsBySituationId(situationId int64) ([]Tag, error) {
	rows, err := r.newStatement().
		Select(fieldsPrefix...).
		From(fmt.Sprintf("%s ts", tableSituations)).
		Join(fmt.Sprintf("%s t ON ts.tag_id = t.id", table)).
		Where(sq.Eq{"ts.situation_id": situationId}).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dbutils.ScanAll(rows, r.scan)
}

func (r *PostgresRepository) CreateLinkWithTemplateInstance(tagID int64, templateInstanceID int64) error {
	_, err := r.newStatement().
		Insert(tableTemplateInstances).
		Columns("tag_id", "template_instance_id").
		Values(tagID, templateInstanceID).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) DeleteLinkWithTemplateInstance(tagID int64, templateInstanceID int64) error {
	_, err := r.newStatement().
		Delete(tableTemplateInstances).
		Where(sq.Eq{"tag_id": tagID, "template_instance_id": templateInstanceID}).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) GetTagsByTemplateInstanceId(templateInstanceId int64) ([]Tag, error) {
	rows, err := r.newStatement().
		Select(fieldsPrefix...).
		From(fmt.Sprintf("%s ts", tableTemplateInstances)).
		Join(fmt.Sprintf("%s t ON ts.tag_id = t.id", table)).
		Where(sq.Eq{"ts.template_instance_id": templateInstanceId}).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return dbutils.ScanAll(rows, r.scan)
}

func (r *PostgresRepository) GetSituationsTags() (map[int64][]Tag, error) {
	rows, err := r.newStatement().
		Select("situation_id", "array_agg(t.id) as tags").
		From(fmt.Sprintf("%s ts", tableSituations)).
		Join(fmt.Sprintf("%s t ON ts.tag_id = t.id", table)).
		GroupBy("situation_id").
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	situationsTags := make(map[int64][]Tag)
	for rows.Next() {
		var situationId int64
		var tagIds []int64
		err = rows.Scan(&situationId, &tagIds)
		if err != nil {
			return nil, err
		}

		var tags []Tag

		// get all tags from db
		rows, err = r.newStatement().
			Select(fields...).
			From(table).
			Where(sq.Eq{"t.id": tagIds}).
			Query()
		if err != nil {
			return nil, err
		}

		for i := 0; rows.Next(); i++ {
			tag, err := r.scan(rows)
			if err != nil {
				return nil, err
			}
			tags = append(tags, tag)
		}

		_ = rows.Close()

		situationsTags[situationId] = tags
	}

	return situationsTags, nil
}

// scan scans a row into a Tag struct
func (r *PostgresRepository) scan(rows *sql.Rows) (Tag, error) {
	var tag Tag
	err := rows.Scan(&tag.Id, &tag.Name, &tag.Description, &tag.Color, &tag.CreatedAt, &tag.UpdatedAt)
	if err != nil {
		return Tag{}, err
	}
	return tag, nil
}

// newStatement creates a new statement builder with Dollar format
func (r *PostgresRepository) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(r.conn.DB)
}

func (r *PostgresRepository) refreshNextIdGen() (int64, bool, error) {
	query := `SELECT setval(pg_get_serial_sequence('fact_definition_v1', 'id'), coalesce(max(id),0) + 1, false) FROM fact_definition_v1`
	rows, err := r.conn.Query(query)

	if err != nil {
		zap.L().Error("Couldn't query the database:", zap.Error(err))
		return 0, false, err
	}
	defer rows.Close()

	var data int64
	if rows.Next() {
		err := rows.Scan(&data)
		if err != nil {
			return 0, false, err
		}
		return data, true, nil
	} else {
		return 0, false, nil
	}
}
