package models

import (
	"bytes"
	"encoding/json"
)

// CaseState state of a evaluated case (caseResult)
type CaseState int

const (
	// OnError case on error
	OnError CaseState = iota + 1
	// NotEvaluated case not evaluated
	NotEvaluated
	// Met case evaluated and condition result equals to true
	Met
	// Unmet case evaluated and condition result equals to false
	Unmet
)

func (s CaseState) String() string {
	return caseStateToString[s]
}

var caseStateToString = map[CaseState]string{
	OnError:      "onerror",
	NotEvaluated: "notevaluated",
	Met:          "met",
	Unmet:        "unmet",
}

var caseStateToID = map[string]CaseState{
	"onerror":      OnError,
	"notevaluated": NotEvaluated,
	"met":          Met,
	"unmet":        Unmet,
}

// MarshalJSON marshals the enum as a quoted json string
func (s CaseState) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(caseStateToString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *CaseState) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Info' in this case.
	*s = caseStateToID[j]
	return nil
}
