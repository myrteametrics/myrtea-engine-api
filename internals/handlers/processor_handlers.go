package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/processor"
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
