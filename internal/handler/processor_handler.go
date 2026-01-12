package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/ingester"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/processor"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/scheduler"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"github.com/myrteametrics/myrtea-sdk/v5/models"
	"go.uber.org/zap"
)

// ProcessorHandler is a basic struct allowing to set up a single aggregateIngester instance for all handlers
type ProcessorHandler struct {
	aggregateIngester *ingester.AggregateIngester
}

// NewProcessorHandler returns a pointer to an ProcessorHandler instance
func NewProcessorHandler() *ProcessorHandler {
	return &ProcessorHandler{
		aggregateIngester: ingester.NewAggregateIngester(),
	}
}

// PostObjects godoc
//
//	@Id				PostObjects
//
//	@Summary		Receive objects to be evaluated
//	@Description	Receive objects to be evaluated
//	@Tags			Service
//	@Produce		json
//	@Param			fact	query	string	true	"Fact object name"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/service/objects [post]
func PostObjects(w http.ResponseWriter, r *http.Request) {
	//TODO: What to do from groups ?
	//groups := GetUserGroupsFromContext(r)
	factObjectName := r.URL.Query().Get("fact")
	if factObjectName == "" {
		zap.L().Warn("fact object name missing")
		httputil.Error(w, r, httputil.ErrAPIMissingParam, errors.New(`parameter "fact" is missing`))
		return
	}

	var objects []models.Document
	err := json.NewDecoder(r.Body).Decode(&objects)
	if err != nil {
		zap.L().Warn("PostObjects.Unmarshal", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	err = processor.ReceiveObjects(factObjectName, objects)
	if err != nil {
		zap.L().Error("PostObjects.ReceiveObjects", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIProcessError, err)
		return
	}

	httputil.OK(w, r)
}

// PostAggregates godoc
//
//	@Id				PostAggregates
//
//	@Summary		Receive ingester to be evaluated
//	@Description	Receive ingester to be evaluated
//	@Tags			Service
//	@Consume		json
//	@Produce		json
//	@Param			query	body	[]scheduler.ExternalAggregate	true	"query (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		429	"Processing queue is full please retry later"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/service/aggregates [post]
func (handler *ProcessorHandler) PostAggregates(w http.ResponseWriter, r *http.Request) {
	var aggregates []scheduler.ExternalAggregate
	err := json.NewDecoder(r.Body).Decode(&aggregates)
	if err != nil {
		zap.L().Warn("PostAggregates.Unmarshal", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = handler.aggregateIngester.Ingest(aggregates)
	if err != nil {

		// Checks whether the queue is full, sends a 429 to prompt the sender to retry
		if err.Error() == "channel overload" {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		zap.L().Error("PostAggregates aggregateIngester.Ingest", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
