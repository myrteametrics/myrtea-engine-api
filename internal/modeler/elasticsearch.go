package modeler

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/coordinator"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"github.com/myrteametrics/myrtea-sdk/v5/modeler"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func UpdateElasticTemplate(id string, w http.ResponseWriter, r *http.Request) (string, error) {
	idModel, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Error("Error on parsing model id", zap.String("modelID", id))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return "", err
	}

	model, found, err := R().Get(idModel)
	if err != nil {
		zap.L().Error("Error while fetching model", zap.String("modelid", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return "", err
	}
	if !found {
		zap.L().Error("Model does not exists", zap.String("modelid", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return "", err
	}

	instanceName := viper.GetString("INSTANCE_NAME")
	var logicalIndexName string

	switch model.ElasticsearchOptions.Rollmode.Type {
	case modeler.RollmodeCron:
		var cronIndex *coordinator.LogicalIndexCron
		cronIndex, _, err = coordinator.NewLogicalIndexCronTemplate(instanceName, model)
		if err == nil {
			logicalIndexName = cronIndex.Name
		}
	case modeler.RollmodeTimeBased:
		var timeBasedIndex *coordinator.LogicalIndexTimeBased
		timeBasedIndex, err = coordinator.NewLogicalIndexTimeBased(instanceName, model)
		if err == nil {
			logicalIndexName = timeBasedIndex.Name
		}
	}

	return logicalIndexName, err
}
