package history

import (
	"github.com/jmoiron/sqlx"
)

type HistoryService struct {
	HistorySituationsQuerier     HistorySituationsQuerier
	HistorySituationFactsQuerier HistorySituationFactsQuerier
	HistoryFactsQuerier          HistoryFactsQuerier
}

func New(db *sqlx.DB) HistoryService {
	return HistoryService{
		HistorySituationsQuerier:     HistorySituationsQuerier{conn: db, Builder: HistorySituationsBuilder{}},
		HistorySituationFactsQuerier: HistorySituationFactsQuerier{conn: db, Builder: HistorySituationFactsBuilder{}},
		HistoryFactsQuerier:          HistoryFactsQuerier{conn: db, Builder: HistoryFactsBuilder{}},
	}
}
