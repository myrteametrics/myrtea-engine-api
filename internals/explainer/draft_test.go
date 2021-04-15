package explainer

import (
	"testing"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/draft"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
)

func TestSaveIssueDraftNotExists(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	newDraft := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Actions: []*models.FrontAction{
				{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false},
				{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false},
			}},
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: false, Custom: false, Actions: []*models.FrontAction{
				{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: true, Custom: false},
				{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: true, Custom: false},
			}},
		},
	}
	err := SaveIssueDraft(nil, models.Issue{ID: 99}, newDraft, users.User{})
	if err == nil {
		t.Error("Draft shoule not be saved on a non existing issue")
	}

	_, found, err := draft.R().Get(99)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if found {
		t.Error("Not draft should be found with id 99")
	}
}

func TestSaveIssueDraftNewOne(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	newDraft := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Actions: []*models.FrontAction{
				{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false},
				{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false},
			}},
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: false, Custom: false, Actions: []*models.FrontAction{
				{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: true, Custom: false},
				{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: true, Custom: false},
			}},
		},
	}

	issue, _, _ := issues.R().Get(1)
	err := SaveIssueDraft(nil, issue, newDraft, users.User{})
	if err != nil {
		t.Error(err)
	}
	_, found, err := draft.R().Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("A draft should have been saved with id 1")
	}
}

func TestSaveIssueDraftUpdateOne(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	newDraft := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Actions: []*models.FrontAction{
				{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false},
				{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false},
			}},
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: false, Custom: false, Actions: []*models.FrontAction{
				{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: true, Custom: false},
				{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: true, Custom: false},
			}},
		},
	}

	issue, _, _ := issues.R().Get(1)
	err := SaveIssueDraft(nil, issue, newDraft, users.User{})
	if err != nil {
		t.Error(err)
	}
	draftGetV1, found, err := draft.R().Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("A draft should have been saved with id 1")
	}

	issue, _, _ = issues.R().Get(1)

	newDraft.Tree[0].Selected = true
	newDraft.ConcurrencyUUID = draftGetV1.ConcurrencyUUID
	err = SaveIssueDraft(nil, issue, newDraft, users.User{})
	if err != nil {
		t.Error(err)
	}
	draftGetV2, found, err := draft.R().Get(1)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if !found {
		t.Error("A draft should have been saved with id 1")
	}
	if draftGetV2.ConcurrencyUUID == draftGetV1.ConcurrencyUUID {
		t.Error("Draft concurrencyID not properly updated")
	}
	if !draftGetV2.Tree[0].Selected {
		t.Error("Draft tree not properly updated")
	}
}

func TestSaveIssueDraftUpdateWithoutConcurrencUUID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgresql test in short mode")
	}
	db := tests.DBClient(t)
	defer dbDestroy(db, t)
	dbInit(db, t)

	newDraft := models.FrontRecommendation{
		Tree: []*models.FrontRootCause{
			{ID: 1, Name: "rc_1", Description: "rc_desc_1", Selected: false, Custom: false, Actions: []*models.FrontAction{
				{ID: 1, Name: "action_1", Description: "action_desc_1", Selected: false, Custom: false},
				{ID: 2, Name: "action_2", Description: "action_desc_2", Selected: false, Custom: false},
			}},
			{ID: 2, Name: "rc_2", Description: "rc_desc_2", Selected: false, Custom: false, Actions: []*models.FrontAction{
				{ID: 3, Name: "action_3", Description: "action_desc_3", Selected: true, Custom: false},
				{ID: 4, Name: "action_4", Description: "action_desc_4", Selected: true, Custom: false},
			}},
		},
	}

	issue, _, _ := issues.R().Get(1)

	// Creation
	err := SaveIssueDraft(nil, issue, newDraft, users.User{})
	if err != nil {
		t.Error("Draft should have been saved")
	}

	// Update (without concurrencyUUID)
	newDraft.Tree[0].Selected = true
	err = SaveIssueDraft(nil, issue, newDraft, users.User{})
	if err == nil {
		t.Error("Draft update cannot be saved without concurrencyUUID")
	}
}
