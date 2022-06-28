package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/evaluator"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tasker"
	"github.com/myrteametrics/myrtea-engine-api/v4/plugins/baseline"
	"github.com/myrteametrics/myrtea-sdk/v4/builder"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"github.com/myrteametrics/myrtea-sdk/v4/ruleeng"
	"go.uber.org/zap"
)

const timeLayout = "2006-01-02T15:04:05.000Z07:00"

// FactCalculationJob represent a scheduler job instance which process a group of facts,
// and persist the result in postgresql
// It also generate situations, persists them and notify the rule engine to evaluate them
type FactCalculationJob struct {
	FactIds        []int64 `json:"factIds"`
	From           string  `json:"from,omitempty"`
	To             string  `json:"to,omitempty"`
	LastDailyValue bool    `json:"lastDailyValue,omitempty"`
	Debug          bool    `json:"debug"`
	ScheduleID     int64   `json:"-"`
}

// NewFactCalculationJob returns a new isntance of FactCalculationJob
func NewFactCalculationJob(cronExpr string, factIds []int64) FactCalculationJob {
	return FactCalculationJob{
		FactIds: factIds,
	}
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
	if !S().ExistingRunningJob(job.ScheduleID) {
		S().AddRunningJob(job.ScheduleID)

		t := time.Now().Truncate(1 * time.Second).UTC()
		//Update the calendar base with last modifications
		calendar.CBase().Update()

		zap.L().Info("Fact calculation job started", zap.Int64s("ids", job.FactIds))

		if job.From != "" {
			err := job.update(t)
			if err != nil {
				zap.L().Error("Error updating fact instances", zap.Error(err))
			}
			zap.L().Info("FactScheduleJob Ended", zap.Int64s("ids", job.FactIds))
			S().RemoveRunningJob(job.ScheduleID)
			return
		}

		if fact.R() == nil {
			zap.L().Error("Fact Repository is not initialized")
			S().RemoveRunningJob(job.ScheduleID)
			return
		}
		situationsToUpdate := make(map[string]situation.HistoryRecord, 0)
		for _, factID := range job.FactIds {
			f, found, err := fact.R().Get(factID)
			if err != nil {
				zap.L().Error("Error Getting the Fact, skipping fact calculation...", zap.Int64("factID", factID))
				continue
			}
			if !found {
				zap.L().Warn("Fact does not exists, skipping fact calculation...", zap.Int64("factID", factID))
				continue
			}
			if job.Debug {
				zap.L().Debug("Debugging fact", zap.Any("f", f))
			}
			factSituationsHistory, err := GetFactSituations(f, t)
			if err != nil {
				continue
			}

			//If no situation within a valid calendar, then no fact calculation at all
			if len(factSituationsHistory) == 0 {
				zap.L().Info("No situation within valid calendar period for the Fact, skipping fact calculation...", zap.Int64("factID", factID))
				S().RemoveRunningJob(job.ScheduleID)
				return
			}

			// is fact template ??
			if f.IsTemplate {
				for _, sh := range factSituationsHistory {

					var fCopy engine.Fact
					fData, _ := json.Marshal(f)
					json.Unmarshal(fData, &fCopy)
					err = job.calculate(t, fCopy, sh.ID, sh.TemplateInstanceID, sh.Parameters, false)
					if err != nil {
						zap.L().Error("Fact calculation Error, skipping fact calculation...", zap.Int64("id", f.ID), zap.Any("fact", f), zap.Error(err))
						continue
					}
					key := fmt.Sprintf("%d-%d", sh.ID, sh.TemplateInstanceID)
					if _, ok := situationsToUpdate[key]; !ok {
						situationsToUpdate[key] = situation.HistoryRecord{
							ID:                 sh.ID,
							TS:                 t,
							FactsIDS:           map[int64]*time.Time{f.ID: &t},
							Parameters:         sh.Parameters,
							TemplateInstanceID: sh.TemplateInstanceID,
						}
					} else {
						situationsToUpdate[key].FactsIDS[f.ID] = &t
					}
				}
			} else {
				err = job.calculate(t, f, 0, 0, nil, false)
				if err != nil {
					zap.L().Error("Fact calculation Error, skipping fact calculation...", zap.Int64("id", f.ID), zap.Any("fact", f), zap.Error(err))
					continue
				}
				for _, sh := range factSituationsHistory {
					key := fmt.Sprintf("%d-%d", sh.ID, sh.TemplateInstanceID)
					if _, ok := situationsToUpdate[key]; !ok {
						situationsToUpdate[key] = situation.HistoryRecord{
							ID:                 sh.ID,
							TS:                 t,
							FactsIDS:           map[int64]*time.Time{f.ID: &t},
							Parameters:         sh.Parameters,
							TemplateInstanceID: sh.TemplateInstanceID,
						}
					} else {
						situationsToUpdate[key].FactsIDS[f.ID] = &t
					}
				}
			}
		}

		situationsToEvaluate, err := UpdateSituations(situationsToUpdate)
		if err != nil {
			zap.L().Error("Cannot update situations", zap.Error(err))
			S().RemoveRunningJob(job.ScheduleID)
			return
		}

		situationEvaluations, err := evaluator.EvaluateSituations(situationsToEvaluate, "standart")
		if err == nil {
			taskBatchs := make([]tasker.TaskBatch, 0)
			for _, situationEvaluation := range situationEvaluations {
				taskBatchs = append(taskBatchs, tasker.TaskBatch{
					Context: map[string]interface{}{
						"situationID":        situationEvaluation.ID,
						"ts":                 situationEvaluation.TS,
						"templateInstanceID": situationEvaluation.TemplateInstanceID,
					},
					Agenda: situationEvaluation.Agenda,
				})
			}

			tasker.T().BatchReceiver <- taskBatchs
		}

		zap.L().Info("FactScheduleJob Ended", zap.Int64s("ids", job.FactIds))
		S().RemoveRunningJob(job.ScheduleID)
	} else {
		zap.L().Info("Skepping FactScheduleJob because last execution is still running", zap.Int64s("ids", job.FactIds))
	}
}

