package history

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
)

type HistorySituationsBuilder struct{}

type GetHistorySituationsOptions struct {
	SituationID          int64
	SituationInstanceIDs []int64
	ParameterFilters     map[string]interface{}
	DeleteBeforeTs       time.Time
	FromTS               time.Time
	ToTS                 time.Time
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

	if len(options.SituationInstanceIDs) > 0 {
		q = q.Where(sq.Eq{"situation_instance_id": options.SituationInstanceIDs})
	}

	if !options.FromTS.IsZero() {
		q = q.Where(sq.GtOrEq{"ts": options.FromTS})
	}

	if !options.ToTS.IsZero() {
		q = q.Where(sq.Lt{"ts": options.ToTS})
	}

	if !options.DeleteBeforeTs.IsZero() {
		q = q.Where(sq.LtOrEq{"ts": options.DeleteBeforeTs})
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
		Select("sh.*, s.name, coalesce(si.name, ''), c.id, c.name, c.description, c.timezone").
		From("situation_definition_v1 s").
		LeftJoin("situation_template_instances_v1 si on s.id = si.situation_id").
		LeftJoin("calendar_v1 c on c.id = COALESCE(si.calendar_id, s.calendar_id)").
		InnerJoin("situation_history_v5 sh on (s.id = sh.situation_id and (sh.situation_instance_id = si.id OR sh.situation_instance_id = 0))").
		Where("sh.id = any ("+subQueryIds+")", subQueryIdsArgs...)
}

func (builder HistorySituationsBuilder) Insert(history HistorySituationsV4, parametersJSON []byte, expressionFactsJSON []byte, metadatasJSON []byte) sq.InsertBuilder {
	return builder.newStatement().Insert("situation_history_v5").
		Columns("id", "situation_id", "situation_instance_id", "ts", "parameters", "expression_facts", "metadatas").
		Values(sq.Expr("DEFAULT"), history.SituationID, history.SituationInstanceID, history.Ts, parametersJSON, expressionFactsJSON, metadatasJSON).
		Suffix("RETURNING id")
}

func (builder HistorySituationsBuilder) Update(id int64, parametersJSON []byte, expressionFactsJSON []byte, metadatasJSON []byte) sq.UpdateBuilder {
	return builder.newStatement().
		Update("situation_history_v5").
		Where(sq.Eq{"id": id}).
		Set("parameters", parametersJSON).
		Set("expression_facts", expressionFactsJSON).
		Set("metadatas", metadatasJSON)
}

func (builder HistorySituationsBuilder) DeleteOrphans() sq.DeleteBuilder {
	return builder.newStatement().
		Delete("situation_history_v5").
		Where(
			builder.newStatement().
				Select("1").
				From("situation_fact_history_v5").
				Where("situation_history_v5.id = situation_fact_history_v5.situation_history_id").
				Prefix("NOT EXISTS (").
				Suffix(")"),
		)
}

func (builder HistorySituationsBuilder) GetLatestHistorySituation(situationID int64, situationInstanceID int64) sq.SelectBuilder {
	startOfLastMonth := getStartDate30DaysAgo()

	return builder.newStatement().
		Select("ts", "metadatas").
		From("situation_history_v5").
		Where(sq.Eq{"situation_id": situationID}).
		Where(sq.Eq{"situation_instance_id": situationInstanceID}).
		Where(sq.Expr("ts >= ?::timestamptz", startOfLastMonth)).
		OrderBy("ts DESC").
		Limit(1)
}

func (builder HistorySituationsBuilder) GetTodaysFactExprResultByParameters(param ParamGetFactExprHistory) sq.SelectBuilder {
	todayStart, tomorrowStart := getTodayTimeRange()

	return builder.newStatement().
		Select("expression_facts, ts").
		From("situation_history_v5").
		Where(sq.Eq{"situation_id": param.SituationID}).
		Where(sq.Eq{"situation_instance_id": param.SituationInstanceID}).
		Where(sq.Expr("ts >= ?::timestamptz", todayStart)).
		Where(sq.Expr("ts < ?::timestamptz", tomorrowStart))
}

func (builder HistorySituationsBuilder) GetFactExprResultByDate(param ParamGetFactExprHistoryByDate) sq.SelectBuilder {
	return builder.newStatement().
		Select("expression_facts, ts").
		From("situation_history_v5").
		Where(sq.Eq{"situation_id": param.SituationID}).
		Where(sq.Eq{"situation_instance_id": param.SituationInstanceID}).
		Where(sq.Expr("ts >= ?::timestamptz", param.StartDate)).
		Where(sq.Expr("ts < ?::timestamptz", param.EndDate))
}
