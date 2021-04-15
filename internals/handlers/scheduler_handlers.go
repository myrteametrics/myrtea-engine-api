package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/scheduler"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/security/permissions"
	"go.uber.org/zap"
)

// StartScheduler godoc
// @Summary Start the scheduler
// @Description Start the fact scheduler
// @Tags Scheduler
// @Success 200 "Status OK"
// @Security Bearer
// @Router /engine/scheduler/start [POST]
func StartScheduler(w http.ResponseWriter, r *http.Request) {
	user, _ := GetUserFromContext(r)
	if !user.HasPermission(permissions.New(permissions.TypeScheduler, permissions.All, permissions.All)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("Missing permission"))
		return
	}

	scheduler.S().C.Start()
	render.OK(w, r)
}

// TriggerJobSchedule godoc
// @Summary Force facts calculation pipeline
// @Description <b>Force facts calculation pipeline</b>
// @Description Example :
// @Description <pre>{"jobtype":"fact","job":{"facts_ids":["fact_1","fact_2"]}}
// @Description {"jobtype":"baseline","job":{"baselines":{"3":["by_day","by_day_week"]}}}</pre>
// @Tags Scheduler
// @Produce json
// @Param job body scheduler.InternalSchedule true "JobSchedule definition (json)"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Router /engine/scheduler/trigger [post]
func TriggerJobSchedule(w http.ResponseWriter, r *http.Request) {

	user, _ := GetUserFromContext(r)
	if !user.HasPermission(permissions.New(permissions.TypeScheduler, permissions.All, permissions.ActionCreate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("Missing permission"))
		return
	}

	var newSchedule scheduler.InternalSchedule
	err := json.NewDecoder(r.Body).Decode(&newSchedule)
	if err != nil {
		zap.L().Warn("Job schedule json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newSchedule.IsValid(); !ok {
		zap.L().Warn("Schedule is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	newSchedule.Job.Run()

	render.OK(w, r)
}

// GetJobSchedules godoc
// @Summary Get all JobSchedules
// @Description Get all JobSchedules from scheduler repository
// @Tags Scheduler
// @Produce json
// @Security Bearer
// @Success 200 {array} scheduler.InternalSchedule "list of schedules"
// @Failure 500 "internal server error"
// @Router /engine/scheduler/jobs [get]
func GetJobSchedules(w http.ResponseWriter, r *http.Request) {
	user, _ := GetUserFromContext(r)
	if !user.HasPermission(permissions.New(permissions.TypeScheduler, permissions.All, permissions.ActionList)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("Missing permission"))
		return
	}

	schedules, err := scheduler.R().GetAll()
	if err != nil {
		zap.L().Error("Cannot get schedules", zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}

	schedulesSlice := make([]scheduler.InternalSchedule, 0)
	for _, schedule := range schedules {
		schedulesSlice = append(schedulesSlice, schedule)
	}

	sort.SliceStable(schedulesSlice, func(i, j int) bool {
		return schedulesSlice[i].ID < schedulesSlice[j].ID
	})

	render.JSON(w, r, schedulesSlice)
}

// GetJobSchedule godoc
// @Summary Get a JobSchedule
// @Description Get a specific JobSchedule by it's ID
// @Tags Scheduler
// @Produce json
// @Param id path string true "job ID"
// @Security Bearer
// @Success 200 {object} scheduler.InternalSchedule "schedule"
// @Failure 500 "internal server error"
// @Router /engine/scheduler/jobs/{id} [get]
func GetJobSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idJob, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Parsing JobSchedule id", zap.String("JobScheduleID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	user, _ := GetUserFromContext(r)
	if !user.HasPermission(permissions.New(permissions.TypeScheduler, strconv.FormatInt(idJob, 10), permissions.ActionGet)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("Missing permission"))
		return
	}

	jobSchedule, found, err := scheduler.R().Get(idJob)
	if err != nil {
		zap.L().Error("Get JobSchedule from repository", zap.Int64("id", idJob), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Job not found", zap.Int64("id", idJob))
		render.Error(w, r, render.ErrAPIDBResourceNotFound, err)
		return
	}

	render.JSON(w, r, jobSchedule)
}

// ValidateJobSchedule godoc
// @Summary validate a new JobSchedule definition
// @Description validate a new JobSchedule definition
// @Tags Scheduler
// @Accept json
// @Produce json
// @Param job body scheduler.InternalSchedule true "JobSchedule definition (json)"
// @Security Bearer
// @Success 200 {object} scheduler.InternalSchedule "schedule"
// @Failure 400	"Status Bad Request"
// @Router /engine/scheduler/jobs/validate [post]
func ValidateJobSchedule(w http.ResponseWriter, r *http.Request) {
	var newSchedule scheduler.InternalSchedule
	err := json.NewDecoder(r.Body).Decode(&newSchedule)
	if err != nil {
		zap.L().Warn("Job schedule json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newSchedule.IsValid(); !ok {
		zap.L().Warn("Schedule is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	render.JSON(w, r, newSchedule)
}

// PostJobSchedule godoc
// @Summary create JobSchedule
// @Description creates new JobSchedule
// @Tags Scheduler
// @Accept json
// @Produce json
// @Param job body scheduler.InternalSchedule true "JobSchedule definition (json)"
// @Security Bearer
// @Success 200 {object} scheduler.InternalSchedule "schedule"
// @Failure 400	"Status Bad Request"
// @Failure 500	"Status Internal Server Error"
// @Router /engine/scheduler/jobs [post]
func PostJobSchedule(w http.ResponseWriter, r *http.Request) {
	user, _ := GetUserFromContext(r)
	if !user.HasPermission(permissions.New(permissions.TypeScheduler, permissions.All, permissions.ActionCreate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("Missing permission"))
		return
	}

	var newSchedule scheduler.InternalSchedule
	err := json.NewDecoder(r.Body).Decode(&newSchedule)
	if err != nil {
		zap.L().Warn("Job schedule json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newSchedule.IsValid(); !ok {
		zap.L().Warn("Schedule is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	idJob, err := scheduler.R().Create(newSchedule)
	if err != nil {
		zap.L().Error("Error while creating JobSchedule", zap.Int64("JobSchedule.ID", newSchedule.ID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBInsertFailed, err)
		return
	}

	jobSchedule, found, err := scheduler.R().Get(idJob)
	if err != nil {
		zap.L().Error("Get JobSchedule from repository", zap.Int64("id", idJob), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Job not found after creation", zap.Int64("id", idJob))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	err = scheduler.S().AddJobSchedule(jobSchedule)
	if err != nil {
		zap.L().Error("Error while adding JobSchedule to the scheduler", zap.Int64("JobSchedule.ID", jobSchedule.ID), zap.Error(err))

		err := scheduler.R().Delete(idJob)
		if err != nil {
			zap.L().Error("Error while rollbacking JobSchedule creation", zap.Int64("JobSchedule.ID", jobSchedule.ID), zap.Error(err))
		}

		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	render.JSON(w, r, jobSchedule)
}

// PutJobSchedule godoc
// @Summary Create or remplace a JobSchedule
// @Description Create or remplace a JobSchedule
// @Tags Scheduler
// @Accept json
// @Produce json
// @Param id path string true "JobSchedule ID"
// @Param rule body interface{} true "JobSchedule (json)"
// @Security Bearer
// @Success 200 {object} scheduler.InternalSchedule "schedule"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status" internal server error"
// @Router /engine/scheduler/jobs/{id} [put]
func PutJobSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idJob, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Parsing JobSchedule id", zap.String("JobScheduleID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	user, _ := GetUserFromContext(r)
	if !user.HasPermission(permissions.New(permissions.TypeScheduler, strconv.FormatInt(idJob, 10), permissions.ActionUpdate)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("Missing permission"))
		return
	}

	var newSchedule scheduler.InternalSchedule
	err = json.NewDecoder(r.Body).Decode(&newSchedule)
	if err != nil {
		zap.L().Warn("Job schedule json decode", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}
	newSchedule.ID = idJob

	if ok, err := newSchedule.IsValid(); !ok {
		zap.L().Warn("Schedule is invalid", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		return
	}

	err = scheduler.R().Update(newSchedule)
	if err != nil {
		zap.L().Error("Error while updating JobSchedule ", zap.Int64("ID", newSchedule.ID), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBUpdateFailed, err)
		return
	}

	jobSchedule, found, err := scheduler.R().Get(idJob)
	if err != nil {
		zap.L().Error("Get JobSchedule from repository", zap.Int64("id", idJob), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Job not found after creation", zap.Int64("id", idJob))
		render.Error(w, r, render.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	err = scheduler.S().AddJobSchedule(jobSchedule)
	if err != nil {
		zap.L().Error("Error while updating JobSchedule", zap.Int64("ID", jobSchedule.ID), zap.Error(err))
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	render.JSON(w, r, jobSchedule)
}

// DeleteJobSchedule godoc
// @Summary delete JobSchedule
// @Description delete JobSchedule
// @Tags Scheduler
// @Produce json
// @Param id path string true "JobSchedule ID"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400	"Status Bad Request"
// @Failure 500	"Status Internal Server Error"
// @Router /engine/scheduler/jobs/{id} [delete]
func DeleteJobSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idJob, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Parsing JobSchedule id", zap.String("JobScheduleID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIParsingInteger, err)
		return
	}

	user, _ := GetUserFromContext(r)
	if !user.HasPermission(permissions.New(permissions.TypeScheduler, strconv.FormatInt(idJob, 10), permissions.ActionDelete)) {
		render.Error(w, r, render.ErrAPISecurityNoPermissions, errors.New("Missing permission"))
		return
	}

	err = scheduler.R().Delete(idJob)
	if err != nil {
		zap.L().Error("Delete DeleteJobSchedule", zap.String("ID", id), zap.Error(err))
		render.Error(w, r, render.ErrAPIDBDeleteFailed, err)
		return
	}

	scheduler.S().RemoveJobSchedule(idJob)

	render.OK(w, r)
}
