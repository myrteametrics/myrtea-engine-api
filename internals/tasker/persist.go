package tasker

import (
	"errors"
	"fmt"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
)

// PersistDataTask struct to represent a persit metadata task (use to persist situation instance metadata)
type PersistDataTask struct {
	Data []models.MetaData
}

func (task PersistDataTask) String() string {
	return fmt.Sprint("Persist data ", fmt.Sprint(task.Data))
}

// GetID returns the task key
func (task PersistDataTask) GetID() string {
	return ""
}

// Perform executes the task
func (task PersistDataTask) Perform(key string, context ContextData) error {

	if context.SituationID == 0 || context.TS.IsZero() {
		return errors.New("errro performinf PersistDataTask, situation instance not defined")
	}

	err := situation.UpdateHistoryMetadata(int64(context.SituationID), context.TS, int64(context.TemplateInstanceID), task.Data)
	if err != nil {
		return errors.New("error when performing PersistDataTask: " + err.Error())
	}
	return nil
}

func (task *PersistDataTask) addData(parameters map[string]interface{}, context ContextData) {

	for key, value := range parameters {
		task.Data = append(task.Data, models.MetaData{
			Key:         key,
			Value:       value,
			RuleID:      context.RuleID,
			RuleVersion: context.RuleVersion,
			CaseName:    context.CaseName,
		})
	}
}
