package coordinator

import (
	"sync"

	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
	"go.uber.org/zap"
)

// Instance represents a functional Myrtea instance
type Instance struct {
	Initialized    bool
	Name           string
	Urls           []string
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

func InitInstance(instanceName string, models map[int64]modeler.Model) error {
	zap.L().Info("Initialize coordinator instance", zap.String("instanceName", instanceName))

	instance := &Instance{
		Initialized:    false,
		Name:           instanceName,
		LogicalIndices: make(map[string]LogicalIndex),
	}

	for _, model := range models {
		switch model.ElasticsearchOptions.Rollmode {
		case "cron":
			logicalIndex, err := NewLogicalIndexCron(instance.Name, model)
			if err != nil {
				zap.L().Error("logicalIndex.initialize()", zap.Error(err))
				return err
			}
			instance.LogicalIndices[model.Name] = logicalIndex

		case "timebased":
			logicalIndex, err := NewLogicalIndexTimeBasedV6(instance.Name, model)
			if err != nil {
				zap.L().Error("NewLogicalIndexTimeBasedV6()", zap.Error(err))
				return err
			}
			instance.LogicalIndices[model.Name] = logicalIndex
		}
	}

	_instance = instance

	return nil
}
