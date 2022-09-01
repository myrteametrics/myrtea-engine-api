package history

import (
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

func (builder HistoryFactsBuilder) Update(history HistoryFactsV4) sq.UpdateBuilder {
	return builder.newStatement().
		Update("fact_history_v5").
		Where("id", history.ID).
		Set("fact_id", history.FactID).
		Set("situation_id", history.SituationID).
		Set("situation_instance_id", history.SituationInstanceID).
		Set("ts", history.Ts).
		Set("result", history.Result)
}
