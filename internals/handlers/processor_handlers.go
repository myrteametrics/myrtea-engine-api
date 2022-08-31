package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/evaluator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/processor"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/scheduler"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tasker"
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
	var aggregates []scheduler.ExternalAggregate
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
func ReceiveAggregates(aggregates []scheduler.ExternalAggregate) error {
	localRuleEngine, err := evaluator.BuildLocalRuleEngine("external-aggs")
	if err != nil {
		zap.L().Error("BuildLocalRuleEngine", zap.Error(err))
		return err
	}

	situationsToUpdate, err := scheduler.ReceiveAndPersistFacts(aggregates)
	if err != nil {
		zap.L().Error("ReceiveAndPersistFacts", zap.Error(err))
		return err
	}

	taskBatchs, err := scheduler.CalculateAndPersistSituations(localRuleEngine, situationsToUpdate)
	if err != nil {
		zap.L().Error("CalculateAndPersistSituations", zap.Error(err))
		return err
	}

	tasker.T().BatchReceiver <- taskBatchs

	return nil
}
