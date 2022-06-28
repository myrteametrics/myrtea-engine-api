package explainer

import (
	"errors"
	"fmt"
	"sort"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/action"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/draft"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/rootcause"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
)

// GetRecommendationTree build a recommendation tree based on issue resolution stats table
func GetRecommendationTree(issue models.Issue) (*models.FrontRecommendation, error) {
	var recommendation *models.FrontRecommendation
	var err error
	switch {
	case issue.State == models.Open:
		recommendation, err = buildRecommendationTree(issue.SituationID, issue.Rule.RuleID)
		if err != nil {
			return nil, err
		}

	case issue.State == models.Draft:
		exists, err := draft.R().CheckExists(nil, issue.ID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("No existing draft found for the issue %d with state Draft", issue.ID)
		}
		draft, found, err := draft.R().Get(issue.ID)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fmt.Errorf("No existing draft found for the issue %d with state Draft", issue.ID)
		}
		recommendation = &draft

	case issue.State.IsClosed():
		exists, err := draft.R().CheckExists(nil, issue.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			draft, found, err := draft.R().Get(issue.ID)
			if err != nil {
				return nil, err
			}
			if !found {
				return nil, fmt.Errorf("No existing draft found for the issue %d with state Draft", issue.ID)
			}
			recommendation = &draft
		} else {
			recommendation, err = buildRecommendationTree(issue.SituationID, issue.Rule.RuleID)
			if err != nil {
				return nil, err
			}
		}
	}
	return recommendation, nil
}

// ExtractSelectedFromTree extracts and returns selected rootcause and actions from a recommendation
func ExtractSelectedFromTree(recommendation models.FrontRecommendation) (*models.FrontRootCause, []*models.FrontAction, error) {
	var selectedRootCause *models.FrontRootCause
	selectedActions := make([]*models.FrontAction, 0)

	for _, rootCause := range recommendation.Tree {
		if rootCause.Selected {
			if selectedRootCause != nil {
				return nil, nil, errors.New("a feedback can't have multiple selected rootcause")
			}

			selectedRootCause = rootCause
			for _, action := range rootCause.Actions {
				if action.Selected {
					selectedActions = append(selectedActions, action)
				}
			}
		}
	}
	if selectedRootCause == nil {
		return nil, nil, errors.New("a feedback must have one rootcause selected")
	}
	if len(selectedActions) == 0 {
		return nil, nil, errors.New("a feedback must have at least one action selected")
	}
	return selectedRootCause, selectedActions, nil
}

func buildRecommendationTree(situationID int64, ruleID int64) (*models.FrontRecommendation, error) {
	tree, err := buildRootCauseTree(situationID, ruleID)
	if err != nil {
		return nil, fmt.Errorf("buildRootCauseTree(): %s", err.Error())
	}

	err = enrichTreeWithActions(situationID, tree)
	if err != nil {
		return nil, fmt.Errorf("enrichTreeWithActions(): %s", err.Error())
	}

	sortRecommendationTree(tree)

	return &models.FrontRecommendation{Tree: tree}, nil
}

func buildRootCauseTree(situationID int64, ruleID int64) ([]*models.FrontRootCause, error) {
	rootCauseStats, err := GetRootCauseStats(situationID)
	if err != nil {
		return nil, err
	}
	rootCauses := make([]*models.FrontRootCause, 0)
	rootCausesDescs, err := rootcause.R().GetAllBySituationIDRuleID(situationID, ruleID)
	if err != nil {
		return nil, err
	}
	if rootCausesDescs == nil {
		return nil, errors.New("nil rootcause map")
	}
	for rootCauseID, rootCauseDesc := range rootCausesDescs {
		rootCause := models.FrontRootCause{
			ID:              rootCauseID,
			Name:            rootCauseDesc.Name,
			Description:     rootCauseDesc.Description,
			Selected:        false,
			Custom:          false,
			Occurrence:      0,
			UsageRate:       0,
			ClusteringScore: -1,
			Actions:         make([]*models.FrontAction, 0),
		}
		if stat, ok := rootCauseStats[rootCauseID]; ok {
			rootCause.Occurrence = stat.Occurrences
			rootCause.UsageRate = stat.GetRelativeFrequency()
		}
		rootCauses = append(rootCauses, &rootCause)
	}
	return rootCauses, nil
}

func enrichTreeWithActions(situationID int64, rootCauses []*models.FrontRootCause) error {
	actionStats, err := GetActionStats(situationID)
	if err != nil {
		return err
	}

	for _, rootCause := range rootCauses {
		actions := make([]*models.FrontAction, 0)
		actionDescs, err := action.R().GetAllByRootCauseID(rootCause.ID)
		if err != nil {
			return err
		}
		if actionDescs == nil {
			return errors.New("nil rootcause map")
		}
		for actionID, actionDesc := range actionDescs {
			action := models.FrontAction{
				ID:          actionID,
				Name:        actionDesc.Name,
				Description: actionDesc.Description,
				Selected:    false,
				Custom:      false,
				Occurrence:  0,
				UsageRate:   0,
			}
			if stat, ok := actionStats[actionID]; ok {
				action.Occurrence = stat.Occurrences
				action.UsageRate = stat.GetRelativeFrequency()
			}
			actions = append(actions, &action)
		}
		rootCause.Actions = actions
	}
	return nil
}

func sortRecommendationTree(rootCauses []*models.FrontRootCause) {
	sort.SliceStable(rootCauses, func(i, j int) bool {
		return rootCauses[i].UsageRate > rootCauses[j].UsageRate
	})

	for _, rootCause := range rootCauses {
		actions := rootCause.Actions
		sort.SliceStable(actions, func(i, j int) bool {
			return actions[i].UsageRate > actions[j].UsageRate
		})
	}
}
