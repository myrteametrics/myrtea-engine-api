package coordinator

import (
	"sync"

	"github.com/myrteametrics/myrtea-sdk/v5/modeler"
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
		var err error
		var logicalIndex LogicalIndex
		switch model.ElasticsearchOptions.Rollmode.Type {
		case modeler.RollmodeCron:
			logicalIndex, err = NewLogicalIndexCron(instance.Name, model)
		case modeler.RollmodeTimeBased:
			logicalIndex, err = NewLogicalIndexTimeBased(instance.Name, model)
		}
		if err != nil {
			return err
		}
		instance.LogicalIndices[model.Name] = logicalIndex
	}

	_instance = instance

	return nil
}
