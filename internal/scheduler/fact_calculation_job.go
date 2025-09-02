package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/calendar"
	fact2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/metadata"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/reader"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/situation"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/evaluator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/rule"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tasker"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/history"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"github.com/myrteametrics/myrtea-sdk/v5/ruleeng"
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

// ResolveFromAndTo resolves the expressions in parameters From and To
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
	if job.From != "" {
		FactRecalculationJob{
			FactIds:        job.FactIds,
			From:           job.From,
			To:             job.To,
			LastDailyValue: job.LastDailyValue,
			Debug:          job.Debug,
			ScheduleID:     job.ScheduleID,
		}.Run()
		return
	}

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
		S().RemoveRunningJob(job.ScheduleID)
		return
	}

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

	situationsToUpdate := make(map[string]history.HistoryRecordV4)
	for _, agg := range aggregates {
		t := agg.Time.UTC().Truncate(time.Second)

		f, found, err := fact2.R().Get(agg.FactID)
		if err != nil {
			zap.L().Error("ReceiveAndPersistFacts fact get error, skipping aggregate", zap.Error(err), zap.Int64("situationId", agg.SituationID), zap.Int64("situationInstanceId", agg.SituationInstanceID), zap.Int64("factId", agg.FactID))
			continue
		}
		if !found {
			zap.L().Error("ReceiveAndPersistFacts fact not found, skipping aggregate", zap.Int64("situationId", agg.SituationID), zap.Int64("situationInstanceId", agg.SituationInstanceID), zap.Int64("factId", agg.FactID))
			continue
		}

		s, found, err := situation.R().Get(agg.SituationID)
		if err != nil {
			zap.L().Error("ReceiveAndPersistFacts situation get error, skipping aggregate", zap.Error(err), zap.Int64("situationId", agg.SituationID), zap.Int64("situationInstanceId", agg.SituationInstanceID), zap.Int64("factId", agg.FactID))
			continue
		}
		if !found {
			zap.L().Error("ReceiveAndPersistFacts situation not found, skipping aggregate", zap.Int64("situationId", agg.SituationID), zap.Int64("situationInstanceId", agg.SituationInstanceID), zap.Int64("factId", agg.FactID))
			continue
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
			zap.L().Error("ReceiveAndPersistFacts situationInstance get error, skipping aggregate", zap.Error(err), zap.Int64("situationId", agg.SituationID), zap.Int64("situationInstanceId", agg.SituationInstanceID), zap.Int64("factId", agg.FactID))
			continue
		}
		if !found {
			zap.L().Error("ReceiveAndPersistFacts situationInstance not found, skipping aggregate", zap.Int64("situationId", agg.SituationID), zap.Int64("situationInstanceId", agg.SituationInstanceID), zap.Int64("factId", agg.FactID))
			continue
		}

		if s.ID != si.SituationID {
			zap.L().Warn("invalid s.ID != si.SituationID")
			continue
		}

		factSituationsHistory, err := GetEnabledSituations(f, t)
		if err != nil {
			zap.L().Warn("getFactSituations", zap.Int64("factID", f.ID), zap.Error(err))
			continue
		}
		if len(factSituationsHistory) == 0 {
			zap.L().Debug("Fact has no enabled situations", zap.Int64("factID", f.ID))
			continue
		}

		// zap.L().Sugar().Info("factSituationsHistory", factSituationsHistory)

		if !f.IsTemplate {
			// zap.L().Sugar().Info("fact is not a template")
			// calculate
			// already done !

			widgetData := &reader.WidgetData{
				Aggregates: &agg.Value,
			}

			fact2.GetBaselineValues(widgetData, agg.FactID, agg.SituationID, agg.SituationInstanceID, agg.Time)

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
				zap.L().Error("", zap.Error(err))
			}

			for _, sh := range factSituationsHistory {
				if sh.SituationID != s.ID || sh.SituationInstanceID != si.ID {
					continue
				}
				key := fmt.Sprintf("%d-%d", sh.SituationID, sh.SituationInstanceID)
				if _, ok := situationsToUpdate[key]; !ok {
					situationsToUpdate[key] = history.HistoryRecordV4{
						SituationID:         sh.SituationID,
						SituationInstanceID: sh.SituationInstanceID,
						Ts:                  t,
						Parameters:          sh.Parameters,
						HistoryFacts:        []history.HistoryFactsV4{historyFactNew},
						EnableDependsOn:     sh.EnableDependsOn,
						DependsOnParameters: sh.DependsOnParameters,
					}
				} else {
					situation := situationsToUpdate[key]
					situation.HistoryFacts = append(situation.HistoryFacts, historyFactNew)
					situationsToUpdate[key] = situation
				}
			}
		} else {
			// zap.L().Sugar().Info("fact IS a template")
			for _, sh := range factSituationsHistory {
				if sh.SituationID != s.ID || sh.SituationInstanceID != si.ID {
					// zap.L().Sugar().Info(sh.SituationID, s.ID, sh.SituationInstanceID, si.ID)
					continue
				}

				// calculate
				// already done !

				widgetData := &reader.WidgetData{
					Aggregates: &agg.Value,
				}

				fact2.GetBaselineValues(widgetData, agg.FactID, agg.SituationID, agg.SituationInstanceID, agg.Time)

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
					zap.L().Error("", zap.Error(err))
				}
				// zap.L().Sugar().Info("insert fact", historyFactNew)

				key := fmt.Sprintf("%d-%d", sh.SituationID, sh.SituationInstanceID)
				if _, ok := situationsToUpdate[key]; !ok {
					situationsToUpdate[key] = history.HistoryRecordV4{
						SituationID:         sh.SituationID,
						SituationInstanceID: sh.SituationInstanceID,
						Ts:                  t,
						Parameters:          sh.Parameters,
						HistoryFacts:        []history.HistoryFactsV4{historyFactNew},
						EnableDependsOn:     sh.EnableDependsOn,
						DependsOnParameters: sh.DependsOnParameters,
					}
				} else {
					situation := situationsToUpdate[key]
					situation.HistoryFacts = append(situation.HistoryFacts, historyFactNew)
					situationsToUpdate[key] = situation
				}
			}

		}
	}
	// zap.L().Sugar().Info("situationToUpdate ", situationsToUpdate)
	return situationsToUpdate, nil
}

