package coordinator

import (
	"context"
	"sync"

	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearch"
	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
	"go.uber.org/zap"
)

// Instance represents a functional Myrtea instance
type Instance struct {
	Initialized    bool
	Name           string
	Urls           []string
	Executor       *elasticsearch.EsExecutor
	LogicalIndices map[string]LogicalIndex
}

var (
	_instance *Instance
	_once     sync.Once
)

func GetInstance() *Instance {
	return _instance
}

func (i *Instance) LogicalIndex(modelName string) LogicalIndex {
	return i.LogicalIndices[modelName]
}

func InitInstance(instanceName string, urls []string, models map[int64]modeler.Model) error {
	zap.L().Info("Initialize coordinator instance", zap.String("instanceName", instanceName))

	instance := &Instance{
		Initialized:    false,
		Name:           instanceName,
		Urls:           urls,
		Executor:       nil,
		LogicalIndices: make(map[string]LogicalIndex),
	}

	for _, model := range models {
		executor, err := elasticsearch.NewEsExecutor(context.Background(), instance.Urls)
		if err != nil {
			zap.L().Error("elasticsearch.NewEsExecutor()", zap.Error(err))
			return err
		}

		switch model.ElasticsearchOptions.Rollmode {
		case "cron":
			logicalIndex, err := NewLogicalIndexCron(instance.Name, model, executor)
			if err != nil {
				zap.L().Error("logicalIndex.initialize()", zap.Error(err))
				return err
			}
			instance.LogicalIndices[model.Name] = logicalIndex

		case "timebased":
			logicalIndex, err := NewLogicalIndexTimeBased(instance.Name, model, executor)
			if err != nil {
				zap.L().Error("NewLogicalIndexTimeBased()", zap.Error(err))
				return err
			}
			instance.LogicalIndices[model.Name] = logicalIndex
		}
	}

	_instance = instance

	return nil
}
