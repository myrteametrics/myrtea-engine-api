package modeler

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/coordinator"
	"github.com/myrteametrics/myrtea-sdk/v5/modeler"
	"github.com/spf13/viper"
)

func UpdateElasticTemplate(model modeler.Model) (string, error) {
	instanceName := viper.GetString("INSTANCE_NAME")
	var logicalIndexName string
	var err error

	switch model.ElasticsearchOptions.Rollmode.Type {
	case modeler.RollmodeCron:
		var cronIndex *coordinator.LogicalIndexCron
		cronIndex, err = coordinator.NewLogicalIndexCron(instanceName, model, true)
		if err == nil {
			logicalIndexName = cronIndex.Name
		}
	case modeler.RollmodeTimeBased:
		var timeBasedIndex *coordinator.LogicalIndexTimeBased
		timeBasedIndex, err = coordinator.NewLogicalIndexTimeBased(instanceName, model, true)
		if err == nil {
			logicalIndexName = timeBasedIndex.Name
		}
	}

	return logicalIndexName, err
}
