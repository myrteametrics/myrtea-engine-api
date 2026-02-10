package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/scheduler"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
)

// GetBoostList godoc
//
//	@Id				GetBoostList
//	@Summary		Get the list of jobs to boost
//	@Description	Returns all jobs that need to be boosted
//	@Tags			Service
//	@Produce		json
//	@Success		200	{array}		scheduler.BoostAction
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Router			/service/boost/list [get]
func GetBoostList(w http.ResponseWriter, r *http.Request) {
	list := scheduler.BM().GetBoostList()
	httputil.JSON(w, r, list)
}

// GetRevertList godoc
//
//	@Id				GetRevertList
//	@Summary		Get the list of jobs to revert
//	@Description	Returns all jobs that need to revert to normal frequency
//	@Tags			Service
//	@Produce		json
//	@Success		200	{array}		scheduler.BoostAction
//	@Failure		403	{object}	httputil.APIError	"Forbidden"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Router			/service/boost/revert [get]
func GetRevertList(w http.ResponseWriter, r *http.Request) {

	list := scheduler.BM().GetRevertList()
	httputil.JSON(w, r, list)
}

// AcknowledgeBoost godoc
//
//	@Id				AcknowledgeBoost
//	@Summary		Acknowledge a boost action
//	@Description	Marks a boost action as read for the given job ID
//	@Tags			Service
//	@Param			jobId	path	string	true	"Job ID"
//	@Success		200		"Status OK"
//	@Failure		400		{object}	httputil.APIError	"Bad Request"
//	@Failure		403		{object}	httputil.APIError	"Forbidden"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Router			/service/boost/{jobId}/ack [post]
func AcknowledgeBoost(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing jobId"))
		return
	}

	err := scheduler.BM().AcknowledgeBoost(jobID)
	if err != nil {
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.OK(w, r)
}

// AcknowledgeRevert godoc
//
//	@Id				AcknowledgeRevert
//	@Summary		Acknowledge a revert action
//	@Description	Marks a revert action as read for the given job ID
//	@Tags			Service
//	@Param			jobId	path	string	true	"Job ID"
//	@Success		200		"Status OK"
//	@Failure		400		{object}	httputil.APIError	"Bad Request"
//	@Failure		403		{object}	httputil.APIError	"Forbidden"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Router			/service/boost/revert/{jobId}/ack [post]
func AcknowledgeRevert(w http.ResponseWriter, r *http.Request) {

	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New("missing jobId"))
		return
	}

	err := scheduler.BM().AcknowledgeRevert(jobID)
	if err != nil {
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.OK(w, r)
}
