package scheduler

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/metadata"
	situation2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/situation"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/evaluator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/rule"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tasker"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/history"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"github.com/myrteametrics/myrtea-sdk/v5/ruleeng"
	"go.uber.org/zap"
)

// FactRecalculationJob represent a scheduler job instance which process a group of facts, and persist the result in postgresql
// It also generate situations, persists them and notify the rule engine to evaluate them
type FactRecalculationJob struct {
	FactIds        []int64 `json:"factIds"`
	From           string  `json:"from,omitempty"`
	To             string  `json:"to,omitempty"`
	LastDailyValue bool    `json:"lastDailyValue,omitempty"`
	Debug          bool    `json:"debug"`
	ScheduleID     int64   `json:"-"`
}

// ResolveFromAndTo resolves the expressions in parameters From and To
func (job *FactRecalculationJob) ResolveFromAndTo(t time.Time) (time.Time, time.Time, error) {

	var from time.Time
	var to time.Time

	if job.From == "" && job.To == "" {
		return from, to, nil
	}
	if job.From == "" || job.To == "" {
		return from, to, errors.New("missing From or To Parameter")
	}

	variables := expression.GetDateKeywords(t)
	result, err := expression.Process(expression.LangEval, job.From, variables)
	if err != nil {
		zap.L().Error("Error processing From expression in fact calculation job", zap.Error(err))
		return from, to, err
	}
	from, err = time.ParseInLocation(timeLayout, result.(string), time.UTC)
	if err != nil {
		zap.L().Error("Error parsing From expression result as datetime in fact calculation job", zap.Error(err))
		return from, to, err
	}

	result, err = expression.Process(expression.LangEval, job.To, variables)
	if err != nil {
		zap.L().Error("Error processing To expression in fact calculation job", zap.Error(err))
		return from, to, err
	}
	to, err = time.ParseInLocation(timeLayout, result.(string), time.UTC)
	if err != nil {
		zap.L().Error("Error parsing To expression result as datetime in fact calculation job", zap.Error(err))
		return from, to, err
	}

	if job.LastDailyValue {
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, to.Location())
	}

	return from, to, nil
}

func (job FactRecalculationJob) Run() {

	if S().ExistingRunningJob(job.ScheduleID) {
		zap.L().Info("Skipping FactScheduleJob because last execution is still running", zap.Int64s("ids", job.FactIds))
		return
	}
	S().AddRunningJob(job.ScheduleID)

	zap.L().Info("Fact calculation job started", zap.Int64s("ids", job.FactIds))

	t := time.Now().Truncate(1 * time.Second).UTC()
	fromTS, toTS, err := job.ResolveFromAndTo(t)
	if err != nil {
		return
	}

	localRuleEngine, err := evaluator.BuildLocalRuleEngine("standart")
	if err != nil {
		zap.L().Error("BuildLocalRuleEngine", zap.Error(err))
		return
	}

	facts, err := fact.R().GetAllByIDs(job.FactIds)
	if err != nil {
		zap.L().Error("fact GetAllByIDs", zap.Error(err), zap.Int64s("ids", job.FactIds))
	}

	situations := make(map[int64]situation2.Situation)
	for _, factID := range job.FactIds {
		ss, _ := situation2.R().GetSituationsByFactID(factID, true)
		for _, s := range ss {
			situations[s.ID] = s
		}
	}

	for _, s := range situations {
		situationHistory, err := history.S().GetHistorySituationsIdsByStandardInterval(history.GetHistorySituationsOptions{
			SituationID: s.ID, SituationInstanceIDs: []int64{}, FromTS: fromTS, ToTS: toTS,
		}, "day")
		if err != nil {
			zap.L().Error("history GetHistorySituationsIdsByStandardInterval", zap.Error(err))
		}

		historyFacts, mapSituationFact, mapFactSituation, err := job.FetchRecalculationData(situationHistory)
		if err != nil {
			continue
		}

		mapFacts := make(map[int64]history.HistoryFactsV4)
		for _, historyFact := range historyFacts {
			mapFacts[historyFact.ID] = historyFact
		}

		mapSituations := make(map[int64]history.HistorySituationsV4)
		for _, historySituation := range situationHistory {
			mapSituations[historySituation.ID] = historySituation
		}

		// Fact history recalculation + update in database
		newFactHistory, err := job.RecalculateAndUpdateFacts(job.FactIds, facts, mapFactSituation, mapSituations, historyFacts)
		if err != nil {
			continue
		}

		err = job.RecalculateAndUpdateSituations(localRuleEngine, s, mapSituationFact, situationHistory, newFactHistory)
		if err != nil {
			continue
		}

	}

	zap.L().Info("FactScheduleJob Ended", zap.Int64s("ids", job.FactIds))

	S().RemoveRunningJob(job.ScheduleID)
}

