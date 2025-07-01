package coordinator

// import (
// 	"encoding/json"
// 	"io/ioutil"
// 	"net/http"

// 	"github.com/spf13/viper"
// 	"github.com/myrteametrics/myrtea-engine-api/v5/internals/coordinator"
// 	"github.com/myrteametrics/myrtea-engine-api/v5/internals/model"
// 	"github.com/myrteametrics/myrtea-sdk/v5/modeler"

// 	"go.uber.org/zap"
// )

// // ProcessNewModel godoc
// // @Title ProcessNewModel
// // @Description receive a model ID which has to be initialized
// // @tags Model
// // @Resource /coordinator
// // @Router /coordinator/models [post]
// // @Accept json
// // @Success 200 "OK"
// // @Failure 400 "Bad Request"
// // @Failure 500 "Internal Server Error"
// func ProcessNewModel(w http.ResponseWriter, r *http.Request) {
// 	b, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		zap.L().Error("ProcessNewModel.ReadAll()", zap.Error(err))
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	event := struct {
// 		ID int64 `json:"id"`
// 	}{}
// 	err = json.Unmarshal(b, &event)
// 	if err != nil {
// 		zap.L().Error("ProcessNewModel.Unmarshal()", zap.Error(err))
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	newModel, found, err := model.GetModelByID(event.ID)
// 	if err != nil {
// 		zap.L().Error("Couldn't get the specified model", zap.Error(err))
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	if !found {
// 		zap.L().Error("Couldn't get the specified model", zap.Error(err))
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	instanceName := viper.GetString("INSTANCE_NAME")
// 	if err = coordinator.GetInstance().Instances[instanceName].InitLogicalIndices([]modeler.Model{newModel}); err != nil {
// 		zap.L().Fatal("Intialisation of coordinator master", zap.String("status", "failed"), zap.Error(err))
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// }
