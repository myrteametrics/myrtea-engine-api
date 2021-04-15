package tasker

import (
	"errors"
	"fmt"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/notifier"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/notifier/notification"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"go.uber.org/zap"
)

// NotifyTask struct to represent a notification task
type NotifyTask struct {
	ID          string                 `json:"id"`
	Level       string                 `json:"level"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Timeout     string                 `json:"timeout"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

func buildNotifyTask(parameters map[string]interface{}) (NotifyTask, error) {
	task := NotifyTask{}

	if val, ok := parameters["id"].(string); ok && val != "" {
		task.ID = val
	} else {
		return task, errors.New("Missing or not valid 'id' parameter (string not empty required)")
	}

	if val, ok := parameters["level"].(string); ok && val != "" {
		task.Level = val
	} else {
		return task, errors.New("Missing or not valid 'level' parameter (string not empty required)")
	}

	if val, ok := parameters["name"].(string); ok && val != "" {
		task.Name = val
	} else {
		return task, errors.New("Missing or not valid 'name' parameter (string not empty required)")
	}

	if val, ok := parameters["description"].(string); ok && val != "" {
		task.Name = val
	} else {
		return task, errors.New("Missing or not valid 'description' parameter (string not empty required)")
	}

	if val, ok := parameters["timeout"].(string); ok && val != "" {
		task.Timeout = val
	} else {
		return task, errors.New("Missing or not valid 'timeout' parameter (string not empty required)")
	}

	if val, ok := parameters["context"].(map[string]interface{}); ok {
		task.Context = val
	} else {
		return task, errors.New("Missing or not valid 'context' parameter (map[string]interface{} required)")
	}

	return task, nil
}

func (task NotifyTask) String() string {
	return fmt.Sprint("Notify ", task.Level, " with content ", task.Description, " and context: ", task.Context)
}

// GetID returns the task key
func (task NotifyTask) GetID() string {
	return task.ID
}

// Perform executes the task
func (task NotifyTask) Perform(key string, context ContextData) error {

	zap.L().Debug("Perform NotifyTask")

	s, found, err := situation.R().Get(int64(context.SituationID))
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("Invalid situation ID or groups not found within the situation")
	}
	notif := notification.NewMockNotification(task.Level, s.Name, task.Name, task.Description,
		time.Now().Truncate(1*time.Millisecond).UTC(), nil, task.Context)
	notif.Type = "generic"

	timeout, err := time.ParseDuration(task.Timeout)
	if err != nil {
		return err
	}

	notifier.C().SendToGroups(key, timeout, notif, s.Groups)

	return nil
}
