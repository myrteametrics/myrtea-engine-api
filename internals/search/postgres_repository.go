package search

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"go.uber.org/zap"
)

type mapResult map[time.Time][]SituationHistoryRecord

type RawSituationRecord struct {
	RawFactIDs         []byte
	RawParams          []byte
	RawExpressionFacts []byte
	RawMetadatas       []byte
	TS                 time.Time
	InstanceID         int64
	InstanceName       interface{}
}

// PostgresRepository is a repository containing the situation definition based on a PSQL database and
//implementing the repository interface
type PostgresRepository struct {
	conn *sqlx.DB
}

// NewPostgresRepository returns a new instance of PostgresRepository
func NewPostgresRepository(dbClient *sqlx.DB) Repository {
	r := PostgresRepository{
		conn: dbClient,
	}
	var ifm Repository = &r
	return ifm
}

// GetSituationHistoryRecords returns the situations records at a specific time
func (r *PostgresRepository) GetSituationHistoryRecords(s situation.Situation, templateInstanceID int64, t time.Time, start time.Time, end time.Time,
	factSource interface{}, expressionFactsSource interface{}, metaDataSource interface{}, parametersSource interface{},
	downSampling DownSampling) (QueryResult, error) {

	var query string
	if downSampling.GranularitySpecial == "year" || downSampling.GranularitySpecial == "quarter" || downSampling.GranularitySpecial == "month" || downSampling.GranularitySpecial == "day" ||
		downSampling.GranularitySpecial == "hour" || downSampling.GranularitySpecial == "minute" || downSampling.GranularitySpecial == "second" {

		if downSampling.Operation == "latest" {
			var order = "DESC"
			if downSampling.Operation == "first" {
				order = "ASC"
			}

			query = `SELECT DISTINCT ON (situation_instance_id, name, interval) situation_instance_id, name, ts, facts_ids, expression_facts, parameters, metadatas
				FROM (
					SELECT situation_history_v1.situation_instance_id, situation_template_instances_v1.name, ts,
					date_trunc('` + downSampling.GranularitySpecial + `', ts) AS interval,
					situation_history_v1.facts_ids, situation_history_v1.expression_facts, situation_history_v1.parameters, situation_history_v1.metadatas
					FROM situation_history_v1 LEFT JOIN situation_template_instances_v1 ON situation_history_v1.situation_instance_id = situation_template_instances_v1.id
					WHERE situation_history_v1.id = :situation_id AND (:situation_instance_id = 0 OR situation_history_v1.situation_instance_id = :situation_instance_id)
					AND ts >= :tsFrom AND ts <= :tsTo
				) AS t
				ORDER BY
					situation_instance_id ` + order + `,
					name ` + order + `,
					interval ` + order + `,
					ts ` + order
		}

	} else if downSampling.Granularity != 0 {
		if downSampling.Operation == "first" || downSampling.Operation == "latest" {

			var order = "DESC"
			if downSampling.Operation == "first" {
				order = "ASC"
			}

			query = `SELECT DISTINCT ON (situation_instance_id, name, interval) situation_instance_id, name, ts, facts_ids, expression_facts, parameters, metadatas
				FROM (
					SELECT situation_history_v1.situation_instance_id, situation_template_instances_v1.name, ts,
					FLOOR(DATE_PART('epoch', ts- :tsFrom)/:granularity) AS interval,
					situation_history_v1.facts_ids, situation_history_v1.expression_facts, situation_history_v1.parameters, situation_history_v1.metadatas
					FROM situation_history_v1 LEFT JOIN situation_template_instances_v1 ON situation_history_v1.situation_instance_id = situation_template_instances_v1.id
					WHERE situation_history_v1.id = :situation_id AND (:situation_instance_id = 0 OR situation_history_v1.situation_instance_id = :situation_instance_id)
					AND ts >= :tsFrom AND ts <= :tsTo
				) AS t
				ORDER BY
					situation_instance_id ` + order + `,
					name ` + order + `,
					interval ` + order + `,
					ts ` + order
		} else {
			query = `SELECT situation_history_v1.situation_instance_id, situation_template_instances_v1.name,
				CAST(:tsFrom AS TIMESTAMP) + INTERVAL '1 second' * :granularity * FLOOR(DATE_PART('epoch', ts- :tsFrom)/:granularity) AS timestamp,
				JSON_AGG(situation_history_v1.facts_ids), JSON_AGG(situation_history_v1.expression_facts), JSON_AGG(situation_history_v1.parameters), JSON_AGG(situation_history_v1.metadatas)
				FROM situation_history_v1 LEFT JOIN situation_template_instances_v1 ON situation_history_v1.situation_instance_id = situation_template_instances_v1.id
				WHERE situation_history_v1.id = :situation_id AND (:situation_instance_id = 0 OR situation_history_v1.situation_instance_id = :situation_instance_id)
				AND ts >= :tsFrom AND ts <= :tsTo
				GROUP BY (situation_history_v1.situation_instance_id, situation_template_instances_v1.name, timestamp)
				ORDER BY timestamp ASC`
		}
	} else {
		if !t.IsZero() {
			query = `SELECT DISTINCT ON (situation_history_v1.situation_instance_id)
					situation_history_v1.situation_instance_id, situation_template_instances_v1.name, situation_history_v1.ts,
					situation_history_v1.facts_ids, situation_history_v1.expression_facts, situation_history_v1.parameters, situation_history_v1.metadatas
					FROM situation_history_v1 LEFT JOIN situation_template_instances_v1 ON situation_history_v1.situation_instance_id = situation_template_instances_v1.id
					WHERE situation_history_v1.id = :situation_id AND (:situation_instance_id = 0 OR situation_history_v1.situation_instance_id = :situation_instance_id)
					AND ts <= :ts
					ORDER BY situation_history_v1.situation_instance_id, situation_history_v1.ts DESC`

		} else {
			query = `SELECT situation_instance_id, situation_template_instances_v1.name, situation_history_v1.ts, situation_history_v1.facts_ids,
					situation_history_v1.expression_facts, situation_history_v1.parameters, situation_history_v1.metadatas
					FROM situation_history_v1 LEFT JOIN situation_template_instances_v1 ON situation_history_v1.situation_instance_id = situation_template_instances_v1.id
					WHERE situation_history_v1.id = :situation_id AND (:situation_instance_id = 0 OR situation_history_v1.situation_instance_id = :situation_instance_id)
					AND ts >= :tsFrom AND ts <= :tsTo
					ORDER BY situation_history_v1.ts ASC`
		}
	}

	params := map[string]interface{}{
		"situation_id":          s.ID,
		"situation_instance_id": templateInstanceID,
		"ts":                    t.UTC(),
		"tsFrom":                start.UTC(),
		"tsTo":                  end.UTC(),
		"granularity":           int(downSampling.Granularity.Seconds()),
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		zap.L().Error("Cannot retrieve situation history data", zap.Int64("situationID", s.ID), zap.Int64("SituationInstanceID", templateInstanceID), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	result := make(QueryResult, 0)
	mapResult := make(mapResult, 0)

	rawSituationRecords := make([]RawSituationRecord, 0)
	for rows.Next() {
		var rawFactIDs []byte
		var rawParams []byte
		var rawExpressionFacts []byte
		var rawMetadatas []byte
		var ts time.Time
		var instanceID int64
		var instanceName interface{}
		err = rows.Scan(&instanceID, &instanceName, &ts, &rawFactIDs, &rawExpressionFacts, &rawParams, &rawMetadatas)
		if err != nil {
			zap.L().Error("Couldn't scan the retrieved data:", zap.Int64("situationID", s.ID), zap.Int64("SituationInstanceID", templateInstanceID), zap.Error(err))
			zap.L().Info("q", zap.String("query", query))
			return nil, err
		}
		rawSituationRecords = append(rawSituationRecords, RawSituationRecord{
			RawFactIDs:         rawFactIDs,
			RawParams:          rawParams,
			RawExpressionFacts: rawExpressionFacts,
			RawMetadatas:       rawMetadatas,
			TS:                 ts,
			InstanceID:         instanceID,
			InstanceName:       instanceName,
		})
	}

	for _, rawSituationRecord := range rawSituationRecords {
		situationInstanceName := ""
		if rawSituationRecord.InstanceName != nil {
			situationInstanceName = rawSituationRecord.InstanceName.(string)
		}

		record := SituationHistoryRecord{
			SituationID:           s.ID,
			SituationName:         s.Name,
			SituationInstanceID:   rawSituationRecord.InstanceID,
			SituationInstanceName: situationInstanceName,
			MetaData:              nil,
			DateTime:              rawSituationRecord.TS.In(start.Location()),
		}

		facts, err := r.getFactHistoryRecords(rawSituationRecord.RawFactIDs, s.ID, rawSituationRecord.InstanceID, start.Location(), factSource, downSampling.Operation)
		if err != nil {
			zap.L().Error("Error getting situation instance history facts", zap.Int64("situationID", s.ID), zap.Int64("SituationInstanceID", templateInstanceID), zap.Time("timestamp", ts), zap.Error(err))
		}
		if facts != nil && len(facts) > 0 {
			record.Facts = facts
		}

		var expressionFacts = make(map[string]interface{}, 0)
		err = extractExpressionFacts(rawSituationRecord.RawExpressionFacts, expressionFacts, expressionFactsSource, downSampling.Operation)
		if err != nil {
			zap.L().Error("Error unmarshalling situation instance history ExpressionFacts", zap.Int64("situationID", s.ID), zap.Int64("SituationInstanceID", templateInstanceID), zap.Time("timestamp", ts), zap.Error(err))
		}
		if len(expressionFacts) > 0 {
			record.ExpressionFacts = expressionFacts
		}

		var params = make(map[string]interface{}, 0)
		err = extractParameters(rawSituationRecord.RawParams, params, parametersSource, downSampling.Operation)
		if err != nil {
			zap.L().Error("Error unmarshalling situation instance history parameters", zap.Int64("situationID", s.ID), zap.Int64("SituationInstanceID", templateInstanceID), zap.Time("timestamp", ts), zap.Error(err))
		}
		if len(params) > 0 {
			record.Parameters = params
		}

		var metaData = make(map[string]interface{}, 0)
		err = extractMetaData(rawSituationRecord.RawMetadatas, metaData, metaDataSource, downSampling.Operation)
		if err != nil {
			zap.L().Error("Error extracting situation instance history metadatas", zap.Int64("situationID", s.ID), zap.Int64("SituationInstanceID", templateInstanceID), zap.Time("timestamp", ts), zap.Error(err))
		}
		if len(metaData) > 0 {
			record.MetaData = metaData
		}

		if _, ok := mapResult[record.DateTime]; ok {
			mapResult[record.DateTime] = append(mapResult[record.DateTime], record)
		} else {
			mapResult[record.DateTime] = []SituationHistoryRecord{record}
		}
	}

	for dt, records := range mapResult {
		result = append(result, SituationHistoryRecords{
			DateTime:   dt,
			Situations: records,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].DateTime.Before(result[j].DateTime)
	})

	return result, nil
}

func (r *PostgresRepository) getFactHistoryRecords(rawFactIDs []byte, situationID int64, templateInstanceID int64, timeLocation *time.Location, factSource interface{}, downSamplingOperation string) ([]FactHistoryRecord, error) {
	if rawFactIDs == nil {
		return nil, nil
	}
	params := map[string]interface{}{
		"situation_id":          situationID,
		"situation_instance_id": templateInstanceID,
	}
	var filterCondition string
	switch value := factSource.(type) {
	case bool:
		if !value {
			return nil, nil
		}
	case string:
		filterCondition = "AND fact_definition_v1.name = :fact_name "
		params["fact_name"] = value
	case []string:
		filterCondition = "AND ("
		for i, name := range value {
			if i > 0 {
				filterCondition = filterCondition + " OR "
			}
			filterCondition = filterCondition + fmt.Sprintf("fact_definition_v1.name = :fact_name_%d", i)
			params[fmt.Sprintf("fact_name_%d", i)] = name
		}
		filterCondition = filterCondition + ") "
	}

	var conditions string
	var query string
	if downSamplingOperation == "" || downSamplingOperation == "first" || downSamplingOperation == "latest" {

		var factIDs map[int64]*time.Time
		err := json.Unmarshal(rawFactIDs, &factIDs)
		if err != nil {
			zap.L().Error("Error unmarshalling situation instance history fact ids", zap.Error(err))
		}

		for factID, ts := range factIDs {
			if ts == nil {
				continue
			}
			if conditions != "" {
				conditions = conditions + " OR "
			}
			conditions = conditions + fmt.Sprintf("(fact_history_v1.id = :fact_id_%d AND fact_history_v1.ts = :ts_%d)", factID, factID)
			params[fmt.Sprintf("fact_id_%d", factID)] = factID
			params[fmt.Sprintf("ts_%d", factID)] = *ts
		}

		query = `SELECT fact_definition_v1.id, fact_definition_v1.name, ts, fact_history_v1.result
					FROM fact_history_v1 JOIN fact_definition_v1 ON fact_history_v1.id = fact_definition_v1.id
					WHERE fact_history_v1.success = true AND (` + conditions + `) ` + filterCondition + `
						AND (situation_id = 0 OR situation_id = :situation_id) AND (situation_instance_id = 0 OR situation_instance_id = :situation_instance_id)`

	} else {
		var factIDsList []map[int64]*time.Time
		err := json.Unmarshal(rawFactIDs, &factIDsList)
		if err != nil {
			zap.L().Error("Error unmarshalling situation instance history fact ids", zap.Error(err))
		}

		factIDFrom := make(map[int64]*time.Time, 0)
		factIDTo := make(map[int64]*time.Time, 0)

		for _, factIDs := range factIDsList {
			for key, value := range factIDs {
				if value == nil {
					continue
				}
				if val, ok := factIDFrom[key]; ok {
					if value.Before(*val) {
						factIDFrom[key] = value
					}
				} else {
					factIDFrom[key] = value
				}
				if val, ok := factIDTo[key]; ok {
					if value.After(*val) {
						factIDTo[key] = value
					}
				} else {
					factIDTo[key] = value
				}
			}
		}

		for factID := range factIDFrom {
			if conditions != "" {
				conditions = conditions + " OR "
			}
			conditions = conditions + fmt.Sprintf("(fact_history_v1.id = :fact_id_%d AND fact_history_v1.ts >= :tsFrom_%d AND fact_history_v1.ts <= :tsTo_%d)", factID, factID, factID)
			params[fmt.Sprintf("fact_id_%d", factID)] = factID
			params[fmt.Sprintf("tsFrom_%d", factID)] = factIDFrom[factID]
			params[fmt.Sprintf("tsTo_%d", factID)] = factIDTo[factID]
		}

		query = `SELECT fact_definition_v1.id, fact_definition_v1.name, MIN(ts), JSON_AGG(fact_history_v1.result)
					FROM fact_history_v1 JOIN fact_definition_v1 ON fact_history_v1.id = fact_definition_v1.id
					WHERE fact_history_v1.success = true AND (` + conditions + `) ` + filterCondition + `
						AND (situation_id = 0 OR situation_id = :situation_id) AND (situation_instance_id = 0 OR situation_instance_id = :situation_instance_id)
					GROUP BY (fact_definition_v1.id, fact_definition_v1.name)`
	}

	rows, err := r.conn.NamedQuery(query, params)
	if err != nil {
		zap.L().Error("Cannot retrieve facts history data",
			zap.Int64("situationID", situationID),
			zap.Int64("SituationInstanceID", templateInstanceID),
			zap.Error(err),
		)
		return nil, err
	}
	defer rows.Close()

	facts := make([]FactHistoryRecord, 0)
	for rows.Next() {
		var factID int64
		var factName string
		var ts time.Time
		var result []byte
		err = rows.Scan(&factID, &factName, &ts, &result)
		if err != nil {
			zap.L().Error("Couldn't scan the fact history retrieved data:", zap.Error(err))
			continue
		}

		fact := FactHistoryRecord{
			FactID:   factID,
			FactName: factName,
			DateTime: ts.In(timeLocation),
		}

		err = extractFactHistoryRecordValues(result, &fact, downSamplingOperation)
		if err != nil {
			zap.L().Error("Error extracting situation instance facts history", zap.Error(err))
		} else {
			facts = append(facts, fact)
		}
	}

	return facts, nil
}
