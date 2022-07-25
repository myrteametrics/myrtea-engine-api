package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/explainer/issues"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"go.uber.org/zap"
)

var allowedSortByFields = []string{"id", "created_at", "last_modified"}

// GetIssues godoc
// @Summary Get all issues
// @Description Get all issues
// @Tags Issues
// @Produce json
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 500 "internal server error"
// @Router /admin/engine/issues_all [get]
func GetIssues(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("No user found in context"))
		return
	}
	groups := GetUserGroupsFromUser(user)

	issues, err := issues.R().GetAll(groups)
	if err != nil {
		zap.L().Error("Cannot retrieve issues", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	render.JSON(w, r, issues)
}

// GetIssuesByStatesByPage godoc
// @Summary Get issues by issues states (paginated)
// @Description Get issues by issues states (paginated)
// @Tags Issues
// @Produce json
// @Param states query string true "Issue states (comma separated) (Available: open, draft, closedfeedback, closednofeedback, closedtimeout)"
// @Param limit query string false "Result limit (default: 50)"
// @Param offset query string false "Result offset (default: 0)"
// @Param sort_by query string false "Result offset (example: 'sort_by=desc(last_modified),asc(id)')"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 500 "internal server error"
// @Router /engine/issues [get]
func GetIssuesByStatesByPage(w http.ResponseWriter, r *http.Request) {

	user := GetUserFromContext(r)
	if user == nil {
		render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("No user found in context"))
		return
	}
	groups := GetUserGroupsFromUser(user)

	var err error
	var limit int
	var offset int
	var sortOptions = make([]models.SortOption, 0)

	states := strings.Split(r.URL.Query().Get("states"), ",")

	if rawSize := r.URL.Query().Get("limit"); rawSize != "" {
		limit, err = ParseInt(rawSize)
		if err != nil {
			zap.L().Warn("Parse input limit", zap.Error(err), zap.String("rawNhit", rawSize))
			render.Error(w, r, render.ErrAPIParsingInteger, err)
			return
		}
	}

	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		offset, err = ParseInt(rawOffset)
		if err != nil {
			zap.L().Warn("Parse input offset", zap.Error(err), zap.String("raw offset", rawOffset))
			render.Error(w, r, render.ErrAPIParsingInteger, err)
			return
		}
	}

	if rawSortBy := r.URL.Query().Get("sort_by"); rawSortBy != "" {
		sortOptions, err = ParseSortBy(rawSortBy, allowedSortByFields)
		if err != nil {
			zap.L().Warn("Parse input sort_by", zap.Error(err), zap.String("raw sort_by", rawSortBy))
			render.Error(w, r, render.ErrAPIParsingSortBy, err)
			return
		}
	}

	searchOptions := models.SearchOptions{
		Limit:  limit,
		Offset: offset,
		SortBy: sortOptions,
	}

	issuesSlice, total, err := issues.R().GetByStateByPage(states, searchOptions, groups)
	if err != nil {
		zap.L().Error("Error on getting issues", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	paginatedResource := models.PaginatedResource{
		Total: total,
		Items: issuesSlice,
	}

	render.JSON(w, r, paginatedResource)
}

// GetIssue godoc
// @Summary Get an issue
// @Description Get an issue
// @Tags Issues
// @Produce json
// @Param id path string true "Issue ID"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/issues/{id} [get]
func GetIssue(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("No user found in context"))
		return
	}
	groups := GetUserGroupsFromUser(user)

	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	issue, found, err := issues.R().Get(idIssue, groups)

	if err != nil {
		zap.L().Error("Cannot retrieve issue", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("issue does not exists", zap.String("issueID", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, issue)
}

// GetIssueHistory godoc
// @Summary Get an issue history
// @Description Get an issue history
// @Tags Issues
// @Produce json
// @Param id path string true "Issue ID"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/issues/{id}/history [get]

func GetIssueHistory(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("No user found in context"))
		return
	}
	groups := GetUserGroupsFromUser(user)

	var err error
	var limit int
	var offset int
	var sortOptions = make([]models.SortOption, 0)
	
	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	issue, found, err := issues.R().Get(idIssue, groups)
	if err != nil {
		zap.L().Error("Cannot retrieve issue", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("issue does not exists", zap.String("issueID", id))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	if rawSize := r.URL.Query().Get("limit"); rawSize != "" {
		limit, err = ParseInt(rawSize)
		if err != nil {
			zap.L().Warn("Parse input limit", zap.Error(err), zap.String("rawNhit", rawSize))
			render.Error(w, r, render.ErrAPIParsingInteger, err)
			return
		}
	}

	if rawOffset := r.URL.Query().Get("offset"); rawOffset != "" {
		offset, err = ParseInt(rawOffset)
		if err != nil {
			zap.L().Warn("Parse input offset", zap.Error(err), zap.String("raw offset", rawOffset))
			render.Error(w, r, render.ErrAPIParsingInteger, err)
			return
		}
	}

	if rawSortBy := r.URL.Query().Get("sort_by"); rawSortBy != "" {
		sortOptions, err = ParseSortBy(rawSortBy, allowedSortByFields)
		if err != nil {
			zap.L().Warn("Parse input sort_by", zap.Error(err), zap.String("raw sort_by", rawSortBy))
			render.Error(w, r, render.ErrAPIParsingSortBy, err)
			return
		}
	}

	searchOptions := models.SearchOptions{

		Limit:  limit,
		Offset: offset,
		SortBy: sortOptions,
	}

	issuesSlice, total, err := issues.R().GetByKey(issue.Key, searchOptions, groups)
	if err != nil {
		zap.L().Error("Error on getting issues", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	paginatedResource := models.PaginatedResource{
		Total: total,
		Items: issuesSlice,
	}

	render.JSON(w, r, paginatedResource)
}

// GetIssueFactsHistory godoc
// @Summary Get the facts history for an issue
// @Description Get the facts history for an issue
// @Tags Issues
// @Produce json
// @Param id path string true "Issue ID"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 500 "Status Internal Server Error"
// @Failure 404 "Status Not Found"
// @Router /engine/issues/{id}/facts_history [get]
func GetIssueFactsHistory(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("No user found in context"))
		return
	}
	groups := GetUserGroupsFromUser(user)

	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	history, found, err := explainer.GetFactsHistory(idIssue, groups)
	if err != nil {
		zap.L().Error("An error has occurred", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Not found", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, history)
}

// PostIssue godoc
// @Summary Creates an issue
// @Description Creates an issue
// @Tags Issues
// @Accept json
// @Produce json
// @Param issue body interface{} true "Issue (json)"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/issues [post]
func PostIssue(w http.ResponseWriter, r *http.Request) {
	var newIssue models.Issue
	err := json.NewDecoder(r.Body).Decode(&newIssue)
	if err != nil {
		zap.L().Warn("Invalid issue json defined", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	//TODO: No need to return the new issue id ?
	_, err = issues.R().Create(newIssue)
	if err != nil {
		zap.L().Error("Error while creating the issue", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	render.OK(w, r)
}

// GetIssueFeedbackTree godoc
// @Summary Generate the rootcauses/actions recommendation tree
// @Description Generate the rootcauses/actions recommendation tree for an issue
// @Tags Issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Security Bearer
// @Success 200 {object} models.FrontRecommendation "recommendation"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/issues/{id}/recommendation [get]
func GetIssueFeedbackTree(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	tree, err := explainer.GetRecommendationTree(idIssue)
	if err != nil {
		zap.L().Error("Generating rootcauses / actions tree", zap.Int64("id", idIssue), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	render.JSON(w, r, tree)
}

// PostIssueDraft godoc
// @Summary Send a rootcauses/actions feedback draft on an issue
// @Description Post a rootcauses/actions recommendation tree as a feedback draft on an issue
// @Tags Issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Param issue body models.FrontRecommendation true "Draft Recommendation tree (json)"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/issues/{id}/draft [post]
func PostIssueDraft(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("No user found in context"))
		return
	}
	groups := GetUserGroupsFromUser(user)

	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	var newDraft models.FrontRecommendation
	err = json.NewDecoder(r.Body).Decode(&newDraft)
	if err != nil {
		zap.L().Warn("Body decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	err = explainer.SaveIssueDraft(nil, idIssue, newDraft, groups, *user)
	if err != nil {
		zap.L().Error("SaveIssueDraft", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	render.OK(w, r)
}

// PostIssueCloseWithFeedback godoc
// @Summary Send a rootcauses/actions feedback on an issue
// @Description Post a rootcauses/actions recommendation tree as a feedback on an issue
// @Tags Issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Param issue body interface{} true "Recommendation tree (json)"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/issues/{id}/feedback [post]
func PostIssueCloseWithFeedback(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("No user found in context"))
		return
	}
	groups := GetUserGroupsFromUser(user)

	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	var newFeedback models.FrontRecommendation
	err = json.NewDecoder(r.Body).Decode(&newFeedback)
	if err != nil {
		zap.L().Warn("Body decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	err = explainer.CloseIssueWithFeedback(postgres.DB(), idIssue, newFeedback, groups, *user)
	if err != nil {
		zap.L().Error("CloseIssueWithFeedback", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	render.OK(w, r)
}

// PostIssueCloseWithoutFeedback godoc
// @Summary Close an issue without feedback
// @Description Close an issue without feedback
// @Tags Issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Param reason body interface{} false "Close reason (json)"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/issues/{id}/close [post]
func PostIssueCloseWithoutFeedback(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("No user found in context"))
		return
	}
	groups := GetUserGroupsFromUser(user)

	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	reason := struct {
		S string `json:"reason"`
	}{
		S: "unknown",
	}
	if r.Body != nil && r.Body != http.NoBody {
		err = json.NewDecoder(r.Body).Decode(&reason)
		if err != nil {
			zap.L().Warn("Body decode", zap.Error(err))
			render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
			return
		}
	}

	err = explainer.CloseIssueWithoutFeedback(postgres.DB(), idIssue, reason.S, groups, *user)
	if err != nil {
		zap.L().Error("CloseIssueWithoutFeedback", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	render.OK(w, r)
}

// PostIssueDetectionFeedback godoc
// @Summary Add a new detection feedback
// @Description Add a new detection feedback on an issue (or replace an existing one if the user already made a feedback)
// @Tags Issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Param reason body interface{} false "Rating"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/issues/{id}/detection/feedback [post]
func PostIssueDetectionFeedback(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r)
	if user == nil {
		render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("No user found in context"))
		return
	}
	groups := GetUserGroupsFromUser(user)

	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
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
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	err = explainer.AddIssueDetectionFeedback(postgres.DB(), idIssue, user.ID, feedback.Rating, groups)
	if err != nil {
		zap.L().Warn("AddIssueDetectionFeedback", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	render.OK(w, r)
}

// UpdateIssueComment godoc
// @Summary Update an issue comment
// @Description Update an issue comment
// @Tags Issues
// @Accept json
// @Produce json
// @Param id path string true "Issue ID"
// @Param reason body interface{} false "Comment to update"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/issues/{id}/comment [put]
func UpdateIssueComment(w http.ResponseWriter, r *http.Request) {

	user := GetUserFromContext(r)
	if user == nil {
		render.Error(w, r, render.ErrAPISecurityMissingContext, errors.New("No user found in context"))
		return
	}

	id := chi.URLParam(r, "id")
	idIssue, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing issue id", zap.String("issueID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	type IssueComment struct {
		Comment string `json:"comment"`
	}

	var comment IssueComment
	err = json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		zap.L().Warn("Body decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	//zap.L().Info("UpdateComment", zap.String("comment", comment.Comment))

	err = issues.R().UpdateComment(postgres.DB(), idIssue, comment.Comment, *user)
	if err != nil {
		zap.L().Error("Cannot update issue comment", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	render.OK(w, r)
}
