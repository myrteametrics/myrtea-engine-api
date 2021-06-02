package tasker

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

//Task interface of tasks
type Task interface {
	String() string
	GetID() string
	Perform(key string, input ContextData) error
}

//ApplyTasks applies the task of an evaluated situation
func ApplyTasks(batch TaskBatch) (err error) {

	persistTask := PersistDataTask{}
	for _, action := range batch.Agenda {

		switch action.GetName() {
		case "set":
			persistTask.addData(action.GetParameters(), buildContextData(action.GetMetaData(), batch.Context))
		case "create-issue":
			task, err := buildCreateIssueTask(action.GetParameters())
			if err != nil {
				zap.L().Warn("Error building CreateIssueTask: ", zap.Any("Parameters:", action.GetParameters()), zap.Error(err))
				continue
			}
			taskContext := buildContextData(action.GetMetaData(), batch.Context)
			task.Perform(buildTaskKey(taskContext, task), taskContext)
		case "close-today-issues":
			task, err := buildCloseTodayIssuesTask(action.GetParameters())
			if err != nil {
				zap.L().Warn("Error building CloseTodayIssuesTask: ", zap.Any("Parameters:", action.GetParameters()), zap.Error(err))
				continue
			}
			taskContext := buildContextData(action.GetMetaData(), batch.Context)
			task.Perform(buildTaskKey(taskContext, task), taskContext)
		case "notify":
			task, err := buildNotifyTask(action.GetParameters())
			if err != nil {
				zap.L().Warn("Error building NotifyTask: ", zap.Any("Parameters:", action.GetParameters()), zap.Error(err))
				continue
			}
			taskContext := buildContextData(action.GetMetaData(), batch.Context)
			task.Perform(buildTaskKey(taskContext, task), taskContext)
		case "situation-reporting":
			task, err := buildSituationReportingTask(action.GetParameters())
			if err != nil {
				zap.L().Warn("Error building SituationReportingTask: ", zap.Any("Parameters:", action.GetParameters()), zap.Error(err))
				continue
			}
			taskContext := buildContextData(action.GetMetaData(), batch.Context)
			task.Perform(buildTaskKey(taskContext, task), taskContext)
		default:
			continue
		}
	}
	persistTask.Perform("", buildContextData(batch.Context))

	return nil
}

func buildTaskKey(input ContextData, task Task) string {
	key := fmt.Sprintf("%d-%d-%s", input.SituationID, input.RuleID, task.GetID())
	return key
}

//ApplyBatchs applies the tasks batchs
func ApplyBatchs(batchs []TaskBatch) {
	for _, batch := range batchs {
		err := ApplyTasks(batch)
		if err != nil {
			zap.L().Error("ApplyBatch error on evaluated Situation: ", zap.Any("Context:", batch.Context), zap.String(" at", time.Now().String()))
			continue
		}
	}
}
