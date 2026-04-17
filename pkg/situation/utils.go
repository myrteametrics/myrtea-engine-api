package situation

import (
	"time"

	"go.uber.org/zap"

	"github.com/myrteametrics/myrtea-sdk/v5/expression"
)

// shouldParseForEvaluation checks if the expression should be parsed for evaluation
func shouldParseForEvaluation(translateOpt ...bool) bool {
	if len(translateOpt) > 0 {
		return translateOpt[0]
	}
	return true
}

// evalParameters evaluates the parameters of a situation instance using gval
func evalParameters(m map[string]interface{}, ts time.Time) {

	variable := make(map[string]interface{})
	for key, value := range expression.GetDateKeywords(ts) {
		variable[key] = value
	}
	for key, value := range m {
		translate, err := expression.Process(expression.LangEval, value.(string), variable)

		if err != nil {
			zap.L().Error("Error: Unrecognized global variable in this parameter", zap.Any("key", key),
				zap.Any("value", value), zap.Error(err))
		} else {
			m[key] = translate
		}
	}

}
