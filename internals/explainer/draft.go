package explainer

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/explainer/draft"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
)

// SaveIssueDraft generate and persist an issue draft
func SaveIssueDraft(tx *sqlx.Tx, issue models.Issue, issueDraft models.FrontDraft, user users.User) error {

	if issue.State.IsClosed() {
		return fmt.Errorf("Issue with id %d is already in a closed state", issue.ID)
	}

	exists, err := draft.R().CheckExists(nil, issue.ID)
	if err != nil {
		return err
	}

	switch issue.State {
	case models.Open:
		if exists {
			return fmt.Errorf("A draft has been found on an issue %d with state Open", issue.ID)
		}
		err = draft.R().Create(tx, issue.ID, issueDraft)
		if err != nil {
			return err
		}

	case models.Draft:
		existsWithUUID, err := draft.R().CheckExistsWithUUID(nil, issue.ID, issueDraft.ConcurrencyUUID)
		if err != nil {
			return err
		}
		if !existsWithUUID {
			return errors.New("the existing draft has already been modified by someone else")
		}
		err = draft.R().Update(tx, issue.ID, issueDraft)
		if err != nil {
			return err
		}
	}

	err = updateIssueState(tx, issue, models.Draft, user)
	if err != nil {
		return err
	}

	return nil
}
