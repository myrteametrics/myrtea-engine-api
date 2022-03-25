package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/evaluator"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/processor"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/scheduler"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tasker"
	"github.com/myrteametrics/myrtea-sdk/v4/models"
	"go.uber.org/zap"
)

// PostObjects godoc
// @Summary Receive objects to be evaluated
// @Description Receive objects to be evaluated
// @Tags Service
// @Produce json
// @Param fact query string true "Fact object name"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 500 "internal server error"
// @Router /service/objects [post]
func PostObjects(w http.ResponseWriter, r *http.Request) {
	//TODO: What to do from groups ?
	//groups := GetUserGroupsFromContext(r)
	factObjectName := r.URL.Query().Get("fact")
	if factObjectName == "" {
		zap.L().Warn("fact object name missing")
		render.Error(w, r, render.ErrAPIMissingParam, errors.New(`Parameter "fact" is missing`))
		return
	}

	var objects []models.Document
	err := json.NewDecoder(r.Body).Decode(&objects)
	if err != nil {
		zap.L().Warn("PostObjects.Unmarshal", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	err = processor.ReceiveObjects(factObjectName, objects)
	if err != nil {
		zap.L().Error("PostObjects.ReceiveObjects", zap.Error(err))
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	render.OK(w, r)
}

// ExternalAggregate contains all information to store a new aggregat in postgresql
type ExternalAggregate struct {
	FactID              int64       `json:"factId"`
	SituationID         int64       `json:"situationId"`
	SituationInstanceID int64       `json:"situationInstanceId"`
	Time                time.Time   `json:"time"`
	Value               reader.Item `json:"value"`
}

// PostAggregates godoc
// @Summary Receive aggregates to be evaluated
// @Description Receive aggregates to be evaluated
// @Tags Service
// @Consume json
// @Produce json
// @Param query body []ExternalAggregate true "query (json)"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 500 "internal server error"
// @Router /service/aggregates [post]
func PostAggregates(w http.ResponseWriter, r *http.Request) {
	var aggregates []ExternalAggregate
	err := json.NewDecoder(r.Body).Decode(&aggregates)
	if err != nil {
		zap.L().Warn("PostAggregates.Unmarshal", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = ReceiveAggregates(aggregates)
	if err != nil {
		zap.L().Error("ReceiveAggregates", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ReceiveAggregates process a slice of ExternalAggregates and trigger all standard fact-situation-rule process
func ReceiveAggregates(aggregates []ExternalAggregate) error {

	situationsToUpdate := make(map[string]situation.HistoryRecord, 0)

	for _, agg := range aggregates {

		t := agg.Time.UTC().Truncate(time.Second)

		f, found, err := fact.R().Get(agg.FactID)
		if err != nil {
			return err
		}
		if !found {
			return errors.New("not found")
		}

		s, found, err := situation.R().Get(agg.SituationID, groups.GetTokenAllGroups())
		if err != nil {
			return err
		}
		if !found {
			return errors.New("not found")
		}

		found = false
		for _, factID := range s.Facts {
			if f.ID == factID {
				found = true
			}
		}
		if !found {
			zap.L().Error("Fact doesn't exist in situation")
			continue
		}

		si, found, err := situation.R().GetTemplateInstance(agg.SituationInstanceID)
		if err != nil {
			return err
		}
		if !found {
			return errors.New("not found")
		}

		if s.ID != si.SituationID {
			zap.L().Error("invalid s.ID != si.SituationID")
			continue
		}

		factSituationsHistory, err := scheduler.GetFactSituations(f, t)
		if err != nil {
			zap.L().Warn("getFactSituations", zap.Int64("factID", f.ID), zap.Error(err))
			continue
		}
		if len(factSituationsHistory) == 0 {
			zap.L().Warn("fact has no situation history", zap.Int64("factID", f.ID))
			continue
		}

		if f.IsTemplate {
			for _, sh := range factSituationsHistory {
				if sh.ID != s.ID || sh.TemplateInstanceID != si.ID {
					continue
				}

				err := fact.PersistFactResult(f.ID, t, s.ID, si.ID, &agg.Value, true)
				if err != nil {
					zap.L().Error("Cannot persist fact instance", zap.Error(err))
					return err
				}

				key := fmt.Sprintf("%d-%d", sh.ID, sh.TemplateInstanceID)
				if _, ok := situationsToUpdate[key]; !ok {
					situationsToUpdate[key] = situation.HistoryRecord{
						ID:                 sh.ID,
						TS:                 t,
						FactsIDS:           map[int64]*time.Time{f.ID: &t},
						Parameters:         sh.Parameters,
						TemplateInstanceID: sh.TemplateInstanceID,
					}
				} else {
					situationsToUpdate[key].FactsIDS[f.ID] = &t
				}
			}
		} else {
			err := fact.PersistFactResult(f.ID, t, s.ID, si.ID, &agg.Value, true)
			if err != nil {
				zap.L().Error("Cannot persist fact instance", zap.Error(err))
				return err
			}

			for _, sh := range factSituationsHistory {
				if sh.ID != s.ID || sh.TemplateInstanceID != si.ID {
					continue
				}
				key := fmt.Sprintf("%d-%d", sh.ID, sh.TemplateInstanceID)
				if _, ok := situationsToUpdate[key]; !ok {
					situationsToUpdate[key] = situation.HistoryRecord{
						ID:                 sh.ID,
						TS:                 t,
						FactsIDS:           map[int64]*time.Time{f.ID: &t},
						Parameters:         sh.Parameters,
						TemplateInstanceID: sh.TemplateInstanceID,
					}
				} else {
					situationsToUpdate[key].FactsIDS[f.ID] = &t
				}
			}
		}
	}

	zap.L().Debug("situationsToUpdate", zap.Any("situationsToUpdate", situationsToUpdate))

	situationsToEvaluate, err := scheduler.UpdateSituations(situationsToUpdate)
	if err != nil {
		zap.L().Error("Cannot update situations", zap.Error(err))
		return err
	}

	situationEvaluations, err := evaluator.EvaluateSituations(situationsToEvaluate, "external-aggs")
	if err == nil {
		taskBatchs := make([]tasker.TaskBatch, 0)
		for _, situationEvaluation := range situationEvaluations {
			taskBatchs = append(taskBatchs, tasker.TaskBatch{
				Context: map[string]interface{}{
					"situationID":        situationEvaluation.ID,
					"ts":                 situationEvaluation.TS,
					"templateInstanceID": situationEvaluation.TemplateInstanceID,
				},
				Agenda: situationEvaluation.Agenda,
			})
		}

		tasker.T().BatchReceiver <- taskBatchs
	}

	return nil
}
