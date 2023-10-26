package utils

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/variablesconfig"
	sdkvariablesconfig "github.com/myrteametrics/myrtea-sdk/v4/variablesconfig"
	"go.uber.org/zap"
)

func RemoveDuplicates(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func UpdateParametersWithConfig(params *map[string]string) {
	listKeyValue, err := variablesconfig.R().GetAllAsMap()

	if err != nil {
		zap.L().Error("Can't get list of variable config in the database", zap.Error(err))
		return
	}

	sdkvariablesconfig.ReplaceKeysWithValues(params, listKeyValue)
}
