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
		From("situation_fact_history_v5").
		Where(sq.Eq{"situation_history_id": historySituationsIds})
}

func (builder HistorySituationFactsBuilder) InsertBulk(historySituationFacts []HistorySituationFactsV4) sq.InsertBuilder {
	b := builder.newStatement().
		Insert("situation_fact_history_v5").
		Columns("situation_history_id", "fact_history_id", "fact_id")
	for _, hishistorySituationFact := range historySituationFacts {
		b = b.Values(hishistorySituationFact.HistorySituationID, hishistorySituationFact.HistoryFactID, hishistorySituationFact.FactID)
	}

	return b
}

func (builder HistorySituationFactsBuilder) DeleteHistoryFrom(situationHistoryQueryBuilder sq.SelectBuilder) sq.DeleteBuilder {
	return builder.newStatement().
		Delete("situation_fact_history_v5").
		Where(
			situationHistoryQueryBuilder.
				Prefix("situation_history_id IN (").
				Suffix(")"),
		)
}
