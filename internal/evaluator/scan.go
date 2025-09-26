package evaluator

import (
	"context"
	"fmt"
	"github.com/PaesslerAG/gval"
	"github.com/myrteametrics/myrtea-sdk/v5/connector"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"strings"
	"time"
)

func scanExpressionInternal(expressionToScan string, parameters map[string]interface{}) ([]string, interface{}, error) {
	var variables []string

	scannedExpression, err := gval.Evaluate(expressionToScan,
		parameters,
		gval.VariableSelector(func(path gval.Evaluables) gval.Evaluable {
			return func(c context.Context, v interface{}) (interface{}, error) {
				keys, err := path.EvalStrings(c, v)
				if err != nil {
					return nil, err
				}
				variables = append(variables, strings.Join(keys, "."))

				if value, ok := connector.LookupNestedMap(keys, v); ok {
					return value, nil
				}

				return fmt.Sprintf("{{ %s }}", strings.Join(keys, " ")), nil
			}
		}),
		expression.LangEval,
	)

	return variables, scannedExpression, err
}

func mergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy the first map into the result
	for k, v := range map1 {
		result[k] = v
	}

	// Ajouter ou remplacer avec les valeurs de la deuxième map
	for k, v := range map2 {
		result[k] = v
	}

	return result
}

// Try to inject value on variables when expression fails to be evaluated
// Doesn't support functions with multiple variable parameters
func testParameters(expressionToScan string, parameters map[string]interface{}, variable string) (map[string]interface{}, error) {
	simulatedParameters := make(map[string]interface{})
	simulatedParameters[variable] = 0
	mergedParameters := mergeMaps(parameters, simulatedParameters)
	variables, _, err := scanExpressionInternal(expressionToScan, mergedParameters)
	if err == nil || variables[len(variables)-1] != variable {
		return mergedParameters, nil
	}
	simulatedParameters[variable] = 0.0
	mergedParameters = mergeMaps(parameters, simulatedParameters)
	variables, _, err = scanExpressionInternal(expressionToScan, mergedParameters)
	if err == nil || variables[len(variables)-1] != variable {
		return mergedParameters, nil
	}
	simulatedParameters[variable] = []float64{0}
	mergedParameters = mergeMaps(parameters, simulatedParameters)
	variables, _, err = scanExpressionInternal(expressionToScan, mergedParameters)
	if err == nil || variables[len(variables)-1] != variable {
		return mergedParameters, nil
	}
	simulatedParameters[variable] = []string{"default"}
	mergedParameters = mergeMaps(parameters, simulatedParameters)
	variables, _, err = scanExpressionInternal(expressionToScan, mergedParameters)
	if err == nil || variables[len(variables)-1] != variable {
		return mergedParameters, nil
	}
	simulatedParameters[variable] = map[string]interface{}{"default": "default"}
	mergedParameters = mergeMaps(parameters, simulatedParameters)
	variables, _, err = scanExpressionInternal(expressionToScan, mergedParameters)
	if err == nil || variables[len(variables)-1] != variable {
		return mergedParameters, nil
	}
	simulatedParameters[variable] = time.Now().Format("2006-01-02 15:04:05-07:00")

	mergedParameters = mergeMaps(parameters, simulatedParameters)
	variables, _, err = scanExpressionInternal(expressionToScan, mergedParameters)
	if err == nil || variables[len(variables)-1] != variable {
		return mergedParameters, nil
	}

	return nil, fmt.Errorf("unable to find a valid value for the variable %s", variable)
}

func ScanExpression(expressionToScan string, parameters map[string]interface{}, injectValuesOnFail bool) ([]string, interface{}, error) {
	for {
		variables, scannedExpression, err := scanExpressionInternal(expressionToScan, parameters)
		if !injectValuesOnFail {
			return variables, scannedExpression, err
		}
		if err != nil && len(variables) != 0 {
			mergedParameters, err2 := testParameters(expressionToScan, parameters, variables[len(variables)-1])
			if err2 != nil {
				return nil, nil, err
			}
			parameters = mergedParameters
		} else {
			return variables, scannedExpression, err
		}
	}
}
