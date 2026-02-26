package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/scheduler"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
)

// GetJobBoostList godoc
//
//	@Id				GetJobBoostList
//	@Summary		Get the list of jobs to boost
//	@Description	Returns all jobs that need to be boosted
//	@Tags			Service
//	@Produce		json
//	@Success		200	{array}		scheduler.JobBoostAction
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Router			/service/jobs/boost/list [get]
func GetJobBoostList(w http.ResponseWriter, r *http.Request) {
	list := scheduler.JBM().GetJobBoostList()
	httputil.JSON(w, r, list)
}

// GetJobRevertList godoc
//
//	@Id				GetJobRevertList
//	@Summary		Get the list of jobs to revert
//	@Description	Returns all jobs that need to revert to normal frequency
//	@Tags			Service
//	@Produce		json
//	@Success		200	{array}		scheduler.JobBoostAction
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Router			/service/jobs/boost/revert [get]
func GetJobRevertList(w http.ResponseWriter, r *http.Request) {

	list := scheduler.JBM().GetJobRevertList()
	httputil.JSON(w, r, list)
}

// AcknowledgeJobBoost godoc
//
//	@Id				AcknowledgeJobBoost
//	@Summary		Acknowledge a boost action
//	@Description	Marks a boost action as read for the given job ID
//	@Tags			Service
//	@Param			jobId	path	string	true	"Job ID"
//	@Success		200		"Status OK"
//	@Failure		400		{object}	httputil.APIError	"Bad Request"
//	@Failure		403		{object}	httputil.APIError	"Forbidden"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Router			/service/jobs/boost/{jobId}/ack [post]
func AcknowledgeJobBoost(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing jobId"))
		return
	}

	err := scheduler.JBM().AcknowledgeJobBoost(jobID)
	if err != nil {
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.OK(w, r)
}

// AcknowledgeJobRevert godoc
//
//	@Id				AcknowledgeJobRevert
//	@Summary		Acknowledge a revert action
//	@Description	Marks a revert action as read for the given job ID
//	@Tags			Service
//	@Param			jobId	path	string	true	"Job ID"
//	@Success		200		"Status OK"
//	@Failure		400		{object}	httputil.APIError	"Bad Request"
//	@Failure		403		{object}	httputil.APIError	"Forbidden"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Router			/service/jobs/boost/revert/{jobId}/ack [post]
func AcknowledgeJobRevert(w http.ResponseWriter, r *http.Request) {

	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing jobId"))
		return
	}

	err := scheduler.JBM().AcknowledgeJobRevert(jobID)
	if err != nil {
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.OK(w, r)
}
