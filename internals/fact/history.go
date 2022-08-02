package fact

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"go.uber.org/zap"
)

// SituationHistoryRecord represents a single and unique situation history entry
type SituationHistoryRecord struct {
	ID                 int64
	TS                 time.Time
	TemplateInstanceID int64
	Parameters         map[string]string
}

// HistoryRecord fact history record
type HistoryRecord struct {
	Fact               engine.Fact
	ID                 int64
	TS                 time.Time
	SituationID        int64
	TemplateInstanceID int64
	SituationInstances []SituationHistoryRecord
}

// PersistFactResult persists a fact result (related to a specific time) in postgresql
func PersistFactResult(factID int64, t time.Time, situationID int64, templateInstanceID int64, item *reader.Item, success bool) error {
	if postgres.DB() == nil {
		return errors.New("db Client is not initialized")
	}
	if item == nil {
		return errors.New("cannot persist nil item")
	}

	itemJSON, err := json.Marshal(*item)
	if err != nil {
		zap.L().Error("PersistFactResult.Marshal:", zap.Error(err))
		return err
	}

	query := `INSERT INTO fact_history_v1(id, ts, situation_id, situation_instance_id, result, success) VALUES (:id, :ts, :situation_id, :situation_instance_id,  :result, :success)`

	params := map[string]interface{}{
		"id":                    factID,
		"ts":                    t,
		"situation_id":          situationID,
		"situation_instance_id": templateInstanceID,
		"result":                string(itemJSON),
		"success":               success,
	}
	res, err := postgres.DB().NamedExec(query, params)
	if err != nil {
		return err
	}
	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i != 1 {
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// UpdateFactResult updates a fact result (related to a specific time) in postgresql
func UpdateFactResult(factID int64, t time.Time, situationID int64, templateInstanceID int64, item *reader.Item) error {
	if postgres.DB() == nil {
		return errors.New("db Client is not initialized")
	}
	if item == nil {
		return errors.New("cannot persist nil item")
	}

	itemJSON, err := json.Marshal(*item)
	if err != nil {
		zap.L().Error("PersistFactResult.Marshal:", zap.Error(err))
		return err
	}

	query := `UPDATE fact_history_v1 set result = :result WHERE id = :id AND ts = :ts AND situation_id = :situation_id AND situation_instance_id = :situation_instance_id`
	params := map[string]interface{}{
		"id":                    factID,
		"ts":                    t,
		"situation_id":          situationID,
		"situation_instance_id": templateInstanceID,
		"result":                string(itemJSON),
	}
	res, err := postgres.DB().NamedExec(query, params)
	if err != nil {
		return err
	}
	i, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if i != 1 {
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}

// GetFactResultFromHistory returns the fact result matching the input timestamp.
// It can be based on an exact timestamp or approximated to the closest calculated result
func GetFactResultFromHistory(factID int64, t time.Time, situationID int64, templateInstanceID int64, closest bool, notOlderThan time.Duration) (*reader.Item, time.Time, error) {
	if postgres.DB() == nil {
		return nil, time.Time{}, errors.New("db Client is not initialized")
	}

	params := map[string]interface{}{
		"fact_id":               factID,
		"ts":                    t,
		"situation_id":          situationID,
		"situation_instance_id": templateInstanceID,
	}

	var query string
	if !closest {
		query = "SELECT ts, result FROM fact_history_v1 WHERE success = true AND id = :fact_id AND ts = :ts AND (situation_id = 0 OR situation_id = :situation_id) AND (situation_instance_id = 0 OR situation_instance_id = :situation_instance_id)"
	} else {
		query = "SELECT ts, result FROM fact_history_v1 WHERE success = true AND id = :fact_id AND ts <= :ts AND (situation_id = 0 OR situation_id = :situation_id) AND (situation_instance_id = 0 OR situation_instance_id = :situation_instance_id)"
		if notOlderThan > 0 {
			query += " AND ts > :ts_min"
			params["ts_min"] = t.Add(-1 * notOlderThan)
		}
		query += " ORDER BY ts DESC LIMIT 1"
	}

	rows, err := postgres.DB().NamedQuery(query, params)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return nil, time.Time{}, rows.Err()
	}
	if rows.Next() {
		var result []byte
		var ts time.Time
		err = rows.Scan(&ts, &result)
		if err != nil {
			return nil, time.Time{}, err
		}

		var item reader.Item
		err = json.Unmarshal(result, &item)
		if err != nil {
			return nil, time.Time{}, err
		}
		return &item, ts, nil
	}
	return nil, time.Time{}, nil
}

//GetFactRangeFromHistory get the facts history within a date range
func GetFactRangeFromHistory(factID int64, situationID int64, templateInstanceID int64, tsFrom time.Time, tsTo time.Time) (map[time.Time]reader.Item, error) {
	items := make(map[time.Time]reader.Item, 0)

	if postgres.DB() == nil {
		return nil, errors.New("db Client is not initialized")
	}

	query := `SELECT ts, result 
		FROM fact_history_v1 
		WHERE success = true AND id = :factid 
		AND (situation_id = 0 OR situation_id = :situation_id) 
		AND (situation_instance_id = 0 OR situation_instance_id = :situation_instance_id)
		AND ts >= :tsFrom AND ts <= :tsTo ORDER BY ts`
	params := map[string]interface{}{
		"factid":                factID,
		"tsFrom":                tsFrom,
		"tsTo":                  tsTo,
		"situation_id":          situationID,
		"situation_instance_id": templateInstanceID,
	}
	rows, err := postgres.DB().NamedQuery(query, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	for rows.Next() {
		var result []byte
		var ts time.Time

		err = rows.Scan(&ts, &result)
		if err != nil {
			return nil, err
		}

		var item reader.Item

		err = json.Unmarshal(result, &item)
		if err != nil {
			return nil, err
		}

		items[ts] = item
	}

	return items, nil
}

//GetFactSituationInstances get the situation instances for fact instances between to and from
func GetFactSituationInstances(factIDs []int64, from time.Time, to time.Time, lastDailyValue bool) ([]HistoryRecord, error) {
	mapRecords := make(map[string]HistoryRecord, 0)

	if postgres.DB() == nil {
		return nil, errors.New("db Client is not initialized")
	}

	var factsFilterQuery string
	if lastDailyValue {
		factsFilterQuery = `SELECT DISTINCT ON (id, situation_id, situation_instance_id, interval) 
								id, ts, situation_id, situation_instance_id, FLOOR(DATE_PART('epoch', ts - $2)/86400) AS interval 
							FROM fact_history_v1 WHERE id = ANY ($1) AND ts >= $2 AND ts <= $3 
							ORDER by id, situation_id, situation_instance_id, interval, ts DESC`
	} else {
		factsFilterQuery = `SELECT id, ts, situation_id, situation_instance_id 
							FROM fact_history_v1 WHERE id = ANY ($1) AND ts >= $2 AND ts <= $3 
							ORDER by id, situation_id, situation_instance_id, ts DESC`
	}

	query := `SELECT fd.definition, f.id, f.ts, f.situation_id, f.situation_instance_id, s.id, s.ts, s.situation_instance_id, s.parameters 
				FROM (` + factsFilterQuery + `) as f
				INNER JOIN (SELECT id, ts, situation_instance_id, parameters, facts_ids FROM situation_history_v1 WHERE ts >= $2 AND ts <= $3) as s 
				ON (f.situation_id = s.id OR f.situation_id = 0) AND (f.situation_instance_id = s.situation_instance_id OR f.situation_instance_id = 0) AND 
					(s.facts_ids ->> CAST(f.id AS TEXT))::timestamptz = f.ts
				INNER JOIN fact_definition_v1 as fd ON fd.id = f.id;`

	rows, err := postgres.DB().Query(query, pq.Array(factIDs), from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	for rows.Next() {
		var rawFact string
		var factID int64
		var factTS time.Time
		var factSID int64
		var factSTemplateID int64
		var situationID int64
		var situationTS time.Time
		var situationTemplateID int64
		var rawSParams []byte

		err = rows.Scan(&rawFact, &factID, &factTS, &factSID, &factSTemplateID, &situationID, &situationTS, &situationTemplateID, &rawSParams)
		if err != nil {
			return nil, err
		}

		var parameters map[string]string
		err = json.Unmarshal(rawSParams, &parameters)
		if err != nil {
			return nil, err
		}

		key := fmt.Sprintf("%d-%s-%d-%d", factID, factTS, factSID, factSTemplateID)
		if record, ok := mapRecords[key]; ok {
			record.SituationInstances = append(record.SituationInstances, SituationHistoryRecord{
				ID:                 situationID,
				TS:                 situationTS,
				TemplateInstanceID: situationTemplateID,
				Parameters:         parameters,
			})
		} else {

			var fact engine.Fact
			err = json.Unmarshal([]byte(rawFact), &fact)
			if err != nil {
				return nil, err
			}
			fact.ID = factID

			mapRecords[key] = HistoryRecord{
				Fact:               fact,
				ID:                 factID,
				TS:                 factTS,
				SituationID:        factSID,
				TemplateInstanceID: factSTemplateID,
				SituationInstances: []SituationHistoryRecord{
					SituationHistoryRecord{
						ID:                 situationID,
						TS:                 situationTS,
						TemplateInstanceID: situationTemplateID,
						Parameters:         parameters,
					},
				},
			}
		}
	}

	records := make([]HistoryRecord, 0)
	for _, record := range mapRecords {
		records = append(records, record)
	}

	return records, nil
}
