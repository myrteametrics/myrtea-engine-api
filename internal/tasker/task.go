package tasker

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Task interface of tasks
type Task interface {
	String() string
	GetID() string
	Perform(key string, input ContextData) error
}

// ApplyTasks applies the task of an evaluated situation.
// pending accumulates the members of any "grouped" situation-reporting task encountered in this
// batch (and, across calls from ApplyBatchs, in every other batch of the same scheduler run), so
// they can be merged into a single email once the whole run has been processed.
func ApplyTasks(batch TaskBatch, pending map[string]*groupedReportAcc) (err error) {

	for _, action := range batch.Agenda {

		switch action.GetName() {

		case ActionCreateIssue:
			task, err := buildCreateIssueTask(action.GetParameters(), batch.JobBoostInfo)
			if err != nil {
				zap.L().Warn("Error building CreateIssueTask: ", zap.Any("Parameters:", action.GetParameters()), zap.Error(err))
				continue
			}

			taskContext := BuildContextData(action.GetMetaData(), batch.Context)
			err = task.Perform(buildTaskKey(taskContext, task), taskContext)
			if err != nil {
				zap.L().Warn("Error while performing task CreateIssueTask", zap.Error(err))
			}

		case ActionCloseTodayIssues:
			task, err := buildCloseTodayIssuesTask(action.GetParameters())
			if err != nil {
				zap.L().Warn("Error building CloseTodayIssuesTask: ", zap.Any("Parameters:", action.GetParameters()), zap.Error(err))
				continue
			}

			taskContext := BuildContextData(action.GetMetaData(), batch.Context)
			err = task.Perform(buildTaskKey(taskContext, task), taskContext)
			if err != nil {
				zap.L().Warn("Error while performing task CloseTodayIssuesTask", zap.Error(err))
			}

		case ActionCloseAllIssues:
			task, err := buildCloseAllIssuesTask(action.GetParameters())
			if err != nil {
				zap.L().Warn("Error building CloseAllIssuesTask: ", zap.Any("Parameters:", action.GetParameters()), zap.Error(err))
				continue
			}

			taskContext := BuildContextData(action.GetMetaData(), batch.Context)
			err = task.Perform(buildTaskKey(taskContext, task), taskContext)
			if err != nil {
				zap.L().Warn("Error while performing task CloseAllIssuesTask", zap.Error(err))
			}

		case ActionNotify:
			task, err := buildNotifyTask(action.GetParameters())
			if err != nil {
				zap.L().Warn("Error building NotifyTask: ", zap.Any("Parameters:", action.GetParameters()), zap.Error(err))
				continue
			}

			taskContext := BuildContextData(action.GetMetaData(), batch.Context)
			err = task.Perform(buildTaskKey(taskContext, task), taskContext)
			if err != nil {
				zap.L().Warn("Error while performing task NotifyTask", zap.Error(err))
			}

		case ActionSituationReporting:
			task, err := buildSituationReportingTask(action.GetParameters())
			if err != nil {
				zap.L().Warn("Error building SituationReportingTask: ", zap.Any("Parameters:", action.GetParameters()), zap.Error(err))
				continue
			}

			taskContext := BuildContextData(action.GetMetaData(), batch.Context)

			if task.GroupBySituation {
				// Don't send yet: stash this instance's rendered report so it can be merged with
				// the other instances of the same situation once the whole run is done.
				if err := collectGroupedSituationReporting(pending, task, taskContext); err != nil {
					zap.L().Warn("Error while collecting grouped SituationReportingTask", zap.Error(err))
				}
				continue
			}

			err = task.Perform(buildSituationReportingKey(taskContext, task), taskContext)
			if err != nil {
				zap.L().Warn("Error while performing task SituationReportingTask", zap.Error(err))
			}

		default:
			continue
		}
	}

	return nil
}

// buildTaskKey builds the shared dedup/lookup key used by CreateIssueTask and the Close*IssuesTask
// family. It is intentionally scoped to (situation, rule, task) and NOT to the template instance:
// CreateIssueTask relies on this to avoid duplicating an already-open issue, and CloseTodayIssuesTask/
// CloseAllIssuesTask use it to bulk-close every issue sharing that key, across all instances.
// Do not add TemplateInstanceID here without checking those three call sites.
func buildTaskKey(input ContextData, task Task) string {
	key := fmt.Sprintf("%d-%d-%s", input.SituationID, input.RuleID, task.GetID())
	return key
}

// buildSituationReportingKey builds the per-instance dedup/timeout key used by the (non-grouped)
// SituationReportingTask. Unlike buildTaskKey, it includes the template instance id: without it,
// every instance of a template situation shared the same timeout cache entry, so only the first
// instance to fire within the timeout window actually sent its email - the others were silently
// dropped instead of just being deduplicated.
func buildSituationReportingKey(input ContextData, task Task) string {
	return fmt.Sprintf("%d-%d-%d-%s", input.SituationID, input.TemplateInstanceID, input.RuleID, task.GetID())
}

// ApplyBatchs applies the tasks batchs produced by a single scheduler run, then flushes any
// grouped situation-reporting alert accumulated along the way into a single email per group.
func ApplyBatchs(batchs []TaskBatch) {
	pending := make(map[string]*groupedReportAcc)

	for _, batch := range batchs {
		err := ApplyTasks(batch, pending)
		if err != nil {
			zap.L().Error("ApplyBatch error on evaluated Situation: ", zap.Any("Context:", batch.Context), zap.String(" at", time.Now().String()))
			continue
		}
	}

	flushGroupedSituationReporting(pending)
}
