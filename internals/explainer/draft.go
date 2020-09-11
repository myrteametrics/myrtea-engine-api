package explainer

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/draft"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
)

// SaveIssueDraft generate and persist an issue draft
func SaveIssueDraft(tx *sqlx.Tx, issueID int64, issueDraft models.FrontDraft, groupList []int64, user groups.UserWithGroups) error {

	issue, found, err := issues.R().Get(issueID, groupList)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("Issue with id %d not found", issueID)
	}
	if issue.State.IsClosed() {
		return fmt.Errorf("Issue with id %d is already in a closed state", issueID)
	}

	exists, err := draft.R().CheckExists(nil, issueID)
	if err != nil {
		return err
	}

	switch issue.State {
	case models.Open:
		if exists {
			return fmt.Errorf("A draft has been found on an issue %d with state Open", issueID)
		}
		err = draft.R().Create(tx, issueID, issueDraft)
		if err != nil {
			return err
		}

	case models.Draft:
		existsWithUUID, err := draft.R().CheckExistsWithUUID(nil, issueID, issueDraft.ConcurrencyUUID)
		if err != nil {
			return err
		}
		if !existsWithUUID {
			return errors.New("The existing draft has already been modified by someone else")
		}
		err = draft.R().Update(tx, issueID, issueDraft)
		if err != nil {
			return err
		}
	}

	err = updateIssueState(tx, issueID, models.Draft, groupList, user)
	if err != nil {
		return err
	}

	return nil
}
