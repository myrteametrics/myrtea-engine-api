package processor

import (
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/evaluator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/history"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/rule"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tasker"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"go.uber.org/zap"
)

func evaluateFactObjects(factObject engine.Fact, objects []map[string]interface{}) error {

	t := time.Now()

	situations, err := situation.R().GetSituationsByFactID(factObject.ID, false)
	if err != nil {
		return err
	}
	if len(situations) == 0 {
		zap.L().Debug("No situation to evaluate for fact object", zap.String("factName", factObject.Name))
		return nil
	}

	localRuleEngine, err := evaluator.BuildLocalRuleEngine("object")
	if err != nil {
		zap.L().Error("", zap.Error(err))
	}

	taskBatchs := make([]tasker.TaskBatch, 0)
	for _, s := range situations {

		// _, parameters, err := history.ExtractSituationData(s.ID, 0)
		// if err != nil {
		// 	zap.L().Error("", zap.Error(err))
		// 	continue
		// }

		historyFactsAll, historySituationFlattenData, err := history.S().ExtractFactData(s.ID, 0, make([]history.HistoryFactsV4, 0), s.Facts)
		if err != nil {
			zap.L().Error("", zap.Error(err))
			continue
		}
		for key, value := range s.Parameters {
			historySituationFlattenData[key] = value
		}
		for key, value := range expression.GetDateKeywords(t) {
			historySituationFlattenData[key] = value
		}

		expressionFacts := history.EvaluateExpressionFacts(s.ExpressionFacts, historySituationFlattenData)
		for key, value := range expressionFacts {
			historySituationFlattenData[key] = value
		}

		enabledRuleIDs, err := rule.R().GetEnabledRuleIDs(s.ID, t)
		if err != nil {
			zap.L().Error("", zap.Error(err))
		}

		for _, object := range objects {
			localRuleEngine.Reset()
			localRuleEngine.GetKnowledgeBase().SetFacts(historySituationFlattenData)
			localRuleEngine.GetKnowledgeBase().InsertFact(factObject.Name, object)
			localRuleEngine.ExecuteRules(enabledRuleIDs)
			agenda := localRuleEngine.GetResults()

			if len(agenda) > 0 {

				objectKeyValues := make(map[string]*reader.ItemAgg, 0)
				for k, v := range object {
					objectKeyValues[k] = &reader.ItemAgg{Value: v}
				}
				historyFactNew := history.HistoryFactsV4{
					// ID:                  -1,
					SituationID:         s.ID,
					SituationInstanceID: 0,
					FactID:              factObject.ID,
					Ts:                  t,
					Result: reader.Item{
						Key:  object["id"].(string),
						Aggs: objectKeyValues,
					},
				}
				historyFactNew.ID, err = history.S().HistoryFactsQuerier.Insert(historyFactNew)
				if err != nil {
					zap.L().Error("", zap.Error(err))
				}

				historySituationNew := history.HistorySituationsV4{
					// ID:                    -1,
					SituationID:         s.ID,
					SituationInstanceID: 0,
					Ts:                  t,
					Parameters:          s.Parameters,
					ExpressionFacts:     expressionFacts,
					Metadatas:           make([]models.MetaData, 0),
				}
				historySituationNew.ID, err = history.S().HistorySituationsQuerier.Insert(historySituationNew)
				if err != nil {
					zap.L().Error("", zap.Error(err))
				}

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

				taskBatchs = append(taskBatchs, tasker.TaskBatch{
					Context: map[string]interface{}{
						"situationID":        s.ID,
						"templateInstanceID": 0,
						"ts":                 t,
					},
					Agenda: agenda,
				})
			}
		}

	}

	tasker.T().BatchReceiver <- taskBatchs

	return nil
}