func (job FactCalculationJob) update(t time.Time) error {
	from, to, err := job.ResolveFromAndTo(t)
	if err != nil {
		return err
	}
	instances, err := fact.GetFactSituationInstances(job.FactIds, from, to, job.LastDailyValue)
	if err != nil {
		zap.L().Error("Error Getting fact instances to update", zap.Error(err))
		return err
	}

	situationsToEvaluate := make([]evaluator.SituationToEvaluate, 0)
	situationKeys := make(map[string]int)

	for _, instance := range instances {
		var parameters map[string]string
		if instance.SituationID == 0 {
			parameters = make(map[string]string, 0)
		} else {
			parameters = instance.SituationInstances[0].Parameters
		}
		err = job.calculate(instance.TS, instance.Fact, instance.SituationID, instance.TemplateInstanceID, parameters, true)
		if err != nil {
			zap.L().Error("Fact calculation Error, skipping fact calculation...", zap.Int64("id", instance.ID), zap.Any("fact", instance.Fact), zap.Error(err))
			continue
		}

		for _, si := range instance.SituationInstances {
			key := fmt.Sprintf("%d-%s-%d", si.ID, si.TS, si.TemplateInstanceID)
			if _, ok := situationKeys[key]; !ok {
				situationKeys[key] = 0
				situationsToEvaluate = append(situationsToEvaluate, evaluator.SituationToEvaluate{
					ID:                 si.ID,
					TS:                 si.TS,
					TemplateInstanceID: si.TemplateInstanceID,
				})
			}
		}
	}

	for _, s := range situationsToEvaluate {
		record, err := situation.GetFromHistory(s.ID, s.TS, s.TemplateInstanceID, false)
		if err != nil {
			zap.L().Error("Get situation from history", zap.Int64("situationID", record.ID), zap.Time("ts", record.TS), zap.Error(err))
			continue
		}

		evaluatedExpressionFacts, err := evaluateExpressionFacts(*record, s.TS)
		if err != nil {
			zap.L().Error("cannot evaluate expression facts", zap.Error(err))
			continue
		}
		record.EvaluatedExpressionFacts = evaluatedExpressionFacts

		situation.UpdateExpressionFacts(*record)
	}

	situationEvaluations, err := evaluator.EvaluateSituations(situationsToEvaluate, "standart")
	if err == nil {
		//Filter situationEvaluations (keep persist task only)
		taskBatchs := make([]tasker.TaskBatch, 0)
		for _, situationEvaluation := range situationEvaluations {

			//Filter situationEvaluations (keep persist task only)
			agenda := make([]ruleeng.Action, 0)
			for _, action := range situationEvaluation.Agenda {
				if action.GetName() == "set" {
					agenda = append(agenda, action)
				}
			}

			taskBatchs = append(taskBatchs, tasker.TaskBatch{
				Context: map[string]interface{}{
					"situationID":        situationEvaluation.ID,
					"ts":                 situationEvaluation.TS,
					"templateInstanceID": situationEvaluation.TemplateInstanceID,
				},
				Agenda: agenda,
			})
		}

		tasker.T().BatchReceiver <- taskBatchs
	}

	return nil
}

