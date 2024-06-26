package rule

import (
	"encoding/json"
	"errors"

	"github.com/myrteametrics/myrtea-sdk/v5/ruleeng"
)

// Rule ...
type Rule struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	CalendarID  int    `json:"calendarId"`
	ruleeng.DefaultRule
}

// IsValid checks if a rule definition is valid and has no missing mandatory fields
func (r *Rule) IsValid() (bool, error) {
	if r.Name == "" {
		return false, errors.New("missing Name")
	}
	if r.Description == "" {
		return false, errors.New("missing Title")
	}
	if r.Cases == nil {
		return false, errors.New("missing Cases")
	}
	if len(r.Cases) <= 0 {
		return false, errors.New("missing Cases")
	}

	return true, nil
}

// SameCasesAs returns true if the cases of the are equal to the case of the rule passed as parameter or false otherwise
func (r Rule) SameCasesAs(rule Rule) bool {
	rCasesData, err := json.Marshal(r.Cases)
	if err != nil {
		return false
	}
	ruleCasesData, err := json.Marshal(rule.Cases)
	if err != nil {
		return false
	}
	return string(rCasesData) == string(ruleCasesData)
}

// EqualTo returns true if the rule is equal to the rule passed as parameter or false otherwise
func (r Rule) EqualTo(rule Rule) bool {
	if r.Name != rule.Name {
		return false
	}
	if !r.SameCasesAs(rule) {
		return false
	}
	if r.Version != rule.Version {
		return false
	}
	return r.Enabled == rule.Enabled
}

// UnmarshalJSON unmashals a quoted json string to Expression
func (r *Rule) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Name             string                 `json:"name"`
		Description      string                 `json:"description"`
		Enabled          bool                   `json:"enabled"`
		CalendarID       int                    `json:"calendarId"`
		ID               int64                  `json:"id,omitempty"`
		Cases            []ruleeng.Case         `json:"cases"`
		Version          int64                  `json:"version"`
		Parameters       map[string]interface{} `json:"parameters"`
		EvaluateAllCases bool                   `json:"evaluateallcase"`
	}
	aux := Alias{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	r.Name = aux.Name
	r.Description = aux.Description
	r.Enabled = aux.Enabled
	r.CalendarID = aux.CalendarID
	r.ID = aux.ID
	r.Cases = aux.Cases
	r.Version = aux.Version
	r.Parameters = aux.Parameters
	r.EvaluateAllCases = aux.EvaluateAllCases

	return nil
}
