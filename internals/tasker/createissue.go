package tasker

import (
	"errors"
	"fmt"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"go.uber.org/zap"
)

//CreateIssueTask struct for the creation of an issue from the BRMS
type CreateIssueTask struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Level          string `json:"level"`
	Timeout        string `json:"timeout"`
	IsNotification bool   `json:"isNotification"`
}

func buildCreateIssueTask(parameters map[string]interface{}) (CreateIssueTask, error) {
	task := CreateIssueTask{}

	if val, ok := parameters["id"].(string); ok && val != "" {
		task.ID = val
	} else {
		return task, errors.New("missing or not valid 'id' parameter (string not empty required)")
	}

	if val, ok := parameters["name"].(string); ok && val != "" {
		task.Name = val
	} else {
		return task, errors.New("missing or not valid 'name' parameter (string not empty required)")
	}

	if val, ok := parameters["level"].(string); ok && val != "" {
		task.Level = val
	} else {
		return task, errors.New("missing or not valid 'level' parameter (string not empty required)")
	}

	if val, ok := parameters["timeout"].(string); ok && val != "" {
		task.Timeout = val
	} else {
		return task, errors.New("missing or not valid 'timeout' parameter (string not empty required)")
	}

	if val, ok := parameters["isNotification"].(bool); ok {
		task.IsNotification = val
	} else {
		return task, errors.New("missing or not valid 'isNotification' parameter (boolean required)")
	}

	return task, nil
}

func (task CreateIssueTask) String() string {
	return fmt.Sprint("Create issue: ", fmt.Sprint(task.Name))
}

// GetID returns the task key
func (task CreateIssueTask) GetID() string {
	return task.ID
}

// Perform executes the task
func (task CreateIssueTask) Perform(key string, context ContextData) error {
	zap.L().Debug("Perform CreateIssueTask")

	//Parsing the timeout from string to duration
	timeoutDuration, err := time.ParseDuration(task.Timeout)
	if err != nil {
		return err
	}

	issueLevel := models.ToIssueLevel(task.Level)

	issueID, err := explainer.CreateIssue(int64(context.SituationID), context.TS, int64(context.TemplateInstanceID),
		models.RuleData{
			RuleID:      int64(context.RuleID),
			RuleVersion: int64(context.RuleVersion),
			CaseName:    context.CaseName,
		},
		task.Name, issueLevel, timeoutDuration, key)
	if err != nil {
		return err
	}

	_ = issueID
	// TODO: find another way to send notification to a specific "population" after permission system refactoring
	// if issueID > 0 && task.IsNotification {
	// 	ctx := map[string]interface{}{"issueId": issueID}

	// 	s, found, err := situation.R().Get(int64(context.SituationID))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if !found {
	// 		return fmt.Errorf("Invalid situation ID or groups not found within the situation")
	// 	}

	// 	//notification description no needed
	// 	description := ""
	// 	notif := notification.NewMockNotification(task.Level, "", task.Name, description,
	// 		time.Now().Truncate(1*time.Millisecond).UTC(), nil, ctx)
	// 	notif.Type = "case"

	// 	notifier.C().SendToRoles(key, timeoutDuration, notif, s.Groups)
	// }

	return nil
}
