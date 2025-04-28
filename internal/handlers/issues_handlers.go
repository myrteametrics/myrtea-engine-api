package handlers

import (
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
	"github.com/myrteametrics/myrtea-sdk/v5/postgres"
	"go.uber.org/zap"
)

var allowedSortByFields = []string{"id", "created_at", "last_modified"}

// GetIssues godoc
//
//	@Summary		Get all issues
//	@Description	Get all issues
//	@Tags			Issues
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		500	"internal server error"
//	@Router			/admin/engine/issues_all [get]
func GetIssues(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var issueList map[int64]models.Issue
	var err error
	if userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, permissions.All, permissions.ActionGet)) {
		issueList, err = issues.R().GetAll()
	} else {
		situationIDs := userCtx.GetMatchingResourceIDsInt64(permissions.New(permissions.TypeSituationIssues, permissions.All, permissions.ActionGet))
		issueList, err = issues.R().GetAllBySituationIDs(situationIDs)
	}
	if err != nil {
		zap.L().Error("Cannot retrieve issues", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, issueList)
}

// GetIssuesByStatesByPageUnProtected godoc
//
//	@Summary		Get issues by issues states (paginated)
//	@Description	Get issues by issues states (paginated)
//	@Tags			Issues
//	@Produce		json
//	@Param			states	query	string	true	"Issue states (comma separated) (Available: open, draft, closedfeedback, closednofeedback, closedtimeout)"
//	@Param			limit	query	string	false	"Result limit (default: 50)"
//	@Param			offset	query	string	false	"Result offset (default: 0)"
//	@Param			sort_by	query	string	false	"Result offset (example: 'sort_by=desc(last_modified),asc(id)')"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		500	"internal server error"
//	@Router			/engine/issues/unprotected [get]
func GetIssuesByStatesByPageUnProtected(w http.ResponseWriter, r *http.Request) {

	var err error
	var limit int
	var offset int
	var sortOptions = make([]models.SortOption, 0)

	states := strings.Split(r.URL.Query().Get("states"), ",")

	if rawSize := r.URL.Query().Get("limit"); rawSize != "" {
		limit, err = ParseInt(rawSize)
		if err != nil {
			zap.L().Warn("Parse input limit", zap.Error(err), zap.String("rawNhit", rawSize))
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
			return
		}
	}

	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		offset, err = ParseInt(rawOffset)
		if err != nil {
			zap.L().Warn("Parse input offset", zap.Error(err), zap.String("raw offset", rawOffset))
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
			return
		}
	}

	if rawSortBy := r.URL.Query().Get("sort_by"); rawSortBy != "" {
		sortOptions, err = ParseSortBy(rawSortBy, allowedSortByFields)
		if err != nil {
			zap.L().Warn("Parse input sort_by", zap.Error(err), zap.String("raw sort_by", rawSortBy))
			httputil.Error(w, r, httputil.ErrAPIParsingSortBy, err)
			return
		}
	}

	searchOptions := models.SearchOptions{
		Limit:  limit,
		Offset: offset,
		SortBy: sortOptions,
	}

	var issuesSlice []models.Issue
	var total int
	issuesSlice, total, err = issues.R().GetByStateByPage(states, searchOptions)

	if err != nil {
		zap.L().Error("Error on getting issues", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	paginatedResource := models.PaginatedResource{
		Total: total,
		Items: issuesSlice,
	}

	httputil.JSON(w, r, paginatedResource)
}

// GetIssuesByStatesByPage godoc
//
//	@Summary		Get issues by issues states (paginated)
//	@Description	Get issues by issues states (paginated)
//	@Tags			Issues
//	@Produce		json
//	@Param			states	query	string	true	"Issue states (comma separated) (Available: open, draft, closedfeedback, closednofeedback, closedtimeout)"
//	@Param			limit	query	string	false	"Result limit (default: 50)"
//	@Param			offset	query	string	false	"Result offset (default: 0)"
//	@Param			sort_by	query	string	false	"Result offset (example: 'sort_by=desc(last_modified),asc(id)')"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		500	"internal server error"
//	@Router			/engine/issues [get]
func GetIssuesByStatesByPage(w http.ResponseWriter, r *http.Request) {

	var err error
	var limit int
	var offset int
	var sortOptions = make([]models.SortOption, 0)

	states := strings.Split(r.URL.Query().Get("states"), ",")

	if rawSize := r.URL.Query().Get("limit"); rawSize != "" {
		limit, err = ParseInt(rawSize)
		if err != nil {
			zap.L().Warn("Parse input limit", zap.Error(err), zap.String("rawNhit", rawSize))
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
			return
		}
	}

	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		offset, err = ParseInt(rawOffset)
		if err != nil {
			zap.L().Warn("Parse input offset", zap.Error(err), zap.String("raw offset", rawOffset))
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
			return
		}
	}

	if rawSortBy := r.URL.Query().Get("sort_by"); rawSortBy != "" {
		sortOptions, err = ParseSortBy(rawSortBy, allowedSortByFields)
		if err != nil {
			zap.L().Warn("Parse input sort_by", zap.Error(err), zap.String("raw sort_by", rawSortBy))
			httputil.Error(w, r, httputil.ErrAPIParsingSortBy, err)
			return
		}
	}

	searchOptions := models.SearchOptions{
		Limit:  limit,
		Offset: offset,
		SortBy: sortOptions,
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var issuesSlice []models.Issue
	var total int
	if userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, permissions.All, permissions.ActionGet)) {
		issuesSlice, total, err = issues.R().GetByStateByPage(states, searchOptions)
	} else {
		situationIDs := userCtx.GetMatchingResourceIDsInt64(permissions.New(permissions.TypeSituationIssues, permissions.All, permissions.ActionGet))
		issuesSlice, total, err = issues.R().GetByStateByPageBySituationIDs(states, searchOptions, situationIDs)
	}
	if err != nil {
		zap.L().Error("Error on getting issues", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	paginatedResource := models.PaginatedResource{
		Total: total,
		Items: issuesSlice,
	}

	httputil.JSON(w, r, paginatedResource)
}

// GetIssue godoc
//
//	@Summary		Get an issue
//	@Description	Get an issue
//	@Tags			Issues
//	@Produce		json
//	@Param			id	path	string	true	"Issue ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/issues/{id} [get]
func GetIssue(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	issue, found, err := issues.R().Get(idIssue)
	if err != nil {
		zap.L().Error("Cannot retrieve issue", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("issue does not exists", zap.String("issueID", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, strconv.FormatInt(issue.SituationID, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	httputil.JSON(w, r, issue)
}

// GetIssueHistory godoc
//
//	@Summary		Get an issue history
//	@Description	Get an issue history
//	@Tags			Issues
//	@Produce		json
//	@Param			id	path	string	true	"Issue ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/issues/{id}/history [get]
func GetIssueHistory(w http.ResponseWriter, r *http.Request) {

	var err error
	var limit int
	var offset int
	var sortOptions = make([]models.SortOption, 0)

	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	if rawSize := r.URL.Query().Get("limit"); rawSize != "" {
		limit, err = ParseInt(rawSize)
		if err != nil {
			zap.L().Warn("Parse input limit", zap.Error(err), zap.String("rawNhit", rawSize))
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
			return
		}
	}

	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		offset, err = ParseInt(rawOffset)
		if err != nil {
			zap.L().Warn("Parse input offset", zap.Error(err), zap.String("raw offset", rawOffset))
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
			return
		}
	}

	if rawSortBy := r.URL.Query().Get("sort_by"); rawSortBy != "" {
		sortOptions, err = ParseSortBy(rawSortBy, allowedSortByFields)
		if err != nil {
			zap.L().Warn("Parse input sort_by", zap.Error(err), zap.String("raw sort_by", rawSortBy))
			httputil.Error(w, r, httputil.ErrAPIParsingSortBy, err)
			return
		}
	}

	searchOptions := models.SearchOptions{
		Limit:  limit,
		Offset: offset,
		SortBy: sortOptions,
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	issue, found, err := issues.R().Get(idIssue)
	if err != nil {
		zap.L().Error("Cannot retrieve issue", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("issue does not exists", zap.String("issueID", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	issuesSlice, total, err := issues.R().GetByKeyByPage(issue.Key, searchOptions)
	if err != nil {
		zap.L().Error("Error on getting issues", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	paginatedResource := models.PaginatedResource{
		Total: total,
		Items: issuesSlice,
	}

	httputil.JSON(w, r, paginatedResource)
}

// GetIssueFactsHistory godoc
//
//	@Summary		Get the facts history for an issue
//	@Description	Get the facts history for an issue
//	@Tags			Issues
//	@Produce		json
//	@Param			id	path	string	true	"Issue ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		500	"Status Internal Server Error"
//	@Failure		404	"Status Not Found"
//	@Router			/engine/issues/{id}/facts_history [get]
func GetIssueFactsHistory(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	issue, found, err := issues.R().Get(idIssue)
	if err != nil {
		zap.L().Error("Cannot retrieve issue", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("issue does not exists", zap.String("issueID", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, strconv.FormatInt(issue.SituationID, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	history, found, err := explainer.GetFactsHistory(issue)
	if err != nil {
		zap.L().Error("An error has occurred", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Not found", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, history)
}

// PostIssue godoc
//
//	@Summary		Creates an issue
//	@Description	Creates an issue
//	@Tags			Issues
//	@Accept			json
//	@Produce		json
//	@Param			issue	body	interface{}	true	"Issue (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/issues [post]
func PostIssue(w http.ResponseWriter, r *http.Request) {
	var newIssue models.Issue
	err := json.NewDecoder(r.Body).Decode(&newIssue)
	if err != nil {
		zap.L().Warn("Invalid issue json defined", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, strconv.FormatInt(newIssue.SituationID, 10), permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	//TODO: No need to return the new issue id ?
	_, err = issues.R().Create(newIssue)
	if err != nil {
		zap.L().Error("Error while creating the issue", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	httputil.OK(w, r)
}

// GetIssueFeedbackTree godoc
//
//	@Summary		Generate the rootcauses/actions recommendation tree
//	@Description	Generate the rootcauses/actions recommendation tree for an issue
//	@Tags			Issues
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Issue ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	models.FrontRecommendation	"recommendation"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/issues/{id}/recommendation [get]
func GetIssueFeedbackTree(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	issue, found, err := issues.R().Get(idIssue)
	if err != nil {
		zap.L().Error("Cannot retrieve issue", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("issue does not exists", zap.String("issueID", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, strconv.FormatInt(issue.SituationID, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	tree, err := explainer.GetRecommendationTree(issue)
	if err != nil {
		zap.L().Error("Generating rootcauses / actions tree", zap.Int64("id", issue.ID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, tree)
}

// PostIssueDraft godoc
//
//	@Summary		Send a rootcauses/actions feedback draft on an issue
//	@Description	Post a rootcauses/actions recommendation tree as a feedback draft on an issue
//	@Tags			Issues
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string						true	"Issue ID"
//	@Param			issue	body	models.FrontRecommendation	true	"Draft Recommendation tree (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/issues/{id}/draft [post]
func PostIssueDraft(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	issue, found, err := issues.R().Get(idIssue)
	if err != nil {
		zap.L().Error("Cannot retrieve issue", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("issue does not exists", zap.String("issueID", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, strconv.FormatInt(issue.SituationID, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newDraft models.FrontRecommendation
	err = json.NewDecoder(r.Body).Decode(&newDraft)
	if err != nil {
		zap.L().Warn("Body decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	err = explainer.SaveIssueDraft(nil, issue, newDraft, userCtx.User)
	if err != nil {
		zap.L().Error("SaveIssueDraft", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	httputil.OK(w, r)
}

// PostIssuesDraft godoc
//
//	@Summary		Send a rootcauses/actions feedback draft on many issues
//	@Description	Post a rootcauses/actions recommendation tree as a feedback draft on many issues
//	@Tags			Issues
//	@Accept			json
//	@Produce		json
//	@Param			issue	body	models.IssuesIdsToDraf	true	"Issues IDs"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/issues/draft [post]
func PostIssuesDraft(w http.ResponseWriter, r *http.Request) {
	var issueIdsToDraft models.IssuesIdsToDraf
	err := json.NewDecoder(r.Body).Decode(&issueIdsToDraft)
	if err != nil {
		zap.L().Warn("Body decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	status := &models.DraftIssuesStatus{}
	var idIssuesOk []int64
	userCtx, _ := GetUserFromContext(r)

	for _, idIssue := range issueIdsToDraft.Ids {
		issue, found, err := issues.R().Get(idIssue)
		if err != nil {
			zap.L().Error("Cannot retrieve issue", zap.Error(err), zap.Int64("Id Issues ", idIssue))
			explainer.DrafHandleError(status, idIssue, err, httputil.ErrAPIDBSelectFailed)
			continue
		}
		if !found {
			zap.L().Warn("issue does not exist", zap.Int64("issueID", idIssue), zap.Int64("Id Issues ", idIssue))
			explainer.DrafHandleError(status, idIssue, errors.New("issue not found"), httputil.ErrAPIDBResourceNotFound)
			continue
		}
		if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, strconv.FormatInt(issue.SituationID, 10), permissions.ActionGet)) {
			explainer.DrafHandleError(status, idIssue, errors.New("missing permission"), httputil.ErrAPISecurityNoPermissions)
			continue
		}
		idIssuesOk = append(idIssuesOk, idIssue)
		status.SuccessCount++
	}

	explainer.DraftIssues(idIssuesOk, userCtx.User, status)

	if status.AllOk {
		httputil.OK(w, r)
		return
	}
	if status.SuccessCount == 0 {
		httputil.Error(w, r, httputil.ErrAPIProcessError, errors.New(status.ErrorMessages))
		return
	}
	httputil.Error(w, r, httputil.ErrAPIPartialSuccess, errors.New(status.ErrorMessages))
}

// PostIssueCloseWithFeedback godoc
//
//	@Summary		Send a rootcauses/actions feedback on an issue
//	@Description	Post a rootcauses/actions recommendation tree as a feedback on an issue
//	@Tags			Issues
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string		true	"Issue ID"
//	@Param			issue		body	interface{}	true	"Recommendation tree (json)"
//	@Param			isFakeAlert	path	bool		true	"Indicates if the closed issue was a real alert (true) or false positive (false)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/issues/{id}/feedback/{isFakeAlert} [post]
func PostIssueCloseWithFeedback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	isFakeAlertParam := chi.URLParam(r, "isFakeAlert")
	isFakeAlert, err := strconv.ParseBool(isFakeAlertParam)
	if err != nil {
		zap.L().Warn("Error on parsing isFakeAlert parameter", zap.String("issueID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	issue, found, err := issues.R().Get(idIssue)
	if err != nil {
		zap.L().Error("Cannot retrieve issue", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("issue does not exists", zap.String("issueID", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, strconv.FormatInt(issue.SituationID, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newFeedback models.FrontRecommendation
	err = json.NewDecoder(r.Body).Decode(&newFeedback)
	if err != nil {
		zap.L().Warn("Body decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	err = explainer.CloseIssueWithFeedback(postgres.DB(), issue, newFeedback, userCtx.User, isFakeAlert)
	if err != nil {
		zap.L().Error("CloseIssueWithFeedback", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	httputil.OK(w, r)
}

// PostIssueCloseWithoutFeedback godoc
//
//	@Summary		Close an issue without feedback
//	@Description	Close an issue without feedback
//	@Tags			Issues
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string		true	"Issue ID"
//	@Param			isFakeAlert	query	bool		true	"Indicates if the closed issue was a real alert (true) or false positive (false)"
//	@Param			reason		body	interface{}	false	"Close reason (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/issues/{id}/close [post]
func PostIssueCloseWithoutFeedback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	var targetState models.IssueState
	isFakeAlertParam := r.URL.Query().Get("isFakeAlert")

	if isFakeAlertParam == "" {
		targetState = models.ClosedNoFeedback
	} else {
		var parseErr error
		isFakeAlert, parseErr := strconv.ParseBool(isFakeAlertParam)
		if parseErr != nil {
			zap.L().Warn("Error on parsing isFakeAlert parameter", zap.String("issueID", id), zap.Error(parseErr))
			httputil.Error(w, r, httputil.ErrAPIParsingInteger, parseErr)
			return
		}

		if isFakeAlert {
			targetState = models.ClosedRejected
		} else {
			targetState = models.ClosedConfirmed
		}
	}

	issue, found, err := issues.R().Get(idIssue)
	if err != nil {
		zap.L().Error("Cannot retrieve issue", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("issue does not exists", zap.String("issueID", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	if issue.State.IsClosed() {
		zap.L().Warn("Issue with id is already in a closed state", zap.String("issueID", id))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, strconv.FormatInt(issue.SituationID, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = explainer.CloseIssueWithoutFeedback(postgres.DB(), issue, userCtx.User, targetState)
	if err != nil {
		zap.L().Error("CloseIssueWithoutFeedback", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	httputil.OK(w, r)
}

// PostIssueDetectionFeedback godoc
//
//	@Summary		Add a new detection feedback
//	@Description	Add a new detection feedback on an issue (or replace an existing one if the user already made a feedback)
//	@Tags			Issues
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string		true	"Issue ID"
//	@Param			reason	body	interface{}	false	"Rating"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/issues/{id}/detection/feedback [post]
func PostIssueDetectionFeedback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	issue, found, err := issues.R().Get(idIssue)
	if err != nil {
		zap.L().Error("Cannot retrieve issue", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("issue does not exists", zap.String("issueID", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeSituationIssues, strconv.FormatInt(issue.SituationID, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// IssueDetectionFeedback wrap a feedback rating on an issue detection
	type IssueDetectionFeedback struct {
		Rating int `json:"rating"`
	}

	var feedback IssueDetectionFeedback
	err = json.NewDecoder(r.Body).Decode(&feedback)
	if err != nil {
		zap.L().Warn("Body decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	err = explainer.AddIssueDetectionFeedback(postgres.DB(), issue, userCtx.User, feedback.Rating)
	if err != nil {
		zap.L().Warn("AddIssueDetectionFeedback", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	httputil.OK(w, r)
}

// UpdateIssueComment godoc
//
//	@Summary		Update an issue comment
//	@Description	Update an issue comment
//	@Tags			Issues
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string		true	"Issue ID"
//	@Param			reason	body	interface{}	false	"Comment to update"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/issues/{id}/comment [put]
func UpdateIssueComment(w http.ResponseWriter, r *http.Request) {

	// FIXME : UpdateIssueComment permissions
	// userCtx, _ := GetUserFromContext(r)
	// if user == nil {
	// 	render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("no user found in context"))
	// 	return
	// }
	// if !userCtx.HasPermission(permissions.New(permissions.TypeFact, permissions.All, permissions.ActionCreate)) {
	// 	render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("missing permission"))
	// 	return
	// }

	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	type IssueComment struct {
		Comment string `json:"comment"`
	}

	var comment IssueComment
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		zap.L().Warn("Body decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	//zap.L().Info("UpdateComment", zap.String("comment", comment.Comment))

	err = issues.R().UpdateComment(postgres.DB(), idIssue, comment.Comment)
	if err != nil {
		zap.L().Error("Cannot update issue comment", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.OK(w, r)
}
