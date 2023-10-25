package history

import (
	"time"

	sq "github.com/Masterminds/squirrel"
)

type HistoryFactsBuilder struct{}

func (builder HistoryFactsBuilder) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

func (builder HistoryFactsBuilder) GetHistoryFactLast(situationID int64, instanceID int64, factID int64) sq.SelectBuilder {
	return builder.newStatement().
		Select("fh.*, f.name").
		From("fact_history_v5 fh").
		InnerJoin("fact_definition_v1 f on fh.fact_id = f.id").
		OrderBy("fh.ts desc").
		Limit(1).
		Where(sq.Eq{"fh.situation_id": situationID}).
		Where(sq.Eq{"fh.situation_instance_id": instanceID}).
		Where(sq.Eq{"fh.fact_id": factID})
}

func (builder HistoryFactsBuilder) GetHistoryFacts(historyFactsIds []int64) sq.SelectBuilder {
	return builder.newStatement().
		Select("fh.*, f.name").
		From("fact_history_v5 fh").
		InnerJoin("fact_definition_v1 f on fh.fact_id = f.id").
		Where(sq.Eq{"fh.id": historyFactsIds})
}

func (builder HistoryFactsBuilder) Insert(history HistoryFactsV4, resultJSON []byte) sq.InsertBuilder {
	return builder.newStatement().
		Insert("fact_history_v5").
		Columns("id", "fact_id", "situation_id", "situation_instance_id", "ts", "result").
		Values(sq.Expr("DEFAULT"), history.FactID, history.SituationID, history.SituationInstanceID, history.Ts, resultJSON).
		Suffix("RETURNING id")
}

func (builder HistoryFactsBuilder) Update(id int64, resultJSON []byte) sq.UpdateBuilder {
	return builder.newStatement().
		Update("fact_history_v5").
		Where(sq.Eq{"id": id}).
		Set("result", resultJSON)
}

func (builder HistoryFactsBuilder) DeleteOrphans() sq.DeleteBuilder {
	return builder.newStatement().
		Delete("fact_history_v5").
		Where(
			builder.newStatement().
				Select("1").
				From("situation_fact_history_v5").
				Where("fact_history_v5.id = situation_fact_history_v5.fact_history_id").
				Prefix("NOT EXISTS (").
				Suffix(")"),
		)
}

func (builder HistoryFactsBuilder) Delete(ID int64) sq.DeleteBuilder {
	return builder.newStatement().
		Delete("fact_history_v5").
		Where(sq.Eq{"id": ID})
}

func (builder HistoryFactsBuilder) GetTodaysFactResultByParameters(param ParamGetFactHistory) sq.SelectBuilder {
	todayStart, tomorrowStart := getTodayTimeRange()

	return builder.newStatement().
		Select("result, ts").
		From("fact_history_v5").
		Where(sq.Eq{"fact_id": param.FactID}).
		Where(sq.Eq{"situation_id": param.SituationID}).
		Where(sq.Eq{"situation_instance_id": param.SituationInstanceID}).
		Where(sq.Expr("ts >= ?::timestamptz", todayStart)).
		Where(sq.Expr("ts < ?::timestamptz", tomorrowStart))
}

