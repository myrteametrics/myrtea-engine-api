package handler

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"
)

// TODO: Add handler annotation on TestRuleHandler
func TestRule(w http.ResponseWriter, r *http.Request) {
	// TODO: Rule testing not implemented
	httputil.NotImplemented(w, r)

	// id := chi.URLParam(r, "id")
	// idRule, err := strconv.ParseInt(id, 10, 64)
	// if err != nil {
	// 	zap.L().Warn("Error on parsing rule id", zap.String("RuleID", id), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIParsingInteger, err)
	// 	return
	// }

	// sid := r.URL.Query().Get("situationid")
	// idSituation, err := strconv.ParseInt(sid, 10, 64)
	// if err != nil {
	// 	zap.L().Warn("Error on parsing situation id", zap.String("SitationID", id), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIParsingInteger, err)
	// 	return
	// }

	// iid := r.URL.Query().Get("instanceid")
	// idInstance, err := strconv.ParseInt(iid, 10, 64)
	// if err != nil {
	// 	zap.L().Warn("Error on parsing situation id", zap.String("InstanceID", id), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIParsingInteger, err)
	// 	return
	// }

	// record, err := situation.GetFromHistory(idSituation, time.Now(), idInstance, true)
	// if err != nil {
	// 	zap.L().Error("Get situation from history", zap.Int64("situationID", record.ID), zap.Time("ts", record.TS), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIProcessError, err)
	// 	return
	// }

	// var engineID = "debug"
	// if _, ok := evaluator.GetEngine(engineID); !ok {
	// 	err = evaluator.InitEngine(engineID)
	// 	if err != nil {
	// 		zap.L().Error("Init engine", zap.String("engineID", engineID), zap.Error(err))
	// 		render.Error(w, r, render.ErrAPIProcessError, err)
	// 		return
	// 	}
	// }

	// err = evaluator.UpdateEngine(engineID)
	// if err != nil {
	// 	zap.L().Error("Update engine", zap.String("engineID", engineID), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIProcessError, err)
	// 	return
	// }

	// localRuleEngine, err := evaluator.CloneEngine(engineID, true, false)
	// if err != nil {
	// 	zap.L().Error("Clone local engine", zap.String("engineID", engineID), zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIProcessError, err)
	// 	return
	// }

	// sKnowledge, err := evaluator.GetSituationKnowledge(evaluator.SituationToEvaluate{
	// 	ID:                 record.ID,
	// 	TS:                 record.TS,
	// 	TemplateInstanceID: record.TemplateInstanceID,
	// })
	// if err != nil {
	// 	zap.L().Error("Get situation knowledge", zap.Error(err))
	// 	render.Error(w, r, render.ErrAPIProcessError, err)
	// 	return
	// }

	// for key, value := range expression.GetDateKeywords(record.TS) {
	// 	sKnowledge[key] = value
	// }

	// localRuleEngine.Reset()
	// localRuleEngine.GetKnowledgeBase().SetFacts(sKnowledge)
	// localRuleEngine.ExecuteRules([]int64{idRule})
	// agenda := localRuleEngine.GetResults()

	// render.JSON(w, r, agenda)
}
