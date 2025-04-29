package handler

import (
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/scheduler"
	"go.uber.org/zap"
)

// StartScheduler godoc
//
//	@Summary		Start the scheduler
//	@Description	Start the fact scheduler
//	@Tags			Scheduler
//	@Success		200	"Status OK"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Router			/engine/scheduler/start [POST]
func StartScheduler(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeScheduler, permissions.All, permissions.All)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	scheduler.S().C.Start()
	httputil.OK(w, r)
}

// TriggerJobSchedule godoc
//
//	@Summary		Force facts calculation pipeline
//	@Description	<b>Force facts calculation pipeline</b>
//	@Description	Example :
//	@Description	<pre>{"jobtype":"fact","job":{"facts_ids":["fact_1","fact_2"]}}
//	@Description	{"jobtype":"baseline","job":{"baselines":{"3":["by_day","by_day_week"]}}}</pre>
//	@Tags			Scheduler
//	@Produce		json
//	@Param			job	body	scheduler.InternalSchedule	true	"JobSchedule definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/scheduler/trigger [post]
func TriggerJobSchedule(w http.ResponseWriter, r *http.Request) {

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeScheduler, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newSchedule scheduler.InternalSchedule
	err := json.NewDecoder(r.Body).Decode(&newSchedule)
	if err != nil {
		zap.L().Warn("Job schedule json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newSchedule.IsValid(); !ok {
		zap.L().Warn("Schedule is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	newSchedule.Job.Run()

	httputil.OK(w, r)
}

// GetJobSchedules godoc
//
//	@Summary		Get all JobSchedules
//	@Description	Get all JobSchedules from scheduler repository
//	@Tags			Scheduler
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	scheduler.InternalSchedule	"list of schedules"
//	@Failure		500	"internal server error"
//	@Router			/engine/scheduler/jobs [get]
func GetJobSchedules(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeScheduler, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	schedules, err := scheduler.R().GetAll()
	if err != nil {
		zap.L().Error("Cannot get schedules", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	schedulesSlice := make([]scheduler.InternalSchedule, 0)
	for _, schedule := range schedules {
		schedulesSlice = append(schedulesSlice, schedule)
	}

	sort.SliceStable(schedulesSlice, func(i, j int) bool {
		return schedulesSlice[i].ID < schedulesSlice[j].ID
	})

	httputil.JSON(w, r, schedulesSlice)
}

// GetJobSchedule godoc
//
//	@Summary		Get a JobSchedule
//	@Description	Get a specific JobSchedule by it's ID
//	@Tags			Scheduler
//	@Produce		json
//	@Param			id	path	string	true	"job ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	scheduler.InternalSchedule	"schedule"
//	@Failure		500	"internal server error"
//	@Router			/engine/scheduler/jobs/{id} [get]
func GetJobSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idJob, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Parsing JobSchedule id", zap.String("JobScheduleID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeScheduler, strconv.FormatInt(idJob, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	jobSchedule, found, err := scheduler.R().Get(idJob)
	if err != nil {
		zap.L().Error("Get JobSchedule from repository", zap.Int64("id", idJob), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Job not found", zap.Int64("id", idJob))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, jobSchedule)
}

// ValidateJobSchedule godoc
//
//	@Summary		validate a new JobSchedule definition
//	@Description	validate a new JobSchedule definition
//	@Tags			Scheduler
//	@Accept			json
//	@Produce		json
//	@Param			job	body	scheduler.InternalSchedule	true	"JobSchedule definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	scheduler.InternalSchedule	"schedule"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/scheduler/jobs/validate [post]
func ValidateJobSchedule(w http.ResponseWriter, r *http.Request) {
	var newSchedule scheduler.InternalSchedule
	err := json.NewDecoder(r.Body).Decode(&newSchedule)
	if err != nil {
		zap.L().Warn("Job schedule json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newSchedule.IsValid(); !ok {
		zap.L().Warn("Schedule is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	httputil.JSON(w, r, newSchedule)
}

// PostJobSchedule godoc
//
//	@Summary		create JobSchedule
//	@Description	creates new JobSchedule
//	@Tags			Scheduler
//	@Accept			json
//	@Produce		json
//	@Param			job	body	scheduler.InternalSchedule	true	"JobSchedule definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	scheduler.InternalSchedule	"schedule"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status Internal Server Error"
//	@Router			/engine/scheduler/jobs [post]
func PostJobSchedule(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeScheduler, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newSchedule scheduler.InternalSchedule
	err := json.NewDecoder(r.Body).Decode(&newSchedule)
	if err != nil {
		zap.L().Warn("Job schedule json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newSchedule.IsValid(); !ok {
		zap.L().Warn("Schedule is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	idJob, err := scheduler.R().Create(newSchedule)
	if err != nil {
		zap.L().Error("Error while creating JobSchedule", zap.Int64("JobSchedule.ID", newSchedule.ID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	jobSchedule, found, err := scheduler.R().Get(idJob)
	if err != nil {
		zap.L().Error("Get JobSchedule from repository", zap.Int64("id", idJob), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Job not found after creation", zap.Int64("id", idJob))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	err = scheduler.S().AddJobSchedule(jobSchedule)
	if err != nil {
		zap.L().Error("Error while adding JobSchedule to the scheduler", zap.Int64("JobSchedule.ID", jobSchedule.ID), zap.Error(err))

		err := scheduler.R().Delete(idJob)
		if err != nil {
			zap.L().Error("Error while rollbacking JobSchedule creation", zap.Int64("JobSchedule.ID", jobSchedule.ID), zap.Error(err))
		}

		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	httputil.JSON(w, r, jobSchedule)
}

// PutJobSchedule godoc
//
//	@Summary		Create or remplace a JobSchedule
//	@Description	Create or remplace a JobSchedule
//	@Tags			Scheduler
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string		true	"JobSchedule ID"
//	@Param			rule	body	interface{}	true	"JobSchedule (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	scheduler.InternalSchedule	"schedule"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status"	internal	server	error"
//	@Router			/engine/scheduler/jobs/{id} [put]
func PutJobSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idJob, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Parsing JobSchedule id", zap.String("JobScheduleID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeScheduler, strconv.FormatInt(idJob, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newSchedule scheduler.InternalSchedule
	err = json.NewDecoder(r.Body).Decode(&newSchedule)
	if err != nil {
		zap.L().Warn("Job schedule json decode", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	newSchedule.ID = idJob

	if ok, err := newSchedule.IsValid(); !ok {
		zap.L().Warn("Schedule is invalid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	err = scheduler.R().Update(newSchedule)
	if err != nil {
		zap.L().Error("Error while updating JobSchedule ", zap.Int64("ID", newSchedule.ID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	jobSchedule, found, err := scheduler.R().Get(idJob)
	if err != nil {
		zap.L().Error("Get JobSchedule from repository", zap.Int64("id", idJob), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Job not found after creation", zap.Int64("id", idJob))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	err = scheduler.S().AddJobSchedule(jobSchedule)
	if err != nil {
		zap.L().Error("Error while updating JobSchedule", zap.Int64("ID", jobSchedule.ID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	httputil.JSON(w, r, jobSchedule)
}

// DeleteJobSchedule godoc
//
//	@Summary		delete JobSchedule
//	@Description	delete JobSchedule
//	@Tags			Scheduler
//	@Produce		json
//	@Param			id	path	string	true	"JobSchedule ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status Internal Server Error"
//	@Router			/engine/scheduler/jobs/{id} [delete]
func DeleteJobSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idJob, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Parsing JobSchedule id", zap.String("JobScheduleID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeScheduler, strconv.FormatInt(idJob, 10), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = scheduler.R().Delete(idJob)
	if err != nil {
		zap.L().Error("Delete DeleteJobSchedule", zap.String("ID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	scheduler.S().RemoveJobSchedule(idJob)

	httputil.OK(w, r)
}