func (job FactCalculationJob) calculate(t time.Time, f engine.Fact, situationID int64, templateInstanceID int64, placeholders map[string]string, update bool) error {
	pf, err := fact.Prepare(&f, -1, -1, t, placeholders, update)
	if err != nil {
		zap.L().Error("Cannot prepare fact", zap.Int64("id", f.ID), zap.Any("fact", f), zap.Error(err))
		return err
	}
	if job.Debug {
		zap.L().Debug("Debugging prepared fact", zap.Any("pf", pf))
		source, _ := builder.BuildEsSearchSource(pf)
		zap.L().Debug("Debugging final elastic query", zap.Any("query", source))
	}
	widgetData, err := fact.Execute(pf)
	if err != nil {
		zap.L().Error("Cannot execute fact", zap.Int64("id", f.ID), zap.Any("pf", pf), zap.Error(err))
		return err
	}

	pluginBaseline, err := baseline.P()
	if err == nil {
		values, err := pluginBaseline.Baseline.GetBaselineValues(-1, f.ID, situationID, templateInstanceID, t)
		if err != nil {
			zap.L().Error("Cannot fetch fact baselines", zap.Int64("id", f.ID), zap.Error(err))
		}
		widgetData.Aggregates.Baselines = values
	}

	if update {
		err = fact.UpdateFactResult(f.ID, t, situationID, templateInstanceID, widgetData.Aggregates)
		if err != nil {
			zap.L().Error("Cannot update fact instance", zap.Error(err))
			return err
		}
	} else {
		err = fact.PersistFactResult(f.ID, t, situationID, templateInstanceID, widgetData.Aggregates, true)
		if err != nil {
			zap.L().Error("Cannot persist fact instance", zap.Error(err))
			return err
		}
	}
	return nil
}

// UpdateSituations creates the new instances of the situations in the history and evaluates them
func UpdateSituations(situationsToUpdate map[string]situation.HistoryRecord) ([]evaluator.SituationToEvaluate, error) {

	situationsToEvalute := make([]evaluator.SituationToEvaluate, 0)
	for _, record := range situationsToUpdate {

		// create factsHistory from situationFacts
		situationFacts, err := situation.R().GetFacts(record.ID)
		if err != nil {
			zap.L().Error("Get situation facts", zap.Int64("situationID", record.ID), zap.Error(err))
			continue
		}

		// merge values from lastHistoryRecord into factsHistory
		lastHistoryRecord, err := situation.GetFromHistory(record.ID, record.TS, record.TemplateInstanceID, true)
		if err != nil {
			zap.L().Error("Get situation from history", zap.Int64("situationID", record.ID), zap.Time("ts", record.TS), zap.Error(err))
			continue
		}

		factsHistory := make(map[int64]*time.Time)
		for _, factID := range situationFacts {
			factsHistory[factID] = nil
		}

		if lastHistoryRecord != nil {
			for factID, factTS := range lastHistoryRecord.FactsIDS {
				factsHistory[factID] = factTS
			}
		}
		// merge new values into factsHistory
		for factID, factTS := range record.FactsIDS {
			factsHistory[factID] = factTS
		}

		record.FactsIDS = factsHistory

		evaluatedExpressionFacts, err := evaluateExpressionFacts(record, record.TS)
		if err != nil {
			zap.L().Warn("cannot evaluate expression facts", zap.Error(err))
			continue
		}
		record.EvaluatedExpressionFacts = evaluatedExpressionFacts

		err = situation.Persist(record, false)
		if err != nil {
			zap.L().Error("UpdateSituations.persistSituation:", zap.Error(err))
			continue
		}
		situationsToEvalute = append(situationsToEvalute,
			evaluator.SituationToEvaluate{
				ID:                 record.ID,
				TS:                 record.TS,
				TemplateInstanceID: record.TemplateInstanceID,
			},
		)
	}

	return situationsToEvalute, nil
}

