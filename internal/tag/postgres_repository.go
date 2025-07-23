package tag

import (
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/utils/dbutils"
	"github.com/myrteametrics/myrtea-sdk/v5/repositories/utils"
	"strconv"
	"strings"
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
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)
	var id int64
	now := time.Now()
	statement := r.newStatement().
		Insert(table).
		Suffix("RETURNING \"id\"")
	if tag.Id != 0 {
		statement = statement.
			Columns("id", "name", "description", "color", "created_at", "updated_at").
			Values(tag.Id, tag.Name, tag.Description, tag.Color, now, now)
	} else {
		statement = statement.
			Columns("name", "description", "color", "created_at", "updated_at").
			Values(tag.Name, tag.Description, tag.Color, now, now)
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
	_, _, _ = utils.RefreshNextIdGen(r.conn.DB, table)
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
		var tagIdsRaw []byte
		err = rows.Scan(&situationId, &tagIdsRaw)
		if err != nil {
			return nil, err
		}

		tagIds, err := parseInt64Array(tagIdsRaw)
		if err != nil {
			return nil, err
		}

		var tags []Tag

		if len(tagIds) > 0 {
			rows2, err := r.newStatement().
				Select(fields...).
				From(table).
				Where(sq.Eq{"id": tagIds}).
				Query()
			if err != nil {
				return nil, err
			}

			for rows2.Next() {
				tag, err := r.scan(rows2)
				if err != nil {
					_ = rows2.Close()
					return nil, err
				}
				tags = append(tags, tag)
			}
			_ = rows2.Close()
		}

		situationsTags[situationId] = tags
	}

	return situationsTags, nil
}

// GetSituationInstanceTags returns all tags linked to each situation instance for a given situationId
func (r *PostgresRepository) GetSituationInstanceTags(situationId int64) (map[int64][]Tag, error) {
	// Get all situation instance IDs linked to the situationId
	rows, err := r.newStatement().
		Select("id").
		From("situation_template_instances_v1").
		Where(sq.Eq{"situation_id": situationId}).
		Query()
	if err != nil {
		return nil, err
	}

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	_ = rows.Close()

	result := make(map[int64][]Tag)
	if len(ids) == 0 {
		return result, nil
	}

	// Get all tags linked to these situation instances
	rows, err = r.newStatement().
		Select(append(fieldsPrefix, "sti.situation_template_instance_id")...).
		From(fmt.Sprintf("%s sti", tableTemplateInstances)).
		Join(fmt.Sprintf("%s t ON sti.tag_id = t.id", table)).
		Where(sq.Eq{"sti.situation_template_instance_id": ids}).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var templateInstanceId int64
		var tag Tag
		err := rows.Scan(
			&tag.Id, &tag.Name, &tag.Description, &tag.Color, &tag.CreatedAt, &tag.UpdatedAt,
			&templateInstanceId,
		)
		if err != nil {
			return nil, err
		}
		result[templateInstanceId] = append(result[templateInstanceId], tag)
	}

	return result, nil
}

// parseInt64Array parses a Postgres int8[] array in byte slice form (e.g., "{1,2,3}") to []int64
func parseInt64Array(b []byte) ([]int64, error) {
	s := string(b)
	s = strings.Trim(s, "{}")
	if len(s) == 0 {
		return []int64{}, nil
	}
	parts := strings.Split(s, ",")
	result := make([]int64, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
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
