package history

import (
	sq "github.com/Masterminds/squirrel"
)

type HistoryFactsBuilder struct{}

func (builder HistoryFactsBuilder) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

func (builder HistoryFactsBuilder) GetHistoryFacts(historyFactsIds []int64) sq.SelectBuilder {
	return builder.newStatement().
		Select("fh.*, f.name").
		From("fact_history_v4 fh").
		InnerJoin("fact_definition_v1 f on fh.fact_id = f.id").
		Where(sq.Eq{"fh.id": historyFactsIds})
}
