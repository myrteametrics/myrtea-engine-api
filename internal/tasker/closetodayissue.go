package tasker

import (
	"errors"
	"fmt"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
	"go.uber.org/zap"
)

// CloseTodayIssuesTask struct for close issues created in the current day from the BRMS
type CloseTodayIssuesTask struct {
	ID       string `json:"id"`
	Timezone string `json:"timezone"`
}

func buildCloseTodayIssuesTask(parameters map[string]interface{}) (CloseTodayIssuesTask, error) {
	task := CloseTodayIssuesTask{}

	if val, ok := parameters["id"].(string); ok && val != "" {
		task.ID = val
	} else {
		return task, errors.New("missing or not valid 'id' parameter (string not empty required)")
	}

	if val, ok := parameters["timezone"].(string); ok && val != "" {
		task.Timezone = val
		_, err := time.LoadLocation(task.Timezone)
		if err != nil {
			return task, errors.New("invalid time zone")
		}
	}

	return task, nil
}

func (task CloseTodayIssuesTask) String() string {
	return fmt.Sprint("close today issues with id: ", task.ID)
}

// GetID returns the task key
func (task CloseTodayIssuesTask) GetID() string {
	return task.ID
}

// Perform executes the task
func (task CloseTodayIssuesTask) Perform(key string, context ContextData) error {
	zap.L().Debug("Perform close today issues")

	now := time.Now()
	loc, err := time.LoadLocation(task.Timezone)
	if err == nil {
		now = now.In(loc)
	} else {
		now = now.UTC()
	}

	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	to := from.Add(24 * time.Hour)

	err = issues.R().ChangeStateBetweenDates(key, []models.IssueState{models.Open}, models.ClosedDiscard, from, to)

	if err != nil {
		return err
	}
	return nil
}
