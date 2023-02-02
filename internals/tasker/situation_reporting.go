package tasker

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/email"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/explainer"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/export"
	"go.uber.org/zap"
)

// Temp solution before proper task condition trigger
var cache map[string]time.Time = make(map[string]time.Time, 0)

func verifyCache(key string, timeout time.Duration) bool {
	if val, ok := cache[key]; ok && time.Now().UTC().Before(val) {
		return false
	}
	cache[key] = time.Now().UTC().Add(timeout)
	return true
}

// SituationReportingTask struct for close issues created in the current day from the BRMS
type SituationReportingTask struct {
	ID                  string   `json:"id"`
	IssueID             string   `json:"issueId"`
	Subject             string   `json:"subject"`
	BodyTemplate        string   `json:"bodyTemplate"`
	To                  []string `json:"to"`
	AttachmentFileNames []string `json:"attachmentFileNames"`
	AttachmentFactIDs   []int64  `json:"attachmentFactIds"`
	Columns             []string `json:"columns"`
	ColumnsLabel        []string `json:"columnsLabel"`
	Separator           rune     `json:"separator"`
	SMTPHost            string   `json:"smtpHost"`
	SMTPPort            string   `json:"smtpPort"`
	SMTPUsername        string   `json:"smtpUsername"`
	SMTPPassword        string   `json:"smtpPassword"`
	Timeout             string   `json:"timeout"`
}

func buildSituationReportingTask(parameters map[string]interface{}) (SituationReportingTask, error) {
	task := SituationReportingTask{}

	if val, ok := parameters["id"].(string); ok && val != "" {
		task.ID = val
	} else {
		return task, errors.New("missing or invalid 'id' parameter (string not empty required)")
	}

	if val, ok := parameters["issueId"].(string); ok && val != "" {
		task.IssueID = val
	}

	if val, ok := parameters["subject"].(string); ok && val != "" {
		task.Subject = val
	} else {
		return task, errors.New("missing or invalid 'subject' parameter (string not empty required)")
	}

	if val, ok := parameters["bodyTemplate"].(string); ok && val != "" {
		task.BodyTemplate = strings.ReplaceAll(val, "'", "\"")
	} else {
		return task, errors.New("missing or invalid 'bodyTemplate' parameter (string not empty required)")
	}

	if val, ok := parameters["to"].(string); ok && val != "" {
		task.To = strings.Split(val, ",")
	} else {
		return task, errors.New("missing or invalid 'to' parameter (string not empty required)")
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
				return task, errors.New("missing or invalid 'attachmentFactIDs' parameter (string is not an integer)")
			}
		}
	}
	if len(task.AttachmentFileNames) > 1 {
		return task, errors.New("parameter 'attachmentFileName' must only contains one value (for now)")
	}
	if len(task.AttachmentFactIDs) > 1 {
		return task, errors.New("parameter 'attachmentFactID' must only contains one value (for now)")
	}
	if len(task.AttachmentFileNames) != len(task.AttachmentFactIDs) {
		return task, errors.New("parameters 'attachmentFileName' and 'attachmentFactID' have different length")
	}

	if val, ok := parameters["columns"].(string); ok && val != "" {
		task.Columns = strings.Split(val, ",")
	}

	if val, ok := parameters["columnsLabel"].(string); ok && val != "" {
		task.ColumnsLabel = strings.Split(val, ",")
	}

	if val, ok := parameters["separator"].(string); ok && val != "" {
		task.Separator = []rune(val)[0]
	} else {
		task.Separator = ','
	}

	if len(task.Columns) != len(task.ColumnsLabel) {
		return task, errors.New("parameters 'columns' and 'colomns label' have different length")
	}

	if val, ok := parameters["smtpHost"].(string); ok && val != "" {
		task.SMTPHost = val
	} else {
		return task, errors.New("missing or invalid 'smtpHost' parameter (string not empty required)")
	}

	if val, ok := parameters["smtpPort"].(string); ok && val != "" {
		task.SMTPPort = val
	} else {
		return task, errors.New("missing or invalid 'smtpPort' parameter (string not empty required)")
	}

	if val, ok := parameters["smtpUsername"].(string); ok && val != "" {
		task.SMTPUsername = val
	} else {
		return task, errors.New("missing or invalid 'smtpUsername' parameter (string not empty required)")
	}

	if val, ok := parameters["smtpPassword"].(string); ok && val != "" {
		task.SMTPPassword = val
	}

	if val, ok := parameters["timeout"].(string); ok && val != "" {
		task.Timeout = val
	} else {
		return task, errors.New("missing or not valid 'timeout' parameter (string not empty required)")
	}

	return task, nil
}