func CalculateAndPersistFacts(t time.Time, factIDs []int64) (map[string]history.HistoryRecordV4, error) {
	situationsToUpdate := make(map[string]history.HistoryRecordV4)

	for _, factID := range factIDs {
		f, found, err := fact2.R().Get(factID)
		if err != nil {
			zap.L().Error("Error Getting the Fact, skipping fact calculation...", zap.Int64("factID", factID))
			continue
		}
		if !found {
			zap.L().Warn("Fact does not exists, skipping fact calculation...", zap.Int64("factID", factID))
			continue
		}

		// get all enabled situations linked to this fact
		factSituationsHistory, err := GetEnabledSituations(f, t)
		if err != nil {
			zap.L().Debug("Fact has no enabled situations", zap.Int64("factID", f.ID))
			continue
		}
		if len(factSituationsHistory) == 0 {
			continue
		}

		if !f.IsTemplate {
			// execute fact, to get results
			widgetData, err := fact2.ExecuteFact(t, f, 0, 0, make(map[string]interface{}), -1, -1, false)
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
				zap.L().Error("", zap.Error(err))
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
						EnableDependsOn:     sh.EnableDependsOn,
						DependsOnParameters: sh.DependsOnParameters,
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
				err := json.Unmarshal(fData, &fCopy)
				if err != nil {
					zap.L().Error("Fact calculation Error (json.Unmarshal), skipping fact calculation...", zap.Int64("id", f.ID), zap.Any("fact", f), zap.Error(err))
					continue
				}

				widgetData, err := fact2.ExecuteFact(t, fCopy, sh.SituationID, sh.SituationInstanceID, sh.Parameters, -1, -1, false)
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
					zap.L().Error("", zap.Error(err))
				}

				key := fmt.Sprintf("%d-%d", sh.SituationID, sh.SituationInstanceID)
				if _, ok := situationsToUpdate[key]; !ok {
					situationsToUpdate[key] = history.HistoryRecordV4{
						SituationID:         sh.SituationID,
						SituationInstanceID: sh.SituationInstanceID,
						Ts:                  t,
						Parameters:          sh.Parameters,
						HistoryFacts:        []history.HistoryFactsV4{historyFactNew},
						EnableDependsOn:     sh.EnableDependsOn,
						DependsOnParameters: sh.DependsOnParameters,
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

func CalculateAndPersistSituations(localRuleEngine *ruleeng.RuleEngine, situationsToUpdate map[string]history.HistoryRecordV4) ([]tasker.TaskBatch, error) {
	taskBatchs := make([]tasker.TaskBatch, 0)
	taskBatchsMap := make(map[string]tasker.TaskBatch)
	situationHistoryMetadata := make(map[model.Key]map[string]interface{})
	for _, situationToUpdate := range situationsToUpdate {

		// zap.L().Sugar().Info(situationToUpdate)

		// Flatten parameters from situation definition + situation instance definition
		s, parameters, err := history.ExtractSituationData(situationToUpdate.SituationID, situationToUpdate.SituationInstanceID)
		if err != nil {
			zap.L().Error("", zap.Error(err))
			continue
		}

		// zap.L().Sugar().Info(s, parameters)

		historyFactsAll, historySituationFlattenData, err := history.S().ExtractFactData(situationToUpdate.SituationID, situationToUpdate.SituationInstanceID, situationToUpdate.HistoryFacts, s.Facts)
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

		// zap.L().Sugar().Info("historyFactsAll", historyFactsAll)
		// zap.L().Sugar().Info("historySituationFlattenData", historySituationFlattenData)

		// Evaluate expression facts
		expressionFacts := history.EvaluateExpressionFacts(s.ExpressionFacts, historySituationFlattenData)
		for key, value := range expressionFacts {
			historySituationFlattenData[key] = value
		}

		// Evaluate rules
		enabledRuleIDs, err := rule.R().GetEnabledRuleIDs(situationToUpdate.SituationID, situationToUpdate.Ts)
		if err != nil {
			zap.L().Error("", zap.Error(err))
		}

		metadatas := make([]metadata.MetaData, 0)
		agenda := evaluator.EvaluateRules(localRuleEngine, historySituationFlattenData, enabledRuleIDs)
		var filteredAgenda []ruleeng.Action
		var prev *history.HistorySituationsV4 = nil
		for _, agen := range agenda {
			if agen.GetName() == tasker.ActionSet {
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
				continue
			}
			if !agen.GetCheckPrevSetAction() {
				filteredAgenda = append(filteredAgenda, agen)
				continue
			}

			if agen.GetName() == tasker.ActionCreateIssue || agen.GetName() == tasker.ActionSituationReporting {
				// Load previous history if necessary
				if prev == nil {
					latestHistory, err := history.S().HistorySituationsQuerier.GetLatestHistory(situationToUpdate.SituationID, situationToUpdate.SituationInstanceID)
					if err != nil {
						filteredAgenda = append(filteredAgenda, agen)
						continue
					}
					prev = &latestHistory
				}
				isCritical := false
				for _, metadata := range prev.Metadatas {
					if strings.EqualFold(metadata.Value.(string), model.Critical.String()) {
						isCritical = true
						break
					}
				}

				if !isCritical {
					filteredAgenda = append(filteredAgenda, agen)
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
			Metadatas:           metadatas,
		}
		historySituationNew.ID, err = history.S().HistorySituationsQuerier.Insert(historySituationNew)
		if err != nil {
			zap.L().Error("", zap.Error(err))
		}
		// zap.L().Sugar().Info("insert new situation", historySituationNew)

		situationHistoryMetadata[model.Key{SituationID: situationToUpdate.SituationID, SituationInstanceID: situationToUpdate.SituationInstanceID}] = map[string]interface{}{
			"HistorySituation": historySituationNew,
		}
		// Build and insert HistorySituationFactsV4
		historySituationFactNew := make([]history.HistorySituationFactsV4, 0)
		for _, historyFactNew := range historyFactsAll {
			historySituationFactNew = append(historySituationFactNew, history.HistorySituationFactsV4{ // Replace entry for existing factID with new HistorySituationFactsV4{}
				HistorySituationID: historySituationNew.ID,
				HistoryFactID:      historyFactNew.ID,
				FactID:             historyFactNew.FactID,
			})
		}

		// zap.L().Sugar().Info("historySituationFactNew", historySituationFactNew)

		err = history.S().HistorySituationFactsQuerier.Execute(history.S().HistorySituationFactsQuerier.Builder.InsertBulk(historySituationFactNew))
		if err != nil {
			zap.L().Error(fmt.Sprintf("error inserting historySituationFact: make sure you added all facts of situation (%d) to a scheduler", situationToUpdate.SituationID), zap.Error(err))
		}

		if filteredAgenda != nil {
			newTaskBatch := tasker.TaskBatch{
				Context: map[string]interface{}{
					"situationID":                 situationToUpdate.SituationID,
					"templateInstanceID":          situationToUpdate.SituationInstanceID,
					"ts":                          situationToUpdate.Ts,
					"historySituationFlattenData": historySituationFlattenData,
					"situationHistoryID":          historySituationNew.ID,
				},
				Agenda: filteredAgenda,
			}
			taskBatchs = append(taskBatchs, newTaskBatch)
			key := fmt.Sprintf("%v-%v", situationToUpdate.SituationID, situationToUpdate.SituationInstanceID)
			taskBatchsMap[key] = newTaskBatch
		}
	}

	filteredTaskBatch := filterTaskByDependency(situationsToUpdate, situationHistoryMetadata, taskBatchsMap)

	return filteredTaskBatch, nil
}

// filtration
func filterTaskByDependency(situationsToUpdate map[string]history.HistoryRecordV4, situationHistoryMetadata map[model.Key]map[string]interface{}, taskBatchsMap map[string]tasker.TaskBatch) []tasker.TaskBatch {
	filteredTaskBatch := make(map[string]tasker.TaskBatch, len(taskBatchsMap))
	for key, taskBatch := range taskBatchsMap {
		filteredTaskBatch[key] = taskBatch
	}

	for _, situation := range situationsToUpdate {
		if situation.EnableDependsOn {

			DependsOnMetadata := situation.DependsOnParameters[KeyMetadataDependsOn]
			DependsOnMetadataValue := situation.DependsOnParameters[ValueMetadataDependsOn]
			keychild := fmt.Sprintf("%v-%v", situation.SituationID, situation.SituationInstanceID)

			keyParent, idSituationDependsOn, idInstanceDependsOn, err := generateKeyAndValues(situation.DependsOnParameters)
			if err != nil {
				zap.L().Error("Error to generating key", zap.Error(err))
				continue
			}

			IsChildCritical := false

			// check if an child is critical
			if childFilterdTaskBatch, exists := filteredTaskBatch[keychild]; exists {
				for _, agenda := range childFilterdTaskBatch.Agenda {
					if agenda.GetName() == ActionSetValue &&
						agenda.GetEnableDependsForALLAction() &&
						agenda.GetEnabledDependsAction() {
						metadataInterface, err := agenda.GetParameters()[DependsOnMetadata]
						if err {
							metadata, err := metadataInterface.(string)
							if err && metadata == DependsOnMetadataValue {
								IsChildCritical = true
								break
							}
						}
					}
				}
			}
			// check if there an parent who is critical
			if IsChildCritical {
				if ParentFilterdTaskBatch, exists := taskBatchsMap[keyParent]; exists {
					for _, agenda := range ParentFilterdTaskBatch.Agenda {
						if agenda.GetName() == ActionSetValue {
							metadataInterface, err := agenda.GetParameters()[DependsOnMetadata]
							if err {
								metadata, err := metadataInterface.(string)
								if err && metadata == DependsOnMetadataValue {
									// Filter the actions to execute, retaining only those actions that do not adhere to the dependency management.
									// Also, set the situation's action to pending.
									err := filterAgendaAndUpdateHistory(keychild, DependsOnMetadata, filteredTaskBatch, situationHistoryMetadata, situation)
									if err != nil {
										zap.L().Error("Failed to filter agenda and update history", zap.Error(err))
									}

									break
								}

							}
						}
					}
				} else {
					// search for the parent in the database
					Parent, err := history.S().HistorySituationsQuerier.GetLatestHistory(int64(idSituationDependsOn), int64(idInstanceDependsOn))
					if err != nil {
						logDataRetrieval(false, idSituationDependsOn, idInstanceDependsOn, situation.SituationID, situation.SituationInstanceID, err, "")
					} else {
						for _, metadata := range Parent.Metadatas {
							if metadata.Key == DependsOnMetadata && metadata.Value == DependsOnMetadataValue {
								logDataRetrieval(true, idSituationDependsOn, idInstanceDependsOn, situation.SituationID, situation.SituationInstanceID, err, Parent.Ts.String())
								err := filterAgendaAndUpdateHistory(keychild, DependsOnMetadata, filteredTaskBatch, situationHistoryMetadata, situation)
								if err != nil {
									zap.L().Error("Failed to filter agenda and update history", zap.Error(err))
								}
							}
						}
					}

				}
			}

		}
	}

	taskBatchSlice := make([]tasker.TaskBatch, 0, len(filteredTaskBatch))
	for _, taskBatch := range filteredTaskBatch {
		taskBatchSlice = append(taskBatchSlice, taskBatch)
	}

	return taskBatchSlice
}

func filterAgendaAndUpdateHistory(keychild string, DependsOnMetadata string, filteredTaskBatch map[string]tasker.TaskBatch, situationHistoryMetadata map[model.Key]map[string]interface{}, situation history.HistoryRecordV4) error {
	// Filter agenda...
	filteredAgenda := make([]ruleeng.Action, 0)
	for _, action := range filteredTaskBatch[keychild].Agenda {
		if (action.GetEnableDependsForALLAction() == false) || (action.GetEnableDependsForALLAction() == true && action.GetEnabledDependsAction() == false) {
			filteredAgenda = append(filteredAgenda, action)
		}
	}
	taskBatch := filteredTaskBatch[keychild]
	taskBatch.Agenda = filteredAgenda
	filteredTaskBatch[keychild] = taskBatch

	// Update history situation...
	valuesituationHistoryMetadata := situationHistoryMetadata[model.Key{SituationID: situation.SituationID, SituationInstanceID: situation.SituationInstanceID}]
	historySituation := valuesituationHistoryMetadata["HistorySituation"].(history.HistorySituationsV4)
	for i, metadata := range historySituation.Metadatas {
		if metadata.Key == DependsOnMetadata {
			historySituation.Metadatas[i].Value = ActionPendingValue
			break
		}
	}

	return history.S().HistorySituationsQuerier.Update(historySituation)
}

// GetLinkedSituations returns all situation linked to a fact
func GetLinkedSituations(fact engine.Fact) ([]history.HistoryRecordV4, error) {
	factSituationsHistory := make([]history.HistoryRecordV4, 0)
	factSituations, err := situation.R().GetSituationsByFactID(fact.ID, true)
	if err != nil {
		zap.L().Error("Cannot get the situations to update for fact", zap.Int64("id", fact.ID), zap.Any("fact", fact), zap.Error(err))
		return nil, err
	}

	for _, s := range factSituations {
		if !s.IsTemplate {
			factSituationsHistory = append(factSituationsHistory, history.HistoryRecordV4{
				SituationID:         s.ID,
				SituationInstanceID: 0,
				Parameters:          s.Parameters,
			})
		} else {
			templateInstances, err := situation.R().GetAllTemplateInstances(s.ID)
			if err != nil {
				zap.L().Error("Cannot get the situations template instances for situation", zap.Int64("id", s.ID), zap.Any("fact", fact), zap.Error(err))
				return nil, err
			}
			for _, ti := range templateInstances {
				sh := history.HistoryRecordV4{
					SituationID:         s.ID,
					SituationInstanceID: ti.ID,
					Parameters:          map[string]interface{}{},
				}
				sh.OverrideParameters(s.Parameters)
				sh.OverrideParameters(ti.Parameters)
				factSituationsHistory = append(factSituationsHistory, sh)
			}
		}
	}
	return factSituationsHistory, nil
}

// GetEnabledSituations returns all situation linked to a fact
func GetEnabledSituations(fact engine.Fact, t time.Time) ([]history.HistoryRecordV4, error) {
	factSituationsHistory := make([]history.HistoryRecordV4, 0)
	factSituations, err := situation.R().GetSituationsByFactID(fact.ID, true)
	if err != nil {
		zap.L().Error("Cannot get the situations to update for fact", zap.Int64("id", fact.ID), zap.Any("fact", fact), zap.Error(err))
		return nil, err
	}

	// zap.L().Sugar().Info("factID ", fact.ID)
	// zap.L().Sugar().Info("situations ", factSituations)

	for _, s := range factSituations {
		if !s.IsTemplate {
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

		} else {
			templateInstances, err := situation.R().GetAllTemplateInstances(s.ID)
			if err != nil {
				zap.L().Error("Cannot get the situations template instances for situation", zap.Int64("id", s.ID), zap.Any("fact", fact), zap.Error(err))
				return nil, err
			}

			// zap.L().Sugar().Info("instances", templateInstances)
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
						Parameters:          map[string]interface{}{},
						EnableDependsOn:     ti.EnableDependsOn,
						DependsOnParameters: ti.DependsOnParameters,
					}
					sh.OverrideParameters(s.Parameters)
					sh.OverrideParameters(ti.Parameters)
					factSituationsHistory = append(factSituationsHistory, sh)
				} else {
					zap.L().Debug("Situation template not within a valid calendar period, situation id: ", zap.Int64("id", s.ID))
				}
			}
		}
	}
	return factSituationsHistory, nil
}

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
