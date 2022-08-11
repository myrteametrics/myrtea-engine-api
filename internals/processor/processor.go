package processor

import (
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/evaluator"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tasker"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"go.uber.org/zap"
)

func evaluateFactObjects(fact engine.Fact, objects []map[string]interface{}) error {

	situations, err := situation.R().GetSituationsByFactID(fact.ID, false)
	if err != nil {
		return err
	}
	if len(situations) == 0 {
		zap.L().Debug("No situation to evaluate for fact object", zap.String("factName", fact.Name))
		return nil
	}
	zap.L().Debug("situations to evaluate", zap.Any("situations", situations))

	situationsToEvalute := make([]evaluator.SituationToEvaluate, 0)
	ts := time.Now().Truncate(1 * time.Millisecond).UTC()
	for _, objectSituation := range situations {

		situationsToEvalute = append(situationsToEvalute,
			evaluator.SituationToEvaluate{
				ID:         objectSituation.ID,
				TS:         ts,
				Facts:      objectSituation.Facts,
				Parameters: objectSituation.Parameters,
			},
		)
	}

	localEvaluator, err := evaluator.BuildLocalRuleEngine("object")
	// Evaluate rules
	enabledRuleIDs, err := GetEnabledRuleIDs(situationToUpdate.SituationID, situationToUpdate.Ts)
	if err != nil {
		zap.L().Error("", zap.Error(err))
	}

	metadatas := make([]models.MetaData, 0)
	agenda := evaluator.EvaluateRules(localRuleEngine, historySituationFlattenData, enabledRuleIDs)

	// FIXME: Fixing situation object evaluation
	// situationEvaluations, err := evaluator.EvaluateObjectSituations(situationsToEvalute, fact, objects, "object")

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
	if err == nil {
		tasker.T().BatchReceiver <- taskBatchs
	}

	return nil
}
