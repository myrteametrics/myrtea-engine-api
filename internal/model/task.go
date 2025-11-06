package model

import (
	"bytes"
	"encoding/json"
	"time"
)

// CaseInput : Input of a case in a rule evaluation
type CaseInput struct {
	Name      string    `json:"name"`
	Condition string    `json:"condition"`
	State     CaseState `json:"state"`
	Errors    []string  `json:"errors,omitempty"`
}

// RuleInput Input of one evaluation of the rule
type RuleInput struct {
	RuleID          int64       `json:"id"`
	RuleVersion     int64       `json:"version"`
	RuleTitle       string      `json:"title"`
	RuleDescription string      `json:"description"`
	CasesInput      []CaseInput `json:"cases,omitempty"`
}

// InputTask represents the input that the tasker generates for all tasks
type InputTask struct {
	SituationID        int64
	TS                 time.Time
	TemplateInstanceID int64
	Rule               RuleInput
}

// struct Key for an map MetaData
type Key struct {
	SituationID         int64
	SituationInstanceID int64
}

// IssueLevel state of a issue
type IssueLevel int

const (
	// Info information level
	Info IssueLevel = iota + 1
	// Ok stable level
	Ok
	// Warning warning level
	Warning
	// Critical critical level
	Critical
	// Fatal fatal level
	Fatal
)

func (s IssueLevel) String() string {
	if level, ok := issueLevelToString[s]; ok {
		return level
	}
	return ""
}

// ToIssueLevel get the IssueLevel from is string representation
func ToIssueLevel(s string) IssueLevel {
	if level, ok := issueLevelToID[s]; ok {
		return level
	}
	return 0
}

var issueLevelToString = map[IssueLevel]string{
	Info:     "info",
	Ok:       "ok",
	Warning:  "warning",
	Critical: "critical",
	Fatal:    "fatal",
}

var issueLevelToID = map[string]IssueLevel{
	"info":     Info,
	"ok":       Ok,
	"warning":  Warning,
	"critical": Critical,
	"fatal":    Fatal,
}

// MarshalJSON marshals the enum as a quoted json string
func (s IssueLevel) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(issueLevelToString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *IssueLevel) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Info' in this case.
	*s = issueLevelToID[j]
	return nil
}
