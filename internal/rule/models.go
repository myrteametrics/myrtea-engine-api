package rule

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/utils/emailutils"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"github.com/myrteametrics/myrtea-sdk/v5/ruleeng"
	"go.uber.org/zap"
)

// Rule represents a business rule
// It is composed of one or more cases
// Each case is composed of one or more conditions and one or more actions
// If all conditions of a case are met, all actions of the case are executed
// If a rule has multiple cases, the evaluation can stop at the first case that is met or evaluate all cases
// depending on the EvaluateAllCases field
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

	if valid, err := r.DefaultRule.IsValid(); !valid {
		return false, err
	}

	// we want to check whether bodyTemplate is a valid template or not
	for _, c := range r.Cases {
		for _, action := range c.Actions {
			for key, param := range action.Parameters {
				if key == "bodyTemplate" {
					result, err := expression.Process(expression.LangEval, string(param), map[string]interface{}{})
					if err != nil {
						zap.L().Warn("Rule IsValid: bodyTemplate expression syntax is invalid", zap.String("bodyTemplate", string(param)), zap.Error(err))
						continue
					}

					if result != nil {
						if _, ok := result.(string); !ok {
							continue
						}
					}

					// we check if the template is valid
					err = emailutils.VerifyMessageBody(result.(string))
					if err != nil {
						return false, fmt.Errorf("invalid bodyTemplate in case '%s': %w", c.Name, err)
					}
				}
			}
		}
	}

	return true, nil
}

// SameCasesAs returns true if the cases of the are equal to the case of the rule passed as parameter or false otherwise
func (r *Rule) SameCasesAs(rule Rule) bool {
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
func (r *Rule) EqualTo(rule Rule) bool {
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
