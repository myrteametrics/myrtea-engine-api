package explainer

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/action"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/rootcause"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"github.com/myrteametrics/myrtea-sdk/v5/postgres"
)

// CloseIssueWithoutFeedback close an issue without standard feedback on rootcause / action
func CloseIssueWithoutFeedback(dbClient *sqlx.DB, issue model.Issue, user users.User, targetState model.IssueState) error {

	if issue.State.IsClosed() {
		return fmt.Errorf("Issue with id %d is already in a closed state", issue.ID)
	}
	tx, err := dbClient.Beginx()
	if err != nil {
		return err
	}

	err = updateIssueState(tx, issue, targetState, user)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// CloseIssueWithFeedback generate and persist an issue feedback for the rootcause/action stats and ML models
func CloseIssueWithFeedback(dbClient *sqlx.DB, issue model.Issue, recommendation model.FrontRecommendation, user users.User, isFakeAlert bool) error {

	if issue.State.IsClosed() {
		return fmt.Errorf("Issue with id %d is already in a closed state", issue.ID)
	}

	var targetState model.IssueState
	if isFakeAlert {
		targetState = model.ClosedFeedbackRejected
	} else {
		targetState = model.ClosedFeedbackConfirmed
	}

	exists, err := checkExistsIssueResolution(dbClient, issue.ID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("a feedback has already been done on this issue")
	}

	selectedRootCause, selectedActions, err := ExtractSelectedFromTree(recommendation)
	if err != nil {
		return err
	}

	tx, err := dbClient.Beginx()
	if err != nil {
		return err
	}

	err = persistIssueFeedback(tx, issue, selectedRootCause, selectedActions)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = SaveIssueDraft(tx, issue, recommendation, user)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = updateIssueState(tx, issue, targetState, user)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func checkExistsIssueResolution(dbClient *sqlx.DB, issueID int64) (bool, error) {
	var exists bool
	checkNameQuery := `select exists(select 1 from issue_resolution_v1 where issue_id = $1) AS "exists"`
	err := dbClient.QueryRow(checkNameQuery, issueID).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

func updateIssueState(tx *sqlx.Tx, issue model.Issue, targetState model.IssueState, user users.User) error {
	issue.State = targetState
	err := issues.R().Update(tx, issue.ID, issue, user)
	if err != nil {
		return err
	}
	return nil
}

// persistIssueFeedback persist a rootcause and an ensemble of actions related to the resolution of an issue
func persistIssueFeedback(tx *sqlx.Tx, issue model.Issue, selectedRootCause *model.FrontRootCause, selectedActions []*model.FrontAction) error {

	situationID := issue.SituationID
	ruleID := issue.Rule.RuleID

	// Create new rootcause if needed
	dbRootCauseID := selectedRootCause.ID
	var err error
	if selectedRootCause.Custom {
		dbRootCause := model.NewRootCause(-1, selectedRootCause.Name, selectedRootCause.Description, situationID, ruleID)
		dbRootCauseID, err = rootcause.R().Create(tx, dbRootCause)
		if err != nil {
			return err
		}
	} else {
		checkRC, found, err := rootcause.R().Get(selectedRootCause.ID)
		if err != nil {
			return err
		}
		if !found {
			return errors.New("this rootcause doesn't not exists")
		}
		if checkRC.SituationID != situationID {
			return errors.New("this rootcause cannot be used on the current issue/situation")
		}
		if checkRC.RuleID != ruleID {
			return errors.New("this rootcause cannot be used on the current issue/situation (invalid rule)")
		}
	}

	// Create new actions if needed
	dbActionIDs := make([]int64, 0)
	for _, selectedAction := range selectedActions {
		dbActionID := selectedAction.ID
		if selectedAction.Custom {
			dbAction := model.NewAction(-1, selectedAction.Name, selectedAction.Description, dbRootCauseID)
			dbActionID, err = action.R().Create(tx, dbAction)
			if err != nil {
				return err
			}
		} else {
			checkAction, found, err := action.R().Get(selectedAction.ID)
			if err != nil {
				return err
			}
			if !found {
				return errors.New("this action doesn't not exists")
			}
			if checkAction.RootCauseID != dbRootCauseID {
				return errors.New("this action cannot be used on the current rootcause")
			}
		}
		dbActionIDs = append(dbActionIDs, dbActionID)
	}

	// Persist feedback in resolutions stats table
	for _, actionID := range dbActionIDs {
		err := persistIssueResolutionStat(tx, issue.ID, dbRootCauseID, actionID)
		if err != nil {
			return err
		}
	}

	return nil
}

// persistIssueResolutionStat persist a single selected tuple of rootcause/action in the resolution statistics table
func persistIssueResolutionStat(tx *sqlx.Tx, issueID int64, rootCauseID int64, actionID int64) error {
	query := `INSERT into issue_resolution_v1 (feedback_date, issue_id, rootcause_id, action_id) 
		values (:feedback_date, :issue_id, :rootcause_id, :action_id)`
	params := map[string]interface{}{
		"feedback_date": time.Now().UTC(),
		"issue_id":      issueID,
		"rootcause_id":  rootCauseID,
		"action_id":     actionID,
	}

	var err error
	if tx != nil {
		_, err = tx.NamedExec(query, params)
	} else {
		_, err = postgres.DB().NamedExec(query, params)
	}
	if err != nil {
		return errors.New("couldn't query the database:" + err.Error())
	}
	return nil
}
