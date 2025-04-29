package situation

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
)

// Situation is a struct used to represent a situation (or an ensemble of fact)
type Situation struct {
	ID              int64                  `json:"id,omitempty"`
	Name            string                 `json:"name"`
	Facts           []int64                `json:"facts"`
	CalendarID      int64                  `json:"calendarId"`
	Parameters      map[string]interface{} `json:"parameters"`
	ExpressionFacts []ExpressionFact       `json:"expressionFacts"`
	IsTemplate      bool                   `json:"isTemplate"`
	IsObject        bool                   `json:"isObject"`
}

// ExpressionFact represent a custom calculated fact based on gval expression
type ExpressionFact struct {
	Name       string `json:"name"`
	Expression string `json:"expression"`
}

// IsValid checks if an situation definition is valid and has no missing mandatory fields
func (s Situation) IsValid() (bool, error) {
	if s.Name == "" {
		return false, errors.New("missing Name")
	}
	if s.Facts == nil {
		return false, errors.New("missing Facts")
	}
	if len(s.Facts) <= 0 {
		return false, errors.New("missing Facts")
	}

	// we want to verify, if all parameter's syntaxes are valid
	for key, value := range s.Parameters {
		_, err := expression.Process(expression.LangEval, value.(string), map[string]interface{}{})
		if err != nil {
			return false, fmt.Errorf("parameters: the value of the key %s could not be evaluated: %s", key, err.Error())
		}
	}

	return true, nil
}

// MarshalJSON marshals a Situation as a json object
func (s Situation) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID              int64                  `json:"id,omitempty"`
		Name            string                 `json:"name"`
		Facts           []int64                `json:"facts"`
		ExpressionFacts []ExpressionFact       `json:"expressionFacts"`
		CalendarID      int64                  `json:"calendarId"`
		Parameters      map[string]interface{} `json:"parameters"`
		IsTemplate      bool                   `json:"isTemplate"`
		IsObject        bool                   `json:"isObject"`
	}{
		ID:              s.ID,
		Name:            s.Name,
		Facts:           s.Facts,
		ExpressionFacts: s.ExpressionFacts,
		CalendarID:      s.CalendarID,
		Parameters:      s.Parameters,
		IsTemplate:      s.IsTemplate,
		IsObject:        s.IsObject,
	})
}

// TemplateInstance is a struct used to represent a situation template instance
type TemplateInstance struct {
	ID                  int64                  `json:"id"`
	Name                string                 `json:"name"`
	SituationID         int64                  `json:"situationId"`
	Parameters          map[string]interface{} `json:"parameters"`
	CalendarID          int64                  `json:"calendarId"`
	EnableDependsOn     bool                   `json:"enableDependsOn"`
	DependsOnParameters map[string]string      `json:"dependsOnParameters"`
}

// IsValid checks if an situation template definition is valid and has no missing mandatory fields
func (s TemplateInstance) IsValid() (bool, error) {
	if s.Name == "" {
		return false, errors.New("missing Name")
	}

	// we want to verify, if all parameter's syntaxes are valid
	for key, value := range s.Parameters {
		_, err := expression.Process(expression.LangEval, value.(string), map[string]interface{}{})
		if err != nil {
			return false, fmt.Errorf("parameters: the value of the key %s could not be evaluated: %s", key, err.Error())
		}
	}

	return true, nil
}
