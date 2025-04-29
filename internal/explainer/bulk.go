package explainer

import (
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/users"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"github.com/myrteametrics/myrtea-sdk/v5/postgres"
	"go.uber.org/zap"
)

func DraftIssues(idIssues []int64, user users.User, status *model.DraftIssuesStatus) {
	for _, idIssue := range idIssues {
		issue, _, err := issues.R().Get(idIssue)
		if err != nil {
			zap.L().Error("Cannot retrieve issue", zap.Error(err), zap.Int64("Id Issues ", idIssue))
			DrafHandleError(status, idIssue, err, httputil.ErrAPIDBSelectFailed)
			continue
		}

		tree, err := GetRecommendationTree(issue)
		if err != nil {
			zap.L().Error("Generating rootcauses / actions tree", zap.Int64("id", issue.ID), zap.Error(err))
			DrafHandleError(status, idIssue, errors.New("recommendation tree based on issue resolution stats table"), httputil.ErrAPIDBSelectFailed)
			continue
		}

		err = SaveIssueDraft(nil, issue, *tree, user)
		if err != nil {
			zap.L().Error("SaveIssueDraft", zap.Error(err))
			DrafHandleError(status, idIssue, err, httputil.ErrAPIDBInsertFailed)
			continue
		}
		status.SuccessCount++
	}
}

func DrafHandleError(status *model.DraftIssuesStatus, idIssue int64, err error, apiError httputil.APIError) {
	status.AllOk = false
	status.ErrorMessages += fmt.Sprintf("ID Issue: %d, error: %s, Api_Error (Status %d, ErrType %s, Code %d, Message %s)\n",
		idIssue, err.Error(), apiError.Status, apiError.ErrType, apiError.Code, apiError.Message)
}

func CloseIssues(idIssues []int64, user users.User, status *model.CloseIssuesStatus) {
	for _, idIssue := range idIssues {
		issue, _, err := issues.R().Get(idIssue)
		if err != nil {
			zap.L().Error("Cannot retrieve issue", zap.Error(err), zap.Int64("IdIssues", idIssue))
			CloseHandleError(status, idIssue, err, httputil.ErrAPIDBSelectFailed)
			continue
		}

		//reason := model.Reason{S: "unknown"}
		err = CloseIssueWithoutFeedback(postgres.DB(), issue, user, model.ClosedNoFeedback)
		if err != nil {
			zap.L().Error("CloseIssueWithoutFeedback", zap.Error(err))
			CloseHandleError(status, idIssue, err, httputil.ErrAPIDBUpdateFailed)
		}
		status.SuccessCount++
	}
}
func CloseHandleError(status *model.CloseIssuesStatus, idIssue int64, err error, apiError httputil.APIError) {
	status.AllOk = false
	status.ErrorMessages += fmt.Sprintf("ID Issue: %d, error: %s, Api_Error (Status %d, ErrType %s, Code %d, Message %s)\n",
		idIssue, err.Error(), apiError.Status, apiError.ErrType, apiError.Code, apiError.Message)
}
