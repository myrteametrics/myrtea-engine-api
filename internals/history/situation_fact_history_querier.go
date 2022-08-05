package history

import (
	sq "github.com/Masterminds/squirrel"
)

type HistorySituationFactsBuilder struct{}

func (builder HistorySituationFactsBuilder) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

func (builder HistorySituationFactsBuilder) GetHistorySituationFacts(historySituationsIds []int64) sq.SelectBuilder {
	return builder.newStatement().
		Select("*").
		From("situation_fact_history_v4").
		Where(sq.Eq{"situation_history_id": historySituationsIds})
}
