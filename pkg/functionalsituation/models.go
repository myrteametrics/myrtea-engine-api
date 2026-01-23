package functionalsituation

import (
	"errors"
	"time"
)

// FunctionalSituation represents a logical grouping of situation instances
type FunctionalSituation struct {
	ID          int64                  `json:"id,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ParentID    *int64                 `json:"parentId,omitempty"`
	Color       string                 `json:"color"`
	Icon        string                 `json:"icon"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
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
	Metadata          map[string]interface{}        `json:"metadata,omitempty"`
	CreatedAt         time.Time                     `json:"createdAt"`
	UpdatedAt         time.Time                     `json:"updatedAt"`
	CreatedBy         string                        `json:"createdBy"`
	TemplateInstances []TreeTemplateInstance        `json:"templateInstances"`
	Situations        []TreeSituation               `json:"situations"`
	Children          []FunctionalSituationTreeNode `json:"children,omitempty"`
}

// TreeTemplateInstance is a lightweight representation of a template instance for the tree view
type TreeTemplateInstance struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	SituationID int64  `json:"situationId"`
}

// TreeSituation is a lightweight representation of a situation for the tree view
type TreeSituation struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	IsTemplate bool   `json:"isTemplate"`
}

// FunctionalSituationCreate represents the payload for creating a new FS
type FunctionalSituationCreate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ParentID    *int64                 `json:"parentId,omitempty"`
	Color       string                 `json:"color"`
	Icon        string                 `json:"icon"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// FunctionalSituationUpdate represents the payload for updating a FS
type FunctionalSituationUpdate struct {
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	ParentID    *int64                  `json:"parentId,omitempty"` // Use -1 to set to NULL
	Color       *string                 `json:"color,omitempty"`
	Icon        *string                 `json:"icon,omitempty"`
	Metadata    *map[string]interface{} `json:"metadata,omitempty"`
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
	return true, nil
}

// IsValid validates the creation payload
func (fsc FunctionalSituationCreate) IsValid() (bool, error) {
	fs := FunctionalSituation{
		Name:  fsc.Name,
		Color: fsc.Color,
		Icon:  fsc.Icon,
	}
	return fs.IsValid()
}
