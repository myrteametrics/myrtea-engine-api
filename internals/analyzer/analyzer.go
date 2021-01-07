package analyzer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Item is used for ML external component
type Item struct {
	Features map[string]interface{} `json:"features"`
	Targets  map[string]interface{} `json:"targets"`
}

// NewItem build a new Item
func NewItem() Item {
	return Item{
		Features: make(map[string]interface{}, 0),
		Targets:  make(map[string]interface{}, 0),
	}
}

// Prediction is used as a ML prediction result
type Prediction struct {
	Classes map[string]float64 `json:"classes"`
}

// Predict call an external ML model to make prediction on an objet features
func Predict(object map[string]interface{}) (Prediction, error) {

	apiServiceURL := viper.GetString("EXPLAINER_API_SERVICE_URL")
	dataset := viper.GetString("EXPLAINER_API_DATASET_NAME")
	mlModel := viper.GetString("EXPLAINER_API_ML_MODEL_NAME")
	fields := viper.GetStringSlice("EXPLAINER_API_DATASET_FIELDS")

	item := NewItem()
	for _, field := range fields {
		if val, ok := object[field]; ok {
			item.Features[field] = val
		}
	}

	zap.L().Info("item", zap.Any("item", item))

	b, err := json.Marshal(item)
	if err != nil {
		zap.L().Error("marshal", zap.Error(err))
		return Prediction{}, err
	}

	zap.L().Debug("body item", zap.ByteString("b", b))

	url := fmt.Sprintf("%s/api/v1/datasets/%s/classification/%s/predict", apiServiceURL, dataset, mlModel)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		zap.L().Error("post to python api", zap.Error(err))
		return Prediction{}, err
	}

	var prediction Prediction
	err = json.NewDecoder(resp.Body).Decode(&prediction)
	if err != nil {
		zap.L().Error("decode", zap.Error(err), zap.Any("body", resp.Body))
		return Prediction{}, err
	}

	zap.L().Info("prediction", zap.Any("prediction", prediction))

	return prediction, nil
}
