package history

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type HistorySituationsBuilder struct{}

type GetHistorySituationsOptions struct {
	SituationID         int64
	SituationInstanceID int64
	FromTS              time.Time
	ToTS                time.Time
}

func (builder HistorySituationsBuilder) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

// situationID, situationInstanceID, tsFrom, tsTo
func (builder HistorySituationsBuilder) getHistorySituationsIdsBase(options GetHistorySituationsOptions) sq.SelectBuilder {
	q := builder.newStatement().
		Select("id").
		From("situation_history_v4")
	if options.SituationID != -1 {
		q = q.Where(sq.Eq{"situation_id": options.SituationID})
	}
	if options.SituationInstanceID != -1 {
		q = q.Where(sq.Eq{"situation_instance_id": options.SituationInstanceID})
	}
	if !options.FromTS.IsZero() {
		q = q.Where(sq.GtOrEq{"ts": options.FromTS})
	}
	if !options.ToTS.IsZero() {
		q = q.Where(sq.Lt{"ts": options.ToTS})
	}
	return q
}

func (builder HistorySituationsBuilder) GetHistorySituationsIdsLast(options GetHistorySituationsOptions) sq.SelectBuilder {
	return builder.getHistorySituationsIdsBase(options).
		Options("distinct on (situation_id, situation_instance_id)").
		OrderBy("situation_id", "situation_instance_id", "ts desc")
}

func (builder HistorySituationsBuilder) GetHistorySituationsIdsByStandardInterval(options GetHistorySituationsOptions, interval string) sq.SelectBuilder {
	return builder.getHistorySituationsIdsBase(options).
		Options("distinct on (situation_id, situation_instance_id, date_trunc('"+interval+"', ts))").
		OrderBy("situation_id", "situation_instance_id", "date_trunc('"+interval+"', ts)")
}

func (builder HistorySituationsBuilder) GetHistorySituationsIdsByCustomInterval(options GetHistorySituationsOptions, referenceDate time.Time, interval time.Duration) sq.SelectBuilder {
	intervalMin := fmt.Sprintf("%d", int64(interval.Minutes()))
	return builder.getHistorySituationsIdsBase(options).
		Options("distinct on (situation_id, situation_instance_id, CAST('2022-08-01' AS TIMESTAMP) + INTERVAL '1 second' * "+intervalMin+" * FLOOR(DATE_PART('epoch', ts- '2022-08-01')/"+intervalMin+"))").
		OrderBy("situation_id", "situation_instance_id", "CAST('2022-08-01' AS TIMESTAMP) + INTERVAL '1 second' * "+intervalMin+" * FLOOR(DATE_PART('epoch', ts- '2022-08-01')/"+intervalMin+")")
}

func (builder HistorySituationsBuilder) GetHistorySituationsDetails(subQueryIds string, subQueryIdsArgs []interface{}) sq.SelectBuilder {
	return builder.newStatement().
		Select("sh.*, s.name, si.name").
		From("situation_definition_v1 s").
		LeftJoin("situation_template_instances_v1 si on s.id = si.situation_id").
		InnerJoin("situation_history_v4 sh on (s.id = sh.situation_id and (sh.situation_instance_id = si.id OR sh.situation_instance_id = 0))").
		Where("sh.id = any ("+subQueryIds+")", subQueryIdsArgs...)
}
