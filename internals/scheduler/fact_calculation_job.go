package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/evaluator"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/history"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tasker"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"github.com/myrteametrics/myrtea-sdk/v4/ruleeng"
	"go.uber.org/zap"
)

const timeLayout = "2006-01-02T15:04:05.000Z07:00"

// FactCalculationJob represent a scheduler job instance which process a group of facts, and persist the result in postgresql
// It also generate situations, persists them and notify the rule engine to evaluate them
type FactCalculationJob struct {
	FactIds        []int64 `json:"factIds"`
	From           string  `json:"from,omitempty"`
	To             string  `json:"to,omitempty"`
	LastDailyValue bool    `json:"lastDailyValue,omitempty"`
	Debug          bool    `json:"debug"`
	ScheduleID     int64   `json:"-"`
}

//ResolveFromAndTo resolves the expressions in parameters From and To
func (job *FactCalculationJob) ResolveFromAndTo(t time.Time) (time.Time, time.Time, error) {

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
		zap.L().Error("Error processing From expression in fact calculation jon", zap.Error(err))
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

// IsValid checks if an internal schedule job definition is valid and has no missing mandatory fields
func (job FactCalculationJob) IsValid() (bool, error) {
	if job.FactIds == nil {
		return false, errors.New("missing FactIds")
	}
	if len(job.FactIds) <= 0 {
		return false, errors.New("missing FactIds")
	}
	return true, nil
}

// Run contains all the business logic of the job
func (job FactCalculationJob) Run() {
	if S().ExistingRunningJob(job.ScheduleID) {
		zap.L().Info("Skipping FactScheduleJob because last execution is still running", zap.Int64s("ids", job.FactIds))
		return
	}
	S().AddRunningJob(job.ScheduleID)

	zap.L().Info("Fact calculation job started", zap.Int64s("ids", job.FactIds))

	t := time.Now().Truncate(1 * time.Second).UTC()

	calendar.CBase().Update()
	localRuleEngine, err := evaluator.BuildLocalRuleEngine("standart")
	if err != nil {
		zap.L().Error("BuildLocalRuleEngine", zap.Error(err))
		return
	}

	// FIXME: Support retroactive update
	// if job.From != "" {
	// 	err := job.update(t)
	// 	if err != nil {
	// 		zap.L().Error("Error updating fact instances", zap.Error(err))
	// 	}
	// 	zap.L().Info("FactScheduleJob Ended", zap.Int64s("ids", job.FactIds))
	// 	S().RemoveRunningJob(job.ScheduleID)
	// 	return
	// }

	situationsToUpdate, err := CalculateAndPersistFacts(t, job.FactIds)
	if err != nil {
		zap.L().Error("CalculateAndPersistFacts", zap.Error(err))
		S().RemoveRunningJob(job.ScheduleID)
		return
	}

	taskBatchs, err := CalculateAndPersistSituations(localRuleEngine, situationsToUpdate)
	if err != nil {
		zap.L().Error("CalculateAndPersistSituations", zap.Error(err))
		S().RemoveRunningJob(job.ScheduleID)
		return
	}

	tasker.T().BatchReceiver <- taskBatchs
	zap.L().Info("FactScheduleJob Ended", zap.Int64s("ids", job.FactIds))

	S().RemoveRunningJob(job.ScheduleID)
}

// ExternalAggregate contains all information to store a new aggregat in postgresql
type ExternalAggregate struct {
	FactID              int64       `json:"factId"`
	SituationID         int64       `json:"situationId"`
	SituationInstanceID int64       `json:"situationInstanceId"`
	Time                time.Time   `json:"time"`
	Value               reader.Item `json:"value"`
}

// ReceiveAndPersistFacts process a slice of ExternalAggregates and trigger all standard fact-situation-rule process
func ReceiveAndPersistFacts(aggregates []ExternalAggregate) (map[string]history.HistoryRecordV4, error) {

	situationsToUpdate := make(map[string]history.HistoryRecordV4, 0)
	for _, agg := range aggregates {

		t := agg.Time.UTC().Truncate(time.Second)

		f, found, err := fact.R().Get(agg.FactID)
		if err != nil {
			return make(map[string]history.HistoryRecordV4), err
		}
		if !found {
			return make(map[string]history.HistoryRecordV4), errors.New("not found")
		}

		s, found, err := situation.R().Get(agg.SituationID)
		if err != nil {
			return make(map[string]history.HistoryRecordV4), err
		}
		if !found {
			return make(map[string]history.HistoryRecordV4), errors.New("not found")
		}

		found = false
		for _, factID := range s.Facts {
			if f.ID == factID {
				found = true
			}
		}
		if !found {
			zap.L().Warn("Fact doesn't exist in situation", zap.Int64("factID", f.ID), zap.Int64("situationID", s.ID), zap.Int64s("factIDs", s.Facts))
			continue
		}

		si, found, err := situation.R().GetTemplateInstance(agg.SituationInstanceID)
		if err != nil {
			return make(map[string]history.HistoryRecordV4), err
		}
		if !found {
			return make(map[string]history.HistoryRecordV4), errors.New("not found")
		}

		if s.ID != si.SituationID {
			zap.L().Warn("invalid s.ID != si.SituationID")
			continue
		}

		factSituationsHistory, err := GetFactSituations(f, t)
		if err != nil {
			zap.L().Warn("getFactSituations", zap.Int64("factID", f.ID), zap.Error(err))
			continue
		}
		if len(factSituationsHistory) == 0 {
			zap.L().Warn("fact has no situation history", zap.Int64("factID", f.ID))
			continue
		}

		if !f.IsTemplate {
			// calculate
			// already done !

			historyFactNew := history.HistoryFactsV4{
				// ID:                  -1,
				FactID:              f.ID,
				FactName:            f.Name,
				SituationID:         0,
				SituationInstanceID: 0,
				Ts:                  t,
				Result:              agg.Value,
			}
			historyFactNew.ID, err = history.S().HistoryFactsQuerier.Insert(historyFactNew)
			if err != nil {
				// err
			}

			for _, sh := range factSituationsHistory {
				key := fmt.Sprintf("%d-%d", sh.SituationID, sh.SituationInstanceID)
				if _, ok := situationsToUpdate[key]; !ok {
					situationsToUpdate[key] = history.HistoryRecordV4{
						SituationID:         sh.SituationID,
						SituationInstanceID: sh.SituationInstanceID,
						Ts:                  t,
						Parameters:          sh.Parameters,
						HistoryFacts:        []history.HistoryFactsV4{historyFactNew},
					}
				} else {
					situation := situationsToUpdate[key]
					situation.HistoryFacts = append(situation.HistoryFacts, historyFactNew)
					situationsToUpdate[key] = situation
				}
			}
		} else {
			for _, sh := range factSituationsHistory {
				// calculate
				// already done !

				historyFactNew := history.HistoryFactsV4{
					// ID:                  -1,
					FactID:              f.ID,
					FactName:            f.Name,
					SituationID:         sh.SituationID,
					SituationInstanceID: sh.SituationInstanceID,
					Ts:                  t,
					Result:              agg.Value,
				}
				historyFactNew.ID, err = history.S().HistoryFactsQuerier.Insert(historyFactNew)
				if err != nil {
					// err
				}

				key := fmt.Sprintf("%d-%d", sh.SituationID, sh.SituationInstanceID)
				if _, ok := situationsToUpdate[key]; !ok {
					situationsToUpdate[key] = history.HistoryRecordV4{
						SituationID:         sh.SituationID,
						SituationInstanceID: sh.SituationInstanceID,
						Ts:                  t,
						Parameters:          sh.Parameters,
						HistoryFacts:        []history.HistoryFactsV4{historyFactNew},
					}
				} else {
					situation := situationsToUpdate[key]
					situation.HistoryFacts = append(situation.HistoryFacts, historyFactNew)
					situationsToUpdate[key] = situation
				}
			}

		}
	}
	return situationsToUpdate, nil
}

func CalculateAndPersistFacts(t time.Time, factIDs []int64) (map[string]history.HistoryRecordV4, error) {
	situationsToUpdate := make(map[string]history.HistoryRecordV4, 0)

	for _, factID := range factIDs {
		f, found, err := fact.R().Get(factID)
		if err != nil {
			zap.L().Error("Error Getting the Fact, skipping fact calculation...", zap.Int64("factID", factID))
			continue
		}
		if !found {
			zap.L().Warn("Fact does not exists, skipping fact calculation...", zap.Int64("factID", factID))
			continue
		}

		factSituationsHistory, err := GetFactSituations(f, t)
		if err != nil {
			continue
		}
		if len(factSituationsHistory) == 0 {
			continue
		}

		if !f.IsTemplate {
			widgetData, err := calculateFact(t, f, 0, 0, nil, false)
			if err != nil {
				zap.L().Error("Fact calculation Error, skipping fact calculation...", zap.Int64("id", f.ID), zap.Any("fact", f), zap.Error(err))
				continue
			}

			historyFactNew := history.HistoryFactsV4{
				// ID:                  -1,
				FactID:              f.ID,
				FactName:            f.Name,
				SituationID:         0,
				SituationInstanceID: 0,
				Ts:                  t,
				Result:              *widgetData.Aggregates,
			}
			historyFactNew.ID, err = history.S().HistoryFactsQuerier.Insert(historyFactNew)
			if err != nil {
				// err
			}

			for _, sh := range factSituationsHistory {
				key := fmt.Sprintf("%d-%d", sh.SituationID, sh.SituationInstanceID)
				if _, ok := situationsToUpdate[key]; !ok {
					situationsToUpdate[key] = history.HistoryRecordV4{
						SituationID:         sh.SituationID,
						SituationInstanceID: sh.SituationInstanceID,
						Ts:                  t,
						Parameters:          sh.Parameters,
						HistoryFacts:        []history.HistoryFactsV4{historyFactNew},
					}
				} else {
					situation := situationsToUpdate[key]
					situation.HistoryFacts = append(situation.HistoryFacts, historyFactNew)
					situationsToUpdate[key] = situation
				}
			}
		} else {
			for _, sh := range factSituationsHistory {

				var fCopy engine.Fact
				fData, _ := json.Marshal(f)
				json.Unmarshal(fData, &fCopy)
				widgetData, err := calculateFact(t, fCopy, sh.SituationID, sh.SituationInstanceID, sh.Parameters, false)
				if err != nil {
					zap.L().Error("Fact calculation Error, skipping fact calculation...", zap.Int64("id", f.ID), zap.Any("fact", f), zap.Error(err))
					continue
				}

				historyFactNew := history.HistoryFactsV4{
					// ID:                  -1,
					FactID:              f.ID,
					FactName:            f.Name,
					SituationID:         sh.SituationID,
					SituationInstanceID: sh.SituationInstanceID,
					Ts:                  t,
					Result:              *widgetData.Aggregates,
				}
				historyFactNew.ID, err = history.S().HistoryFactsQuerier.Insert(historyFactNew)
				if err != nil {
					// err
				}

				key := fmt.Sprintf("%d-%d", sh.SituationID, sh.SituationInstanceID)
				if _, ok := situationsToUpdate[key]; !ok {
					situationsToUpdate[key] = history.HistoryRecordV4{
						SituationID:         sh.SituationID,
						SituationInstanceID: sh.SituationInstanceID,
						Ts:                  t,
						Parameters:          sh.Parameters,
						HistoryFacts:        []history.HistoryFactsV4{historyFactNew},
					}
				} else {
					situation := situationsToUpdate[key]
					situation.HistoryFacts = append(situation.HistoryFacts, historyFactNew)
					situationsToUpdate[key] = situation
				}
			}
		}
	}

	return situationsToUpdate, nil
}

func calculateFact(t time.Time, f engine.Fact, situationID int64, situationInstanceID int64, placeholders map[string]string, update bool) (reader.WidgetData, error) {
	pf, err := fact.Prepare(&f, -1, -1, t, placeholders, update)
	if err != nil {
		zap.L().Error("Cannot prepare fact", zap.Int64("id", f.ID), zap.Any("fact", f), zap.Error(err))
		return reader.WidgetData{}, err
	}

	widgetData, err := fact.Execute(pf)
	if err != nil {
		zap.L().Error("Cannot execute fact", zap.Int64("id", f.ID), zap.Any("pf", pf), zap.Error(err))
		return reader.WidgetData{}, err
	}

	// pluginBaseline, err := baseline.P()
	// if err == nil {
	// 	values, err := pluginBaseline.Baseline.GetBaselineValues(-1, f.ID, situationID, situationInstanceID, t)
	// 	if err != nil {
	// 		zap.L().Error("Cannot fetch fact baselines", zap.Int64("id", f.ID), zap.Error(err))
	// 	}
	// 	widgetData.Aggregates.Baselines = values
	// }
	return *widgetData, nil
}

func CalculateAndPersistSituations(localRuleEngine *ruleeng.RuleEngine, situationsToUpdate map[string]history.HistoryRecordV4) ([]tasker.TaskBatch, error) {
	taskBatchs := make([]tasker.TaskBatch, 0)
	for _, situationToUpdate := range situationsToUpdate {

		// Flatten parameters from situation definition + situation instance definition
		s, parameters, err := history.ExtractSituationData(situationToUpdate.SituationID, situationToUpdate.SituationInstanceID)
		if err != nil {
			zap.L().Error("", zap.Error(err))
			continue
		}
		zap.L().Info("", zap.Any("parameters", parameters))

		historyFactsAll, historySituationFlattenData, err := history.S().ExtractFactData(situationToUpdate.HistoryFacts, s.Facts)
		if err != nil {
			zap.L().Error("", zap.Error(err))
			continue
		}
		for key, value := range parameters {
			historySituationFlattenData[key] = value
		}
		for key, value := range expression.GetDateKeywords(situationToUpdate.Ts) {
			historySituationFlattenData[key] = value
		}
		zap.L().Info("flatten data", zap.Any("historySituationFlattenData", historySituationFlattenData))

		// Evaluate expression facts
		expressionFacts := history.EvaluateExpressionFacts(s.ExpressionFacts, historySituationFlattenData)
		zap.L().Info("expressionfacts", zap.Any("expressionFacts", expressionFacts))

		for key, value := range expressionFacts {
			historySituationFlattenData[key] = value
		}

		// Evaluate rules
		enabledRuleIDs, err := rule.R().GetEnabledRuleIDs(situationToUpdate.SituationID, situationToUpdate.Ts)
		if err != nil {
			zap.L().Error("", zap.Error(err))
		}

		metadatas := make([]models.MetaData, 0)
		agenda := evaluator.EvaluateRules(localRuleEngine, historySituationFlattenData, enabledRuleIDs)
		if agenda != nil {
			for _, agen := range agenda {
				if agen.GetName() != "set" {
					context := tasker.BuildContextData(agen.GetMetaData())
					for key, value := range agen.GetParameters() {
						metadatas = append(metadatas, models.MetaData{
							Key:         key,
							Value:       value,
							RuleID:      context.RuleID,
							RuleVersion: context.RuleVersion,
							CaseName:    context.CaseName,
						})
					}
				}
			}
		}

		// Build and insert HistorySituationV4
		historySituationNew := history.HistorySituationsV4{
			// ID:                    -1,
			SituationID:         situationToUpdate.SituationID,
			SituationInstanceID: situationToUpdate.SituationInstanceID,
			Ts:                  situationToUpdate.Ts,
			Parameters:          parameters,
			ExpressionFacts:     expressionFacts,
			Metadatas:           make([]models.MetaData, 0),
		}
		historySituationNew.ID, err = history.S().HistorySituationsQuerier.Insert(historySituationNew)
		if err != nil {
			zap.L().Error("", zap.Error(err))
		}
		zap.L().Info("historySituationNew", zap.Any("historySituationNew", historySituationNew))

		// Build and insert HistorySituationFactsV4
		historySituationFactNew := make([]history.HistorySituationFactsV4, 0)
		for _, historyFactNew := range historyFactsAll {
			historySituationFactNew = append(historySituationFactNew, history.HistorySituationFactsV4{ // Replace entry for existing factID with new HistorySituationFactsV4{}
				HistorySituationID: historySituationNew.ID,
				HistoryFactID:      historyFactNew.ID,
				FactID:             historyFactNew.FactID,
			})
		}

		err = history.S().HistorySituationFactsQuerier.Execute(history.S().HistorySituationFactsQuerier.Builder.InsertBulk(historySituationFactNew))
		if err != nil {
			zap.L().Error("", zap.Error(err))
		}
		zap.L().Info("historySituationFactNew", zap.Any("historySituationFactNew", historySituationFactNew))

		if agenda != nil {
			taskBatchs = append(taskBatchs, tasker.TaskBatch{
				Context: map[string]interface{}{
					"situationID":                 situationToUpdate.SituationID,
					"templateInstanceID":          situationToUpdate.SituationInstanceID,
					"ts":                          situationToUpdate.Ts,
					"historySituationFlattenData": historySituationFlattenData,
					"situationHistoryID":          historySituationNew.ID,
				},
				Agenda: agenda,
			})
		}
	}

	return taskBatchs, nil
}

// // UpdateSituations creates the new instances of the situations in the history and evaluates them
// func UpdateSituations(situationsToUpdate map[string]history.HistoryRecordV4) ([]evaluator.SituationToEvaluate, error) {

// 	situationsToEvalute := make([]evaluator.SituationToEvaluate, 0)
// 	for _, record := range situationsToUpdate {

// 		// create factsHistory from situationFacts
// 		situationFacts, err := situation.R().GetFacts(record.ID)
// 		if err != nil {
// 			zap.L().Error("Get situation facts", zap.Int64("situationID", record.ID), zap.Error(err))
// 			continue
// 		}

// 		// merge values from lasthistory.HistoryRecordV4 into factsHistory
// 		lasthistory.HistoryRecordV4, err := situation.GetFromHistory(record.ID, record.TS, record.TemplateInstanceID, true)
// 		if err != nil {
// 			zap.L().Error("Get situation from history", zap.Int64("situationID", record.ID), zap.Time("ts", record.TS), zap.Error(err))
// 			continue
// 		}

// 		factsHistory := make(map[int64]*time.Time)
// 		for _, factID := range situationFacts {
// 			factsHistory[factID] = nil
// 		}

// 		if lasthistory.HistoryRecordV4 != nil {
// 			for factID, factTS := range lasthistory.HistoryRecordV4.FactsIDS {
// 				factsHistory[factID] = factTS
// 			}
// 		}
// 		// merge new values into factsHistory
// 		for factID, factTS := range record.FactsIDS {
// 			factsHistory[factID] = factTS
// 		}

// 		record.FactsIDS = factsHistory

// 		evaluatedExpressionFacts, err := evaluateExpressionFacts(record, record.TS)
// 		if err != nil {
// 			zap.L().Warn("cannot evaluate expression facts", zap.Error(err))
// 			continue
// 		}
// 		record.EvaluatedExpressionFacts = evaluatedExpressionFacts

// 		err = situation.Persist(record, false)
// 		if err != nil {
// 			zap.L().Error("UpdateSituations.persistSituation:", zap.Error(err))
// 			continue
// 		}
// 		situationsToEvalute = append(situationsToEvalute,
// 			evaluator.SituationToEvaluate{
// 				ID:                 record.ID,
// 				TS:                 record.TS,
// 				TemplateInstanceID: record.TemplateInstanceID,
// 			},
// 		)
// 	}

// 	return situationsToEvalute, nil
// }

// func evaluateExpressionFacts(record history.HistoryRecordV4, t time.Time) (map[string]interface{}, error) {
// 	evaluatedExpressionFacts := make(map[string]interface{})

// 	s, found, err := situation.R().Get(record.ID)
// 	if err != nil {
// 		zap.L().Error("Get Situation", zap.Int64("situationID", record.ID), zap.Error(err))
// 		return evaluatedExpressionFacts, err
// 	}
// 	if !found {
// 		zap.L().Warn("Situation not found", zap.Int64("situationID", s.ID))
// 		return evaluatedExpressionFacts, fmzap.L().Errorf("", zap.Error("situation not found with ID = %d", record.ID))
// 	}

// 	data, err := flattenSituationData(record)
// 	if err != nil {
// 		return evaluatedExpressionFacts, err
// 	}

// 	//Add date keywords in situation data
// 	for key, value := range expression.GetDateKeywords(t) {
// 		data[key] = value
// 	}

// 	for _, expressionFact := range s.ExpressionFacts {
// 		result, err := expression.Process(expression.LangEval, expressionFact.Expression, data)
// 		if err != nil {
// 			zap.L().Debug("Cannot process gval factExpression", zap.Error(err))
// 			continue
// 		}
// 		if expression.IsInvalidNumber(result) {
// 			continue
// 		}

// 		data[expressionFact.Name] = result
// 		evaluatedExpressionFacts[expressionFact.Name] = result
// 	}
// 	return evaluatedExpressionFacts, nil
// }

// func flattenSituationData(record history.HistoryRecordV4) (map[string]interface{}, error) {
// 	situationData := make(map[string]interface{})
// 	for factID, factTS := range record.FactsIDS {
// 		f, found, err := fact.R().Get(factID)
// 		if err != nil {
// 			zap.L().Error("get fact", zap.Error(err))
// 			return nil, err
// 		}
// 		if !found {
// 			zap.L().Warn("fact not found", zap.Int64("factID", factID))
// 			return nil, fmzap.L().Errorf("", zap.Error("fact not found with id=%d", factID))
// 		}
// 		if factTS == nil {
// 			// zap.L().Warn("At least one fact has never been calculated", zap.Int64("id", f.ID), zap.String("name", f.Name))
// 			// return nil, fmzap.L().Errorf("", zap.Error("at least one fact has never been calculated, id=%d, name=%s", f.ID, f.Name))
// 			continue
// 		}

// 		item, _, err := fact.GetFactResultFromHistory(factID, *factTS, record.ID, record.TemplateInstanceID, false, -1)
// 		if err != nil {
// 			return nil, err
// 		}
// 		itemData, err := item.ToAbstractMap()
// 		if err != nil {
// 			zap.L().Error("Convert item to abstractmap", zap.Error(err))
// 			return nil, err
// 		}

// 		situationData[f.Name] = itemData
// 	}
// 	for key, value := range record.Parameters {
// 		situationData[key] = value
// 	}

// 	return situationData, nil
// }

// GetFactSituations returns all situation linked to a fact
func GetFactSituations(fact engine.Fact, t time.Time) ([]history.HistoryRecordV4, error) {
	factSituationsHistory := make([]history.HistoryRecordV4, 0)
	factSituations, err := situation.R().GetSituationsByFactID(fact.ID, true)
	if err != nil {
		zap.L().Error("Cannot get the situations to update for fact", zap.Int64("id", fact.ID), zap.Any("fact", fact), zap.Error(err))
		return nil, err
	}
	for _, s := range factSituations {
		if s.IsTemplate {
			templateInstances, err := situation.R().GetAllTemplateInstances(s.ID)
			if err != nil {
				zap.L().Error("Cannot get the situations template instances for situation", zap.Int64("id", s.ID), zap.Any("fact", fact), zap.Error(err))
				return nil, err
			}
			for _, ti := range templateInstances {
				calendarID := ti.CalendarID
				if calendarID == 0 {
					calendarID = s.CalendarID
					zap.L().Debug("Situation template withour calendar id, taking the one from the situation with id: ", zap.Int64("id", s.ID))
				}

				//We consider that if the calendar is not found then is in valid period
				found, valid, _ := calendar.CBase().InPeriodFromCalendarID(calendarID, t)
				if !found || valid {
					sh := history.HistoryRecordV4{
						SituationID:         s.ID,
						SituationInstanceID: ti.ID,
						Parameters:          map[string]string{},
					}
					sh.OverrideParameters(s.Parameters)
					sh.OverrideParameters(ti.Parameters)
					factSituationsHistory = append(factSituationsHistory, sh)
				} else {
					zap.L().Debug("Situation template not within a valid calendar period, situation id: ", zap.Int64("id", s.ID))
				}
			}
		} else {
			//We consider that if the calendar is not found then is in valid period
			found, valid, _ := calendar.CBase().InPeriodFromCalendarID(s.CalendarID, t)
			if !found || valid {
				factSituationsHistory = append(factSituationsHistory, history.HistoryRecordV4{
					SituationID:         s.ID,
					SituationInstanceID: 0,
					Parameters:          s.Parameters,
				})
			} else {
				zap.L().Debug("Situation not within a valid calendar period, situation id: ", zap.Int64("id", s.ID))
			}
		}
	}
	return factSituationsHistory, nil
}

// ----------------------------------

// UnmarshalJSON unmarshals a quoted json string to a valid FactCalculationJob struct
func (job *FactCalculationJob) UnmarshalJSON(data []byte) error {
	type Alias FactCalculationJob
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(job),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if _, _, err := job.ResolveFromAndTo(time.Now()); err != nil {
		return err
	}
	return nil
}
