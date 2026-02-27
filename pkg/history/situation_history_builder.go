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
	// IncludeCalendarStatus, if true, enriches each record with:
	//   - IsNowOutsideCalendar: real-time check at retrieval time (time.Now())
	//   - WereRulesOutsideCalendar: historical check at the record's own timestamp
	IncludeCalendarStatus bool
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

// GetHistorySituationsDetails builds the detail query for situation history records.
// When withRuleCalendars is true, it joins situation_rules_v1 and rules_v1 to collect
// rule calendar IDs via array_agg (needed for EnrichRuleCalendarStatus).
// When false, the joins and aggregation are skipped for better performance.
//
// Note: json columns (parameters, metadatas) are intentionally excluded from GROUP BY.
// PostgreSQL detects that sh.id (PRIMARY KEY) functionally determines all other sh.* columns.
func (builder HistorySituationsBuilder) GetHistorySituationsDetails(subQueryIds string, subQueryIdsArgs []interface{}, withRuleCalendars bool) sq.SelectBuilder {
	base := builder.newStatement().
		Select("sh.id, sh.situation_id, sh.situation_instance_id, sh.ts, sh.parameters, sh.expression_facts, sh.metadatas, s.name, coalesce(si.name, ''), c.id, c.name, c.description, c.timezone").
		From("situation_definition_v1 s").
		LeftJoin("situation_template_instances_v1 si on s.id = si.situation_id").
		LeftJoin("calendar_v1 c on c.id = COALESCE(si.calendar_id, s.calendar_id)").
		InnerJoin("situation_history_v5 sh on (s.id = sh.situation_id and (sh.situation_instance_id = si.id OR sh.situation_instance_id = 0))").
		Where("sh.id = any ("+subQueryIds+")", subQueryIdsArgs...)

	if withRuleCalendars {
		return base.
			Columns("array_agg(DISTINCT r.calendar_id) FILTER (WHERE r.calendar_id IS NOT NULL)").
			LeftJoin("situation_rules_v1 sr on sr.situation_id = s.id").
			LeftJoin("rules_v1 r on r.id = sr.rule_id").
			GroupBy("sh.id, sh.situation_id, sh.situation_instance_id, sh.ts, s.name, si.name, c.id, c.name, c.description, c.timezone")
	}

	return base.Columns("ARRAY[]::bigint[]")
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
