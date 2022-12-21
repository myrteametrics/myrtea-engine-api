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
	ParameterFilters    map[string]string
	FromTS              time.Time
	ToTS                time.Time
}

func (builder HistorySituationsBuilder) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

func (builder HistorySituationsBuilder) GetHistorySituationsIdsBase(options GetHistorySituationsOptions) sq.SelectBuilder {
	q := builder.newStatement().
		Select("id").
		From("situation_history_v5")
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
	for k, v := range options.ParameterFilters {
		q = q.Where(sq.Eq{"parameters->>'" + k + "'": v})
	}
	return q
}

func (builder HistorySituationsBuilder) GetHistorySituationsIdsLast(options GetHistorySituationsOptions) sq.SelectBuilder {
	return builder.GetHistorySituationsIdsBase(options).
		Options("distinct on (situation_id, situation_instance_id)").
		OrderBy("situation_id", "situation_instance_id", "ts desc")
}

func (builder HistorySituationsBuilder) GetHistorySituationsIdsByStandardInterval(options GetHistorySituationsOptions, interval string) sq.SelectBuilder {
	return builder.GetHistorySituationsIdsBase(options).
		Options("distinct on (situation_id, situation_instance_id, date_trunc('"+interval+"', ts))").
		OrderBy("situation_id", "situation_instance_id", "date_trunc('"+interval+"', ts) desc, ts desc")
}

func (builder HistorySituationsBuilder) GetHistorySituationsIdsByCustomInterval(options GetHistorySituationsOptions, interval time.Duration, referenceDate time.Time) sq.SelectBuilder {
	intervalSeconds := fmt.Sprintf("%d", int64(interval.Seconds()))
	referenceDateStr := referenceDate.Format("2006-01-02T15:04:05Z07:00")
	return builder.GetHistorySituationsIdsBase(options).
		Options("distinct on (situation_id, situation_instance_id, CAST('"+referenceDateStr+"' AS TIMESTAMPTZ) + INTERVAL '1 second' * "+intervalSeconds+" * FLOOR(DATE_PART('epoch', ts- '"+referenceDateStr+"')/"+intervalSeconds+"))").
		OrderBy("situation_id", "situation_instance_id", "CAST('"+referenceDateStr+"' AS TIMESTAMPTZ) + INTERVAL '1 second' * "+intervalSeconds+" * FLOOR(DATE_PART('epoch', ts- '"+referenceDateStr+"')/"+intervalSeconds+") desc, ts desc")
}

func (builder HistorySituationsBuilder) GetHistorySituationsDetails(subQueryIds string, subQueryIdsArgs []interface{}) sq.SelectBuilder {
	return builder.newStatement().
		Select("sh.*, s.name, si.name, s.definition::json->'parameters' as situationParameters").
		From("situation_definition_v1 s").
		LeftJoin("situation_template_instances_v1 si on s.id = si.situation_id").
		InnerJoin("situation_history_v5 sh on (s.id = sh.situation_id and (sh.situation_instance_id = si.id OR sh.situation_instance_id = 0))").
		Where("sh.id = any ("+subQueryIds+")", subQueryIdsArgs...)
}

func (builder HistorySituationsBuilder) Insert(history HistorySituationsV4, parametersJSON []byte, expressionFactsJSON []byte, metadatasJSON []byte) sq.InsertBuilder {
	return builder.newStatement().Insert("situation_history_v5").
		Columns("id", "situation_id", "situation_instance_id", "ts", "parameters", "expression_facts", "metadatas").
		Values(sq.Expr("DEFAULT"), history.SituationID, history.SituationInstanceID, history.Ts, parametersJSON, expressionFactsJSON, metadatasJSON).
		Suffix("RETURNING id")
}
