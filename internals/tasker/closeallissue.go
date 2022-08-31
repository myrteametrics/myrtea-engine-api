package tasker

import (
	"errors"
	"fmt"
	"strings"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"go.uber.org/zap"
)

//CloseAllIssuesTask struct for close issues created from the BRMS
type CloseAllIssuesTask struct {
	ID            string `json:"id"`
	StatesToClose string `json:"statesToClose"`
}

func buildCloseAllIssuesTask(parameters map[string]interface{}) (CloseAllIssuesTask, error) {
	task := CloseAllIssuesTask{}

	if val, ok := parameters["id"].(string); ok && val != "" {
		task.ID = val
	} else {
		return task, errors.New("missing or not valid 'id' parameter (string not empty required)")
	}

	if val, ok := parameters["statesToClose"].(string); ok && val != "" {
		task.StatesToClose = val
	} else {
		return task, errors.New("missing or not valid 'statesToClose' parameter (at least 1 value (open, draft) required)")
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

	var states []models.IssueState

	statesToClose := strings.Split(task.StatesToClose, ",")

	for _, s := range statesToClose {
		states = append(states, models.ToIssueState(s))
	}

	err := issues.R().ChangeState(key, states, models.ClosedDiscard)
	if err != nil {
		return err
	}
	return nil
}
