package coordinator

import (
	"sync"

	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
	"github.com/spf13/viper"
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

	version := viper.GetInt("ELASTICSEARCH_VERSION")
	for _, model := range models {
		var err error
		var logicalIndex LogicalIndex
		switch model.ElasticsearchOptions.Rollmode {
		case "cron":
			switch version {
			case 6:
				logicalIndex, err = NewLogicalIndexCronV6(instance.Name, model)
			case 7:
				fallthrough
			case 8:
				logicalIndex, err = NewLogicalIndexCronV8(instance.Name, model)
			default:
				zap.L().Fatal("Unsupported Elasticsearch version", zap.Int("version", version))
			}

		case "timebased":
			switch version {
			case 6:
				logicalIndex, err = NewLogicalIndexTimeBasedV6(instance.Name, model)
			case 7:
				fallthrough
			case 8:
				logicalIndex, err = NewLogicalIndexTimeBasedV8(instance.Name, model)
			default:
				zap.L().Fatal("Unsupported Elasticsearch version", zap.Int("version", version))
			}
		}
		if err != nil {
			return err
		}
		instance.LogicalIndices[model.Name] = logicalIndex
	}

	_instance = instance

	return nil
}
