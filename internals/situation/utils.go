package situation

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/variablesconfig"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
)

func shouldParseGlobalVariables(translateOpt ...bool) bool {
	if len(translateOpt) > 0 {
		return translateOpt[0]
	}
	return true
}

func UpdateParametersWithConfig(params map[string]string) {
	listKeyValue, err := variablesconfig.R().GetAllAsMap()

	if err != nil {
		zap.L().Error("Can't get list of variable config in the database", zap.Error(err))
		return
	}

	ReplaceKeysWithValues(params, listKeyValue)
}

func ReplaceKeysWithValues(m map[string]string, variables map[string]interface{}) {

	for key, value := range m {

		translate, err := expression.Process(expression.LangEval, value, variables)

		if err != nil {
			zap.L().Error("Error: Unrecognized variable Global in this Parameter", zap.Any("key", key), zap.Any("value", value), zap.Error(err))
		} else {

			m[key] = fmt.Sprint(translate)
		}
	}

}
