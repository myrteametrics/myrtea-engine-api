package coordinator

import (
	"context"
	"fmt"

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
	LogicalIndices map[string]*LogicalIndex
}

func (instance *Instance) initialize(models map[int64]modeler.Model) error {
	err := instance.InitLogicalIndices(models)
	if err != nil {
		zap.L().Error("instance.InitLogicalIndices()", zap.Error(err))
		return err
	}

	instance.Initialized = true
	return nil
}

func (instance *Instance) initElasticClient(urls []string) error {
	// TODO: Multiple elasticsearch URL Support
	executor, err := elasticsearch.NewEsExecutor(context.Background(), urls)
	if err != nil {
		zap.L().Error("elasticsearch.NewEsExecutor()", zap.Error(err))
		return err
	}

	instance.Executor = executor
	return nil
}

// InitLogicalIndices initialize an ensemble of logical indices (each based on a specific elasticsearch model)
func (instance *Instance) InitLogicalIndices(models map[int64]modeler.Model) error {
	for _, m := range models {

		zap.L().Info("Initialize model indices", zap.String("model", m.Name), zap.Any("options", m.ElasticsearchOptions))

		executor, err := elasticsearch.NewEsExecutor(context.Background(), instance.Urls)
		if err != nil {
			zap.L().Error("elasticsearch.NewEsExecutor()", zap.Error(err))
			return err
		}

		logicalIndex := &LogicalIndex{
			Initialized: false,
			Name:        fmt.Sprintf("%s-%s", instance.Name, m.Name),
			Cron:        nil,
			Executor:    executor,
			Model:       m,
		}
		err = logicalIndex.initialize()
		if err != nil {
			zap.L().Error("logicalIndex.initialize()", zap.Error(err))
			return err
		}

		instance.LogicalIndices[m.Name] = logicalIndex
	}
	return nil
}
