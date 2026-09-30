package tasker

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/utils/emailutils"
	email2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/email"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/fact"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/export"
	"go.uber.org/zap"
)

// Temp solution before proper task condition trigger
var (
	cache   = make(map[string]time.Time)
	cacheMu sync.Mutex
)

func verifyCache(key string, timeout time.Duration) bool {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if val, ok := cache[key]; ok && time.Now().UTC().Before(val) {
		return false
	}
	cache[key] = time.Now().UTC().Add(timeout)
	return true
}

// SituationReportingTask struct for close issues created in the current day from the BRMS
type SituationReportingTask struct {
	ID                  string          `json:"id"`
	IssueID             string          `json:"issueId"`
	Subject             string          `json:"subject"`
	BodyTemplate        string          `json:"bodyTemplate"`
	To                  []string        `json:"to"`
	Cc                  []string        `json:"cc,omitempty"`
	AttachmentFileNames []string        `json:"attachmentFileNames"`
	AttachmentFactIDs   []int64         `json:"attachmentFactIds"`
	Columns             []export.Column `json:"columns"`
	Separator           rune            `json:"separator"`
	Timeout             string          `json:"timeout"`

	// GroupBySituation, when true, holds this instance's report back instead of sending it right
	// away: every instance of the same situation evaluated during the same scheduler run is merged
	// into a single email once the run completes (see collectGroupedSituationReporting).
	GroupBySituation bool `json:"groupBySituation,omitempty"`
	// GroupSubject is the subject used for that merged email (the "Global Title"). Required when
	// GroupBySituation is true.
	GroupSubject string `json:"groupSubject,omitempty"`
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

	if val, ok := parameters["cc"].(string); ok && val != "" {
		task.Cc = strings.Split(val, ",")
	} else if val == "" {
		task.Cc = []string{}
	} else {
		return task, errors.New("Invalid 'cc' parameter")
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
		columns := strings.Split(val, ",")
		var columnsLabel []string

		if val, ok = parameters["columnsLabel"].(string); ok && val != "" {
			columnsLabel = strings.Split(val, ",")
		}

		if len(columns) != len(columnsLabel) {
			return task, errors.New("parameters 'columns' and 'columns label' have different length")
		}

		formatColumnsDataMap := make(map[string]string)

		if val, ok = parameters["formateColumns"].(string); ok && val != "" {
			formatColumnsData := strings.Split(val, ",")
			for _, formatData := range formatColumnsData {
				parts := strings.Split(formatData, ";")
				if len(parts) != 2 {
					continue
				}
				key := strings.TrimSpace(parts[0])
				formatColumnsDataMap[key] = parts[1]
			}
		}

		for i, column := range columns {
			exportColumn := export.Column{
				Name:  column,
				Label: columnsLabel[i],
			}

			if format, ok := formatColumnsDataMap[column]; ok {
				exportColumn.Format = format
			}

			task.Columns = append(task.Columns, exportColumn)
		}
	}

	if val, ok := parameters["separator"].(string); ok && val != "" {
		task.Separator = []rune(val)[0]
	} else {
		task.Separator = ','
	}

	if val, ok := parameters["timeout"].(string); ok && val != "" {
		task.Timeout = val
	} else {
		return task, errors.New("missing or not valid 'timeout' parameter (string not empty required)")
	}

	// groupBySituation is a boolean expression (like 'isNotification' on create-issue), not a quoted
	// string: rule authors write the bare literal true/false as the parameter value.
	if raw, exists := parameters["groupBySituation"]; exists {
		val, ok := raw.(bool)
		if !ok {
			return task, errors.New("invalid 'groupBySituation' parameter (boolean required)")
		}
		task.GroupBySituation = val
	}

	if task.GroupBySituation {
		if val, ok := parameters["groupSubject"].(string); ok && val != "" {
			task.GroupSubject = val
		} else {
			return task, errors.New("missing or invalid 'groupSubject' parameter (required when 'groupBySituation' is true)")
		}

		// Attachments are not supported on grouped reports yet: each instance could reference a
		// different fact export, and merging several CSV attachments into one email is left for a
		// later iteration. Rule.IsValid() rejects this combination at save time; this is the
		// defensive fallback for any rule that predates that check.
		if len(task.AttachmentFileNames) > 0 || len(task.AttachmentFactIDs) > 0 {
			zap.L().Warn("Attachments are not supported when 'groupBySituation' is enabled, ignoring them",
				zap.String("id", task.ID))
			task.AttachmentFileNames = nil
			task.AttachmentFactIDs = nil
		}
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

// isIssueAlreadyHandled returns true when task.IssueID references an issue that is already open or
// in draft, meaning this report must be skipped. Shared by the single-instance Perform path and by
// the grouped path (where it filters out individual members instead of aborting the whole group).
func (task SituationReportingTask) isIssueAlreadyHandled(key string) (bool, error) {
	if task.IssueID == "" {
		return false, nil
	}
	isOpen, _, err := explainer.IsOpenOrDraftIssue(task.IssueID)
	if err != nil {
		zap.L().Error("Cannot search in issue history", zap.String("key", key), zap.Error(err))
		return false, err
	}
	return isOpen, nil
}

// buildReportContent renders this task's email body and attachments for one specific instance
// context. Shared by the single-instance Perform path and by the grouped path, where each member's
// body is rendered independently before being merged into a single email.
func (task SituationReportingTask) buildReportContent(context ContextData) (body []byte, attachments []email2.MessageAttachment, err error) {
	situationData := context.HistorySituationFlattenData
	zap.L().Debug("GetSituationKnowledge()", zap.Any("situationData", situationData))

	body, err = emailutils.BuildMessageBody(task.BodyTemplate, situationData)
	if err != nil {
		zap.L().Error("Error Building MessageBody", zap.Error(err))
		body = []byte("<p>Error Building MessageBody</p>")
	}
	zap.L().Debug("BuildMessageBody()", zap.Any("situationData", situationData))

	attachments = make([]email2.MessageAttachment, 0)
	for i, attachmentFactID := range task.AttachmentFactIDs {
		f, found, err := fact.R().Get(attachmentFactID)
		if err != nil {
			return nil, nil, err
		}
		if !found {
			return nil, nil, fmt.Errorf("attachment fact %d not found", attachmentFactID)
		}

		fullHits, err := export.ExportFactHitsFull(f)
		if err != nil {
			return nil, nil, err
		}

		csvAttachment, err := export.ConvertHitsToCSV(fullHits, export.CSVParameters{Columns: task.Columns, Separator: string(task.Separator)}, true)
		if err != nil {
			return nil, nil, err
		}

		attachmentFileName := task.AttachmentFileNames[i]
		attachments = append(attachments, email2.MessageAttachment{
			FileName: attachmentFileName,
			Mime:     "application/octet-stream",
			Content:  csvAttachment,
		})

		zap.L().Debug("Attachments Added", zap.Any("factID", attachmentFactID))
	}

	return body, attachments, nil
}

// Perform executes the task (non-grouped path): one instance, one email, sent immediately.
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

	skip, err := task.isIssueAlreadyHandled(key)
	if err != nil {
		return err
	}
	if skip {
		zap.L().Debug("SituationReportingTask creation skipped - open/draft issue already existed")
		return nil
	}

	body, attachments, err := task.buildReportContent(context)
	if err != nil {
		return err
	}

	message := email2.NewMessage(task.Subject, "text/html", string(body))
	message.To = task.To
	message.CC = task.Cc
	message.Attachments = attachments
	zap.L().Debug("Message ready to be sent")

	err = email2.S().Send(message)
	if err != nil {
		return err
	}
	zap.L().Info("Email sent !")

	return nil
}