func (task SituationReportingTask) String() string {
	return fmt.Sprint("situation reporting with id: ", task.ID)
}

// GetID returns the task key
func (task SituationReportingTask) GetID() string {
	return task.ID
}

// Perform executes the task
func (task SituationReportingTask) Perform(key string, context ContextData) error {
	zap.L().Info("Perform SituationReportingTask", zap.Any("task", task), zap.Any("key", key), zap.Any("context", context))

	//	Parsing the timeout from string to duration
	timeoutDuration, err := time.ParseDuration(task.Timeout)
	if err != nil {
		return err
	}

	if !verifyCache(key, timeoutDuration) {
		zap.L().Debug("SituationReportingTask skipped - timeout not reached")
		return nil
	}

	if task.IssueID != "" {
		isOpen, err := explainer.IsOpenOrDraftIssue(task.IssueID)
		if err != nil {
			zap.L().Error("Cannot search in issue history", zap.String("key", key), zap.Error(err))
			return err
		}
		if isOpen {
			zap.L().Debug("SituationReportingTask creation skipped - open/draft issue already existed")
			return nil
		}
	}

	situationData := context.HistorySituationFlattenData
	zap.L().Debug("GetSituationKnowledge()", zap.Any("situationData", situationData))

	var body []byte
	body, err = BuildMessageBody(task.BodyTemplate, situationData)
	if err != nil {
		zap.L().Error("Error Building MessageBody", zap.Error(err))
		body = []byte("<p>Error Building MessageBody</p>")
	}
	zap.L().Debug("BuildMessageBody()", zap.Any("situationData", situationData))

	attachments := make([]email.MessageAttachment, 0)
	for i, attachmentFactID := range task.AttachmentFactIDs {
		fullHits, err := export.ExportFactHitsFull(attachmentFactID)
		if err != nil {
			return err
		}

		csvAttachment, err := export.ConvertHitsToCSV(fullHits, task.Columns, task.ColumnsLabel, task.Separator)
		if err != nil {
			return err
		}

		var attachmentFileName = task.AttachmentFileNames[i]
		if strings.Contains(attachmentFileName, ".") &&
			len(strings.Split(attachmentFileName, ".")) > 1 &&
			strings.Split(attachmentFileName, ".")[1] == "zip" {
			zipAttachment, err := export.CreatePasswordProtectedZipFile(attachmentFileName, csvAttachment)
			if err != nil {
				attachments = append(attachments, email.MessageAttachment{
					FileName: attachmentFileName,
					Mime:     "application/zip",
					Content:  zipAttachment,
				})
			}
		} else {
			attachments = append(attachments, email.MessageAttachment{
				FileName: attachmentFileName,
				Mime:     "application/octet-stream",
				Content:  csvAttachment,
			})
		}

		zap.L().Debug("Attachments Added", zap.Any("factID", attachmentFactID))
	}

	message := email.NewMessage(task.Subject, "text/html", string(body))
	message.To = task.To
	message.Attachments = attachments
	zap.L().Debug("Message ready to be sent")

	sender := email.NewSender(task.SMTPUsername, task.SMTPPassword, task.SMTPHost, task.SMTPPort)
	zap.L().Debug("Email sender ready")

	err = sender.Send(message)
	if err != nil {
		return err
	}
	zap.L().Info("Email sent !")

	return nil
}

func BuildMessageBody(templateBody string, templateData map[string]interface{}) ([]byte, error) {
	tmpl, err := template.New("htmlEmail").Funcs(template.FuncMap{
		"split": func(input string, separator string) []string {
			return strings.Split(input, separator)
		},
	}).Parse(templateBody)
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
