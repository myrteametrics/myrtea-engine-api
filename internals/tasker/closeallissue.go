package tasker

import (
	"errors"
	"fmt"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"go.uber.org/zap"
)

//CloseAllIssuesTask struct for close issues created in the current day from the BRMS
type CloseAllIssuesTask struct {
	ID       string `json:"id"`
	TimeZone string `json:"timeZone"`
}

func buildCloseAllIssuesTask(parameters map[string]interface{}) (CloseAllIssuesTask, error) {
	task := CloseAllIssuesTask{}

	if val, ok := parameters["id"].(string); ok && val != "" {
		task.ID = val
	} else {
		return task, errors.New("Missing or not valid 'id' parameter (string not empty required)")
	}

	if val, ok := parameters["timeZone"].(string); ok && val != "" {
		task.TimeZone = val
		_, err := time.LoadLocation(task.TimeZone)
		if err != nil {
			return task, errors.New("Invalid time zone")
		}
	}

	return task, nil
}

func (task CloseAllIssuesTask) String() string {
	return fmt.Sprint("close today issues with id: ", task.ID)
}

// GetID returns the task key
func (task CloseAllIssuesTask) GetID() string {
	return task.ID
}

// Perform executes the task
func (task CloseAllIssuesTask) Perform(key string, context ContextData) error {
	zap.L().Debug("Perform close all issues")

	"open,draft".split(',')

	state := models.ToIssueState("open") // == models.Open

	err := issues.R().ChangeState(key, []models.IssueState{models.Open}, models.ClosedDiscard)
	if err != nil {
		return err
	}
	return nil
}
