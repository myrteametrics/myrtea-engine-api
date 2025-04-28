package situation

import (
	"go.uber.org/zap"

	"github.com/myrteametrics/myrtea-sdk/v5/expression"
)

func shouldParseForEvaluation(translateOpt ...bool) bool {
	if len(translateOpt) > 0 {
		return translateOpt[0]
	}
	return true
}

func EvalParameters(m map[string]interface{}) {

	for key, value := range m {

		translate, err := expression.Process(expression.LangEval, value.(string), map[string]interface{}{})

		if err != nil {
			zap.L().Error("Error: Unrecognized variable Global in this Parameter", zap.Any("key", key), zap.Any("value", value), zap.Error(err))
		} else {

			m[key] = translate
		}
	}

}
