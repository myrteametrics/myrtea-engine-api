package situation

import (
	"encoding/json"
	"errors"
)

// Situation is a struct used to represent a situation (or an ensemble of fact)
type Situation struct {
	ID              int64             `json:"id,omitempty"`
	Groups          []int64           `json:"groups"`
	Name            string            `json:"name"`
	Facts           []int64           `json:"facts"`
	CalendarID      int64             `json:"calendarId"`
	Parameters      map[string]string `json:"parameters"`
	ExpressionFacts []ExpressionFact  `json:"expressionFacts"`
	IsTemplate      bool              `json:"isTemplate"`
	IsObject        bool              `json:"isObject"`
}

// ExpressionFact represent a custom calculated fact based on gval expression
type ExpressionFact struct {
	Name       string `json:"name"`
	Expression string `json:"expression"`
}

// IsValid checks if an internal schedule definition is valid and has no missing mandatory fields
func (s Situation) IsValid() (bool, error) {
	if s.Name == "" {
		return false, errors.New("missing Name")
	}
	if s.Groups == nil {
		return false, errors.New("missing Groups")
	}
	if len(s.Groups) <= 0 {
		return false, errors.New("missing Groups")
	}
	if s.Facts == nil {
		return false, errors.New("missing Facts")
	}
	if len(s.Facts) <= 0 {
		return false, errors.New("missing Facts")
	}
	return true, nil
}

// MarshalJSON marshals a Situation as a json object
func (s Situation) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID              int64             `json:"id,omitempty"`
		Groups          []int64           `json:"groups"`
		Name            string            `json:"name"`
		Facts           []int64           `json:"facts"`
		ExpressionFacts []ExpressionFact  `json:"expressionFacts"`
		CalendarID      int64             `json:"calendarId"`
		Parameters      map[string]string `json:"parameters"`
		IsTemplate      bool              `json:"isTemplate"`
		IsObject        bool              `json:"isObject"`
	}{
		ID:              s.ID,
		Groups:          s.Groups,
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
	ID          int64             `json:"id"`
	Name        string            `json:"name"`
	SituationID int64             `json:"situationId"`
	Parameters  map[string]string `json:"parameters"`
	CalendarID  int64             `json:"calendarId"`
}