func (job FactRecalculationJob) FetchRecalculationData(historySituations []history.HistorySituationsV4) ([]history.HistoryFactsV4, map[int64][]int64, map[int64]int64, error) {

	situationHistoryIDs := make([]int64, 0)
	for _, sh := range historySituations {
		situationHistoryIDs = append(situationHistoryIDs, sh.ID)
	}
	historySituationFacts, err := history.S().HistorySituationFactsQuerier.Query(history.S().HistorySituationFactsQuerier.Builder.GetHistorySituationFacts(situationHistoryIDs))
	if err != nil {
		zap.L().Error("history GetHistorySituationFacts", zap.Error(err))
		return nil, nil, nil, err
	}

	historyFactIDs := make([]int64, 0)
	for _, sfh := range historySituationFacts {
		historyFactIDs = append(historyFactIDs, sfh.HistoryFactID)
	}
	historyFacts, err := history.S().HistoryFactsQuerier.Query(history.S().HistoryFactsQuerier.Builder.GetHistoryFacts(historyFactIDs))
	if err != nil {
		zap.L().Error("history GetHistoryFacts", zap.Error(err))
		return nil, nil, nil, err
	}

	mapSituationFact := make(map[int64][]int64)
	mapFactSituation := make(map[int64]int64)
	for _, historySituationFact := range historySituationFacts {
		if list, exists := mapSituationFact[historySituationFact.HistorySituationID]; exists {
			mapSituationFact[historySituationFact.HistorySituationID] = append(list, historySituationFact.HistoryFactID)
		} else {
			mapSituationFact[historySituationFact.HistorySituationID] = []int64{historySituationFact.HistoryFactID}
		}

		if _, exists := mapFactSituation[historySituationFact.HistoryFactID]; exists {
			// should not happens :(
		} else {
			mapFactSituation[historySituationFact.HistoryFactID] = historySituationFact.HistorySituationID
		}
	}

	return historyFacts, mapSituationFact, mapFactSituation, nil
}

func (job FactRecalculationJob) RecalculateAndUpdateFacts(factIDs []int64, facts map[int64]engine.Fact,
	mapFactSituation map[int64]int64, mapSituations map[int64]history.HistorySituationsV4, historyFacts []history.HistoryFactsV4) (map[int64]history.HistoryFactsV4, error) {

	// Fact history recalculation + update in database
	newFactHistory := make(map[int64]history.HistoryFactsV4)
	for _, fh := range historyFacts {
		recalculate := false
		for _, factID := range factIDs {
			if fh.FactID == factID {
				recalculate = true
				break
			}
		}

		if !recalculate {
			newFactHistory[fh.ID] = fh
		} else {
			b, _ := json.Marshal(facts[fh.FactID])
			var f engine.Fact
			json.Unmarshal(b, &f) // deep copy, calculateFact is doing non-immutable operation...

			// historyFactID => historySituationFactID => historySituationID

			// recalculate
			parameters := make(map[string]interface{})
			if f.IsTemplate {
				s := mapSituations[mapFactSituation[fh.ID]]
				parameters = s.Parameters
			}

			widgetData, err := fact.ExecuteFact(fh.Ts, f, fh.SituationID, fh.SituationInstanceID, parameters, 0, 0, true)
			if err != nil {
				zap.L().Error("fact.ExecuteFact", zap.Error(err))
				continue
			}

			newFH := history.HistoryFactsV4{
				ID:                  fh.ID,
				FactID:              fh.FactID,
				FactName:            fh.FactName,
				SituationID:         fh.SituationID,
				SituationInstanceID: fh.SituationInstanceID,
				Ts:                  fh.Ts,
				Result:              *widgetData.Aggregates,
			}
			err = history.S().HistoryFactsQuerier.Update(newFH)
			if err != nil {
				zap.L().Error("HistoryFactsQuerier.Update", zap.Error(err))
				continue
			}
			newFactHistory[newFH.ID] = newFH
		}
	}
	return newFactHistory, nil
}

func (job FactRecalculationJob) RecalculateAndUpdateSituations(localRuleEngine *ruleeng.RuleEngine, s situation2.Situation, mapSituationFact map[int64][]int64,
	historySituations []history.HistorySituationsV4, newFactHistory map[int64]history.HistoryFactsV4) error {

	for _, sh := range historySituations {
		historySituationFlattenData := make(map[string]interface{})

		// Get all corresponding fact history rows
		for _, historyFactID := range mapSituationFact[sh.ID] {
			historyFact := newFactHistory[historyFactID]
			historyFactData, err := historyFact.Result.ToAbstractMap()
			if err != nil {
				zap.L().Error("", zap.Error(err))
				return err
			}
			historySituationFlattenData[historyFact.FactName] = historyFactData
		}

		for key, value := range sh.Parameters {
			historySituationFlattenData[key] = value
		}
		for key, value := range expression.GetDateKeywords(sh.Ts) {
			historySituationFlattenData[key] = value
		}

		// Evaluate expression facts
		expressionFacts := history.EvaluateExpressionFacts(s.ExpressionFacts, historySituationFlattenData)
		for key, value := range expressionFacts {
			historySituationFlattenData[key] = value
		}

		// Evaluate rules
		enabledRuleIDs, err := rule.R().GetEnabledRuleIDs(sh.SituationID, sh.Ts)
		if err != nil {
			zap.L().Error("", zap.Error(err))
			return err
		}

		metadatas := make([]metadata.MetaData, 0)
		agenda := evaluator.EvaluateRules(localRuleEngine, historySituationFlattenData, enabledRuleIDs)
		for _, agen := range agenda {
			if agen.GetName() == "set" {
				context := tasker.BuildContextData(agen.GetMetaData())
				for key, value := range agen.GetParameters() {
					metadatas = append(metadatas, metadata.MetaData{
						Key:         key,
						Value:       value,
						RuleID:      context.RuleID,
						RuleVersion: context.RuleVersion,
						CaseName:    context.CaseName,
					})
				}
			}
		}

		// Build and insert HistorySituationV4
		historySituationNew := history.HistorySituationsV4{
			ID:                  sh.ID,
			SituationID:         sh.SituationID,
			SituationInstanceID: sh.SituationInstanceID,
			Ts:                  sh.Ts,
			Parameters:          sh.Parameters,
			ExpressionFacts:     expressionFacts,
			Metadatas:           metadatas,
		}

		err = history.S().HistorySituationsQuerier.Update(historySituationNew)
		if err != nil {
			zap.L().Error("", zap.Error(err))
			return err
		}
	}
	return nil
}
