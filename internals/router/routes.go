package router

import (
	"net/http"

	"github.com/go-chi/chi"
	_ "github.com/myrteametrics/myrtea-engine-api/v4/docs" // docs is generated by Swag CLI, you have to import it.
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers"
)

func adminRouter() http.Handler {
	r := chi.NewRouter()

	// security
	r.Get("/security/users", handlers.GetUsers)
	r.Get("/security/users/{id}", handlers.GetUser)
	r.Post("/security/users/validate", handlers.ValidateUser)
	r.Post("/security/users", handlers.PostUser)
	r.Put("/security/users/{id}", handlers.PutUser)
	r.Delete("/security/users/{id}", handlers.DeleteUser)

	r.Get("/security/groups", handlers.GetGroups)
	r.Get("/security/groups/{id}", handlers.GetGroup)
	r.Post("/security/groups/validate", handlers.ValidateGroup)
	r.Post("/security/groups", handlers.PostGroup)
	r.Put("/security/groups/{id}", handlers.PutGroup)
	r.Delete("/security/groups/{id}", handlers.DeleteGroup)

	r.Get("/security/groups/{groupid}/users", handlers.GetUsersOfGroup)
	r.Put("/security/groups/{groupid}/users/{userid}", handlers.PutMembership)
	r.Delete("/security/groups/{groupid}/users/{userid}", handlers.DeleteMembership)

	r.Get("/engine/issues_all", handlers.GetIssues)

	return r
}

func engineRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/security/myself", handlers.GetUserSelf)

	r.Get("/models", handlers.GetModels)
	r.Get("/models/{id}", handlers.GetModel)
	r.Post("/models", handlers.PostModel)
	r.Post("/models/validate", handlers.ValidateModel)
	r.Put("/models/{id}", handlers.PutModel)
	r.Delete("/models/{id}", handlers.DeleteModel)

	r.Get("/crons", handlers.GetAllCron)
	r.Post("/crons/start", handlers.StartAllCron)
	r.Post("/crons/stop", handlers.StopAllCron)
	r.Get("/cron", handlers.GetCron)
	r.Post("/cron/{instance}/{logicalIndex}/start", handlers.StartCron)
	r.Post("/cron/{instance}/{logicalIndex}/stop", handlers.StopCron)

	r.Get("/facts", handlers.GetFacts)
	r.Get("/facts/{id}", handlers.GetFact)
	r.Post("/facts/validate", handlers.ValidateFact)
	r.Post("/facts", handlers.PostFact)
	r.Put("/facts/{id}", handlers.PutFact)
	r.Delete("/facts/{id}", handlers.DeleteFact)
	r.Get("/facts/{id}/execute", handlers.ExecuteFact)       // ?time=2019-05-10T12:00:00.000+02:00 debug=<boolean>
	r.Post("/facts/execute", handlers.ExecuteFactFromSource) // ?time=2019-05-10T12:00:00.000 debug=<boolean>
	r.Get("/facts/{id}/hits", handlers.GetFactHits)          // ?time=2019-05-10T12:00:00.000 debug=<boolean>
	r.Get("/facts/{id}/es", handlers.FactToESQuery)

	r.Get("/situations", handlers.GetSituations)
	r.Get("/situations/{id}", handlers.GetSituation)
	r.Post("/situations/validate", handlers.ValidateSituation)
	r.Post("/situations", handlers.PostSituation)
	r.Put("/situations/{id}", handlers.PutSituation)
	r.Delete("/situations/{id}", handlers.DeleteSituation)
	r.Get("/situations/{id}/facts", handlers.GetSituationFacts)
	r.Get("/situations/{id}/rules", handlers.GetSituationRules)
	r.Put("/situations/{id}/rules", handlers.SetSituationRules)
	r.Get("/situations/{id}/evaluation", handlers.GetSituationEvaluation)
	r.Get("/situations/{id}/instances", handlers.GetSituationTemplateInstances)
	r.Post("/situations/{id}/instances", handlers.PostSituationTemplateInstance)
	r.Put("/situations/{id}/instances/{instanceid}", handlers.PutSituationTemplateInstance)
	r.Put("/situations/{id}/instances", handlers.PutSituationTemplateInstances)
	r.Delete("/situations/{id}/instances/{instanceid}", handlers.DeleteSituationTemplateInstance)

	r.Get("/rules", handlers.GetRules)
	r.Get("/rules/{id}", handlers.GetRule)
	r.Get("/rules/{id}/versions/{versionid}", handlers.GetRuleByVersion)
	r.Post("/rules/validate", handlers.ValidateRule)
	r.Post("/rules", handlers.PostRule)
	r.Put("/rules/{id}", handlers.PutRule)
	r.Delete("/rules/{id}", handlers.DeleteRule)
	r.Get("/rules/{id}/situations", handlers.GetRuleSituations)
	r.Post("/rules/{id}/situations", handlers.PostRuleSituations)

	r.Get("/issues", handlers.GetIssuesByStatesByPage)
	r.Get("/issues/{id}", handlers.GetIssue)
	r.Get("/issues/{id}/facts_history", handlers.GetIssueFactsHistory)
	r.Post("/issues", handlers.PostIssue)
	r.Get("/issues/{id}/recommendation", handlers.GetIssueFeedbackTree)
	r.Post("/issues/{id}/feedback", handlers.PostIssueCloseWithFeedback)
	r.Post("/issues/{id}/draft", handlers.PostIssueDraft)
	r.Post("/issues/{id}/close", handlers.PostIssueCloseWithoutFeedback)
	r.Post("/issues/{id}/detection/feedback", handlers.PostIssueDetectionFeedback)

	r.Post("/scheduler/start", handlers.StartScheduler)
	r.Post("/scheduler/trigger", handlers.TriggerJobSchedule)
	r.Get("/scheduler/jobs", handlers.GetJobSchedules)
	r.Get("/scheduler/jobs/{id}", handlers.GetJobSchedule)
	r.Post("/scheduler/jobs/validate", handlers.ValidateJobSchedule)
	r.Post("/scheduler/jobs", handlers.PostJobSchedule)
	r.Put("/scheduler/jobs/{id}", handlers.PutJobSchedule)
	r.Delete("/scheduler/jobs/{id}", handlers.DeleteJobSchedule)

	r.Put("/notifications/{id}/read", handlers.UpdateRead)

	r.HandleFunc("/notifications/ws", handlers.NotificationsWSRegister)
	r.HandleFunc("/notifications/sse", handlers.NotificationsSSERegister)
	r.Get("/notifications", handlers.GetNotifications)
	r.Post("/notifications/trigger", handlers.TriggerNotification)

	r.Get("/rootcauses", handlers.GetRootCauses)
	r.Get("/rootcauses/{id}", handlers.GetRootCause)
	r.Post("/rootcauses/validate", handlers.ValidateRootCause)
	r.Post("/rootcauses", handlers.PostRootCause)
	r.Put("/rootcauses/{id}", handlers.PutRootCause)
	r.Delete("/rootcauses/{id}", handlers.DeleteRootCause)

	r.Get("/actions", handlers.GetActions)
	r.Get("/actions/{id}", handlers.GetAction)
	r.Post("/actions/validate", handlers.ValidateAction)
	r.Post("/actions", handlers.PostAction)
	r.Put("/actions/{id}", handlers.PutAction)
	r.Delete("/actions/{id}", handlers.DeleteAction)

	r.Post("/search", handlers.Search)

	r.Get("/calendars", handlers.GetCalendars)
	r.Get("/calendars/{id}", handlers.GetCalendar)
	r.Get("/calendars/{id}/contains", handlers.IsInCalendarPeriod) // ?time=2019-05-10T12:00:00.000
	r.Get("/calendars/resolved/{id}", handlers.GetResolvedCalendar)
	r.Post("/calendars", handlers.PostCalendar)
	r.Put("/calendars/{id}", handlers.PutCalendar)
	r.Delete("/calendars/{id}", handlers.DeleteCalendar)

	r.Get("/connector/{id}/executions/last", handlers.GetlastConnectorExecutionDateTime)

	return r
}

func serviceRouter() http.Handler {
	r := chi.NewRouter()

	r.Post("/objects", handlers.PostObjects)
	r.Post("/aggregates", handlers.PostAggregates)

	return r
}