func evaluateExpressionFacts(record situation.HistoryRecord, t time.Time) (map[string]interface{}, error) {
	evaluatedExpressionFacts := make(map[string]interface{})

	s, found, err := situation.R().Get(record.ID)
	if err != nil {
		zap.L().Error("Get Situation", zap.Int64("situationID", record.ID), zap.Error(err))
		return evaluatedExpressionFacts, err
	}
	if !found {
		zap.L().Warn("Situation not found", zap.Int64("situationID", s.ID))
		return evaluatedExpressionFacts, fmt.Errorf("situation not found with ID = %d", record.ID)
	}

	data, err := flattenSituationData(record)
	if err != nil {
		return evaluatedExpressionFacts, err
	}

	//Add date keywords in situation data
	for key, value := range expression.GetDateKeywords(t) {
		data[key] = value
	}

	for _, expressionFact := range s.ExpressionFacts {
		result, err := expression.Process(expression.LangEval, expressionFact.Expression, data)
		if err != nil {
			zap.L().Debug("Cannot process gval factExpression", zap.Error(err))
			continue
		}
		if expression.IsInvalidNumber(result) {
			continue
		}

		data[expressionFact.Name] = result
		evaluatedExpressionFacts[expressionFact.Name] = result
	}
	return evaluatedExpressionFacts, nil
}

func flattenSituationData(record situation.HistoryRecord) (map[string]interface{}, error) {
	situationData := make(map[string]interface{})
	for factID, factTS := range record.FactsIDS {
		f, found, err := fact.R().Get(factID)
		if err != nil {
			zap.L().Error("get fact", zap.Error(err))
			return nil, err
		}
		if !found {
			zap.L().Warn("fact not found", zap.Int64("factID", factID))
			return nil, fmt.Errorf("fact not found with id=%d", factID)
		}
		if factTS == nil {
			// zap.L().Warn("At least one fact has never been calculated", zap.Int64("id", f.ID), zap.String("name", f.Name))
			// return nil, fmt.Errorf("at least one fact has never been calculated, id=%d, name=%s", f.ID, f.Name)
			continue
		}

		item, _, err := fact.GetFactResultFromHistory(factID, *factTS, record.ID, record.TemplateInstanceID, false, -1)
		if err != nil {
			return nil, err
		}
		itemData, err := item.ToAbstractMap()
		if err != nil {
			zap.L().Error("Convert item to abstractmap", zap.Error(err))
			return nil, err
		}

		situationData[f.Name] = itemData
	}
	for key, value := range record.Parameters {
		situationData[key] = value
	}

	return situationData, nil
}

// GetFactSituations returns all situation linked to a fact
func GetFactSituations(fact engine.Fact, t time.Time) ([]situation.HistoryRecord, error) {
	factSituationsHistory := make([]situation.HistoryRecord, 0)
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
					sh := situation.HistoryRecord{
						ID:                 s.ID,
						Parameters:         map[string]string{},
						TemplateInstanceID: ti.ID,
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
				factSituationsHistory = append(factSituationsHistory, situation.HistoryRecord{
					ID:         s.ID,
					Parameters: s.Parameters,
				})
			} else {
				zap.L().Debug("Situation not within a valid calendar period, situation id: ", zap.Int64("id", s.ID))
			}
		}
	}
	return factSituationsHistory, nil
}
