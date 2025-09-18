package evaluator

import (
	"context"
	"fmt"
	"github.com/PaesslerAG/gval"
	"github.com/myrteametrics/myrtea-sdk/v5/connector"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"strings"
)

func scanExpression(expressionToScan string, parameters map[string]interface{}) ([]string, interface{}, error) {
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
