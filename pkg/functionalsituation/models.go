package functionalsituation

import (
	"errors"
	"fmt"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/situation"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
)

// FunctionalSituation represents a logical grouping of situation instances
type FunctionalSituation struct {
	ID          int64                  `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ParentID    *int64                 `json:"parentId,omitempty"`
	Color       string                 `json:"color"`
	Icon        string                 `json:"icon"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	CreatedBy   string                 `json:"createdBy"`

	// Computed fields (non persisted, used for responses)
	Children            []FunctionalSituation `json:"children,omitempty"`
	TemplateInstanceIDs []int64               `json:"templateInstanceIds,omitempty"`
	SituationIDs        []int64               `json:"situationIds,omitempty"`
}

// FunctionalSituationOverview provides a summary view with aggregated status
type FunctionalSituationOverview struct {
	ID               int64                         `json:"id"`
	Name             string                        `json:"name"`
	Description      string                        `json:"description"`
	Color            string                        `json:"color"`
	Icon             string                        `json:"icon"`
	ParentID         *int64                        `json:"parentId,omitempty"`
	InstanceCount    int                           `json:"instanceCount"`
	SituationCount   int                           `json:"situationCount"`
	ChildrenCount    int                           `json:"childrenCount"`
	AggregatedStatus string                        `json:"aggregatedStatus"` // "ok", "warning", "critical", "unknown"
	Children         []FunctionalSituationOverview `json:"children,omitempty"`
}

// FunctionalSituationTreeNode represents a node in the enriched tree with full instance/situation data
type FunctionalSituationTreeNode struct {
	ID                int64                         `json:"id"`
	Name              string                        `json:"name"`
	Description       string                        `json:"description"`
	ParentID          *int64                        `json:"parentId,omitempty"`
	Color             string                        `json:"color"`
	Icon              string                        `json:"icon"`
	Parameters        map[string]interface{}        `json:"parameters,omitempty"`
	CreatedAt         time.Time                     `json:"createdAt"`
	UpdatedAt         time.Time                     `json:"updatedAt"`
	CreatedBy         string                        `json:"createdBy"`
	TemplateInstances []TreeTemplateInstance        `json:"templateInstances"`
	Situations        []TreeSituation               `json:"situations"`
	Children          []FunctionalSituationTreeNode `json:"children,omitempty"`
}

// TreeTemplateInstance is a lightweight representation of a template instance for the tree view
type TreeTemplateInstance struct {
	ID            int64                  `json:"id"`
	Name          string                 `json:"name"`
	SituationID   int64                  `json:"situationId"`
	SituationName string                 `json:"situationName"`
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
}

// TreeSituation is a lightweight representation of a situation for the tree view
type TreeSituation struct {
	ID         int64                  `json:"id"`
	Name       string                 `json:"name"`
	IsTemplate bool                   `json:"isTemplate"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// InstanceReference represents a reference to a template instance with its unique parameters
type InstanceReference struct {
	TemplateInstanceID int64                  `json:"templateInstanceId"`
	Parameters         map[string]interface{} `json:"parameters,omitempty"`
}

// SituationReference represents a reference to a situation with its unique parameters
type SituationReference struct {
	SituationID int64                  `json:"situationId"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// TemplateInstanceWithParameters combines a template instance with its functional situation parameters
type TemplateInstanceWithParameters struct {
	Instance   situation.TemplateInstance `json:"instance"`
	Parameters map[string]interface{}     `json:"parameters,omitempty"`
}

// SituationWithParameters combines a situation with its functional situation parameters
type SituationWithParameters struct {
	Situation  situation.Situation    `json:"situation"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// FunctionalSituationCreate represents the payload for creating a new FS
type FunctionalSituationCreate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ParentID    *int64                 `json:"parentId,omitempty"`
	Color       string                 `json:"color"`
	Icon        string                 `json:"icon"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// FunctionalSituationUpdate represents the payload for updating a FS
type FunctionalSituationUpdate struct {
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	ParentID    *int64                  `json:"parentId,omitempty"` // Use -1 to set to NULL
	Color       *string                 `json:"color,omitempty"`
	Icon        *string                 `json:"icon,omitempty"`
	Parameters  *map[string]interface{} `json:"parameters,omitempty"`
}

// ValidateParameters verifies that all parameter values are valid expressions
func ValidateParameters(parameters map[string]interface{}) (bool, error) {
	if parameters == nil {
		return true, nil
	}

	// Verify if all parameter's syntaxes are valid
	for key, value := range parameters {
		if value == nil {
			continue
		}

		// Convert to string if possible
		strValue, ok := value.(string)
		if !ok {
			continue
		}

		_, err := expression.Process(expression.LangEval, strValue, map[string]interface{}{})
		if err != nil {
			return false, fmt.Errorf("parameters: the value of the key %s could not be evaluated: %s", key, err.Error())
		}
	}
	return true, nil
}

// IsValid validates the functional situation definition
func (fs FunctionalSituation) IsValid() (bool, error) {
	if fs.Name == "" {
		return false, errors.New("missing name")
	}
	if len(fs.Name) > 100 {
		return false, errors.New("name too long (max 100 characters)")
	}
	if fs.Color != "" {
		if len(fs.Color) != 7 || fs.Color[0] != '#' {
			return false, errors.New("invalid color format (expected #RRGGBB)")
		}
	}
	if fs.Icon != "" && len(fs.Icon) > 50 {
		return false, errors.New("icon name too long (max 50 characters)")
	}

	// Validate parameters
	if ok, err := ValidateParameters(fs.Parameters); !ok {
		return false, err
	}

	return true, nil
}

// IsValid validates the creation payload
func (fsc FunctionalSituationCreate) IsValid() (bool, error) {
	fs := FunctionalSituation{
		Name:       fsc.Name,
		Color:      fsc.Color,
		Icon:       fsc.Icon,
		Parameters: fsc.Parameters,
	}
	return fs.IsValid()
}

// IsValid validates the instance reference
func (ir InstanceReference) IsValid() (bool, error) {
	if ir.TemplateInstanceID <= 0 {
		return false, errors.New("invalid template instance ID")
	}
	return ValidateParameters(ir.Parameters)
}

// IsValid validates the situation reference
func (sr SituationReference) IsValid() (bool, error) {
	if sr.SituationID <= 0 {
		return false, errors.New("invalid situation ID")
	}
	return ValidateParameters(sr.Parameters)
}
