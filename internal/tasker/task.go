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

// ApplyTasks applies the task of an evaluated situation
func ApplyTasks(batch TaskBatch) (err error) {

	for _, action := range batch.Agenda {

		switch action.GetName() {

		case ActionCreateIssue:
			task, err := buildCreateIssueTask(action.GetParameters())
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
			err = task.Perform(buildTaskKey(taskContext, task), taskContext)
			if err != nil {
				zap.L().Warn("Error while performing task SituationReportingTask", zap.Error(err))
			}

		default:
			continue
		}
	}

	return nil
}

func buildTaskKey(input ContextData, task Task) string {
	key := fmt.Sprintf("%d-%d-%s", input.SituationID, input.RuleID, task.GetID())
	return key
}

// ApplyBatchs applies the tasks batchs
func ApplyBatchs(batchs []TaskBatch) {
	for _, batch := range batchs {
		err := ApplyTasks(batch)
		if err != nil {
			zap.L().Error("ApplyBatch error on evaluated Situation: ", zap.Any("Context:", batch.Context), zap.String(" at", time.Now().String()))
			continue
		}
	}
}
