package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

// Action object that represents the actions to be taken for an Issue
type Action struct {
	ID          int64  `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	RootCauseID int64  `json:"rootCauseId"`
}

// NewAction returns a new action
func NewAction(id int64, name string, description string, rootCauseID int64) Action {
	return Action{
		ID:          id,
		Name:        name,
		Description: description,
		RootCauseID: rootCauseID,
	}
}

// IsValid checks if an action definition is valid and has no missing mandatory fields
func (action *Action) IsValid() (bool, error) {
	if action.Name == "" {
		return false, errors.New("missing Name")
	}
	if action.RootCauseID == 0 {
		return false, errors.New("missing RootCauseID (or 0 value)")
	}
	return true, nil
}

// RootCause is the causes for the Issues
type RootCause struct {
	ID          int64  `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	SituationID int64  `json:"situationId"`
	RuleID      int64  `json:"ruleId"`
}

// NewRootCause returns a new action
func NewRootCause(id int64, name string, description string, situationID int64, ruleID int64) RootCause {
	return RootCause{
		ID:          id,
		Name:        name,
		Description: description,
		SituationID: situationID,
		RuleID:      ruleID,
	}
}

// IsValid checks if a rootcause definition is valid and has no missing mandatory fields
func (rootcause *RootCause) IsValid() (bool, error) {
	if rootcause.Name == "" {
		return false, errors.New("missing Name")
	}
	if rootcause.SituationID == 0 {
		return false, errors.New("missing SituationID (or 0 value)")
	}
	if rootcause.RuleID == 0 {
		return false, errors.New("missing RuleID (or 0 value)")
	}
	return true, nil
}

// Issue is created for a situation to take Root causes and Actions
type Issue struct {
	ID                 int64      `json:"id"`
	Key                string     `json:"key"`
	Name               string     `json:"name"`
	Level              IssueLevel `json:"level"`
	SituationHistoryID int64      `json:"situationHistoryId"`
	SituationID        int64      `json:"situationId"`
	TemplateInstanceID int64      `json:"templateInstanceId"`
	SituationTS        time.Time  `json:"situationDate"`
	ExpirationTS       time.Time  `json:"expirationDate"`
	Rule               RuleData   `json:"rule"`
	State              IssueState `json:"state"`
	CreationTS         time.Time  `json:"createdAt,omitempty"`
	LastModificationTS time.Time  `json:"lastModified"`
	DetectionRatingAvg float64    `json:"detectionRatingAvg,omitempty"`
	AssignedAt         *time.Time `json:"assignedAt,omitempty"`
	AssignedTo         *string    `json:"assignedTo,omitempty"`
	ClosedAt           *time.Time `json:"closedAt,omitempty"`
	CloseBy            *string    `json:"closedBy,omitempty"`
	Comment            *string    `json:"comment,omitempty"`
}

// RuleData rule identification
type RuleData struct {
	RuleID      int64  `json:"ruleId"`
	RuleVersion int64  `json:"ruleVersion"`
	CaseName    string `json:"caseName"`
}

// IssueState represents the state in which the Issue is
type IssueState int

const (
	// Open state of issue
	Open IssueState = iota + 1
	// Draft state of issue
	Draft
	// ClosedNoFeedback state of issue
	ClosedNoFeedback
	// ClosedTimeout state of issue
	ClosedTimeout
	// ClosedDiscard state of issue
	ClosedDiscard
	// ClosedFeedbackConfirmed state of issue
	ClosedFeedbackConfirmed
	// ClosedFeedbackRejected state of issue
	ClosedFeedbackRejected
	// ClosedConfirmed  state of issue
	ClosedConfirmed
	// ClosedRejected  state of issue
	ClosedRejected
)

// IsClosed returns if the IssueState is a closed state
func (s IssueState) IsClosed() bool {
	switch s {
	case
		ClosedNoFeedback,
		ClosedTimeout,
		ClosedDiscard,
		ClosedConfirmed,
		ClosedRejected,
		ClosedFeedbackConfirmed,
		ClosedFeedbackRejected:
		return true
	}
	return false
}

func (s IssueState) String() string {
	return issueStateToString[s]
}

// ToIssueState get the IssueState from is string representation
func ToIssueState(s string) IssueState {
	if state, ok := issueStateToID[s]; ok {
		return state
	}
	return 0
}

var issueStateToString = map[IssueState]string{
	Open:                    "open",
	Draft:                   "draft",
	ClosedNoFeedback:        "closednofeedback",
	ClosedTimeout:           "closedtimeout",
	ClosedDiscard:           "closeddiscard",
	ClosedFeedbackConfirmed: "closedfeedbackconfirmed",
	ClosedFeedbackRejected:  "closedfeedbackrejected",
	ClosedConfirmed:         "closedconfirmed",
	ClosedRejected:          "closedrejected",
}

var issueStateToID = map[string]IssueState{
	"open":                    Open,
	"draft":                   Draft,
	"closednofeedback":        ClosedNoFeedback,
	"closedtimeout":           ClosedTimeout,
	"closeddiscard":           ClosedDiscard,
	"closedfeedbackconfirmed": ClosedFeedbackConfirmed,
	"closedfeedbackrejected":  ClosedFeedbackRejected,
	"closedconfirmed":         ClosedConfirmed,
	"closedrejected":          ClosedRejected,
}

// GetStringIssueState gets the string representation of a IssueState
func GetStringIssueState(issueState IssueState) string {
	if state, ok := issueStateToString[issueState]; ok {
		return state
	}
	return ""
}

// MarshalJSON marshals the enum as a quoted json string
func (s IssueState) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(issueStateToString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *IssueState) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Info' in this case.
	*s = issueStateToID[j]
	return nil
}
