package tasker

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/email"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/export"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"go.uber.org/zap"
)

// SituationReportingTask struct for close issues created in the current day from the BRMS
type SituationReportingTask struct {
	ID                  string   `json:"id"`
	Subject             string   `json:"subject"`
	BodyTemplate        string   `json:"bodyTemplate"`
	To                  []string `json:"to"`
	AttachmentFileNames []string `json:"attachmentFileNames"`
	AttachmentFactIDs   []int64  `json:"attachmentFactIds"`
	Columns             []string `json:"columns"`
	ColumnsLabel        []string `json:"columnsLabel"`
	SMTPHost            string   `json:"smtpHost"`
	SMTPPort            string   `json:"smtpPort"`
	SMTPUsername        string   `json:"smtpUsername"`
	SMTPPassword        string   `json:"smtpPassword"`
}

func buildSituationReportingTask(parameters map[string]interface{}) (SituationReportingTask, error) {
	task := SituationReportingTask{}

	if val, ok := parameters["id"].(string); ok && val != "" {
		task.ID = val
	} else {
		return task, errors.New("Missing or invalid 'id' parameter (string not empty required)")
	}

	if val, ok := parameters["subject"].(string); ok && val != "" {
		task.Subject = val
	} else {
		return task, errors.New("Missing or invalid 'subject' parameter (string not empty required)")
	}

	if val, ok := parameters["bodyTemplate"].(string); ok && val != "" {
		task.BodyTemplate = val
	} else {
		return task, errors.New("Missing or invalid 'bodyTemplate' parameter (string not empty required)")
	}

	if val, ok := parameters["to"].(string); ok && val != "" {
		task.To = strings.Split(val, ",")
	} else {
		return task, errors.New("Missing or invalid 'to' parameter (string not empty required)")
	}

	if val, ok := parameters["attachmentFileNames"].(string); ok && val != "" {
		task.AttachmentFileNames = strings.Split(val, ",")
	}

	if val, ok := parameters["attachmentFactIds"].(string); ok && val != "" {
		factIDs := strings.Split(val, ",")
		for _, factIDStr := range factIDs {
			if factID, err := strconv.ParseInt(factIDStr, 10, 64); err == nil {
				task.AttachmentFactIDs = append(task.AttachmentFactIDs, factID)
			} else {
				return task, errors.New("Missing or invalid 'attachmentFactIDs' parameter (string is not an integer)")
			}
		}
	}
	if len(task.AttachmentFileNames) > 1 {
		return task, errors.New("Parameter 'attachmentFileName' must only contains one value (for now)")
	}
	if len(task.AttachmentFactIDs) > 1 {
		return task, errors.New("Parameter 'attachmentFactID' must only contains one value (for now)")
	}
	if len(task.AttachmentFileNames) != len(task.AttachmentFactIDs) {
		return task, errors.New("Parameters 'attachmentFileName' and 'attachmentFactID' have different length")
	}

	if val, ok := parameters["columns"].(string); ok && val != "" {
		task.Columns = strings.Split(val, ",")
	}

	if val, ok := parameters["columnsLabel"].(string); ok && val != "" {
		task.ColumnsLabel = strings.Split(val, ",")
	}

	if len(task.Columns) != len(task.ColumnsLabel) {
		return task, errors.New("Parameters 'attachmentFileName' and 'attachmentFactId' have different length")
	}

	if val, ok := parameters["smtpHost"].(string); ok && val != "" {
		task.SMTPHost = val
	} else {
		return task, errors.New("Missing or invalid 'smtpHost' parameter (string not empty required)")
	}

	if val, ok := parameters["smtpPort"].(string); ok && val != "" {
		task.SMTPPort = val
	} else {
		return task, errors.New("Missing or invalid 'smtpPort' parameter (string not empty required)")
	}

	return task, nil
}

func (task SituationReportingTask) String() string {
	return fmt.Sprint("close today issues with id: ", task.ID)
}

// GetID returns the task key
func (task SituationReportingTask) GetID() string {
	return task.ID
}

// Perform executes the task
func (task SituationReportingTask) Perform(key string, context ContextData) error {
	zap.L().Debug("Perform SituationReportingTask")

	situationData, err := GetSituationKnowledge(context.SituationID, context.TemplateInstanceID, context.TS)
	if err != nil {
		return err
	}

	body, err := BuildMessageBody(task.BodyTemplate, situationData)
	if err != nil {
		return err
	}

	attachments := make([]email.MessageAttachment, 0)
	for i, attachmentFactID := range task.AttachmentFactIDs {
		fullHits, err := export.ExportFactHitsFull(attachmentFactID)
		if err != nil {
			return err
		}

		csvAttachment, err := export.ConvertHitsToCSV(fullHits, task.Columns, task.ColumnsLabel)
		if err != nil {
			return err
		}

		attachments = append(attachments, email.MessageAttachment{
			FileName: task.AttachmentFileNames[i],
			Mime:     "application/octet-stream",
			Content:  csvAttachment,
		})
	}

	message := email.NewMessage(task.Subject, "text/html", string(body))
	message.To = task.To
	message.Attachments = attachments

	sender := email.NewSender(task.SMTPUsername, task.SMTPPassword, task.SMTPHost, task.SMTPPort)
	err = sender.Send(message)
	if err != nil {
		return err
	}

	return nil
}

func BuildMessageBody(templateBody string, templateData map[string]interface{}) ([]byte, error) {
	tmpl, err := template.New("htmlEmail").Parse(templateBody)
	if err != nil {
		return nil, err
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, templateData)
	if err != nil {
		return nil, err
	}

	return body.Bytes(), nil
}

func GetSituationKnowledge(situationID int64, situationInstanceID int64, situationTS time.Time) (map[string]interface{}, error) {
	situationData := make(map[string]interface{})
	record, err := situation.GetFromHistory(situationID, situationTS, situationInstanceID, false)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, errors.New("situation was not found in the history")
	}

	for factID, factTS := range record.FactsIDS {
		f, found, err := fact.R().Get(factID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fmt.Errorf("Fact not found with id=%d", factID)
		}
		if factTS == nil {
			return nil, fmt.Errorf("At least one fact has never been calculated, id=%d, name=%s", f.ID, f.Name)
		}

		item, _, err := fact.GetFactResultFromHistory(factID, *factTS, situationID, situationInstanceID, false, -1)
		if err != nil {
			return nil, err
		}
		itemData, err := item.ToAbstractMap()
		if err != nil {
			return nil, err
		}
		situationData[f.Name] = itemData
	}
	for key, value := range record.Parameters {
		situationData[key] = value
	}
	for key, value := range record.EvaluatedExpressionFacts {
		situationData[key] = value
	}
	return situationData, nil
}
