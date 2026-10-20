package evaluator

import (
	"sort"
	"strings"

	"github.com/myrteametrics/myrtea-sdk/v5/connector"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	"go.uber.org/zap"
)

// FactVariable is a variable read by a fact condition expression and resolved by none of the known parameters
type FactVariable struct {
	Name string `json:"name"`
	// Optional is true when the variable is only read by the value of an optional operator,
	// whose condition is dropped instead of failing when the variable cannot be resolved
	Optional bool `json:"optional"`
}

// ScanFactConditionVariables lists the variables read by the expressions of a fact condition tree
// (value, value2 and timezone of every leaf) that parameters does not resolve, sorted by name.
// It evaluates the expressions the way the fact contextualization does, so parameters must already hold
// every variable available at execution time (date keywords, global variables, situation parameters...).
func ScanFactConditionVariables(condition engine.ConditionFragment, parameters map[string]interface{}) []FactVariable {
	optionalByName := make(map[string]bool)
	scanCondition(condition, parameters, optionalByName)

	variables := make([]FactVariable, 0, len(optionalByName))
	for name, optional := range optionalByName {
		variables = append(variables, FactVariable{Name: name, Optional: optional})
	}
	sort.Slice(variables, func(i, j int) bool { return variables[i].Name < variables[j].Name })
	return variables
}

func scanCondition(condition engine.ConditionFragment, parameters map[string]interface{}, optionalByName map[string]bool) {
	switch c := condition.(type) {
	case *engine.BooleanFragment:
		for _, fragment := range c.Fragments {
			scanCondition(fragment, parameters, optionalByName)
		}
	case *engine.LeafConditionFragment:
		optional := c.Operator == engine.OptionalFor || c.Operator == engine.OptionalRegexp || c.Operator == engine.OptionalWildcard
		scanValue(c.Value, parameters, optional, optionalByName)
		scanValue(c.Value2, parameters, false, optionalByName)
		if c.TimeZone != "" {
			scanValue(c.TimeZone, parameters, false, optionalByName)
		}
	}
}

// scanValue records the missing variables of an expression; a variable read by a mandatory expression stays mandatory
func scanValue(value interface{}, parameters map[string]interface{}, optional bool, optionalByName map[string]bool) {
	exp, ok := value.(string)
	if !ok || exp == "" {
		return
	}

	variables, _, err := ScanExpression(exp, parameters, true)
	if err != nil {
		// No simulated value lets the evaluation go through: keep the variables read before it failed
		zap.L().Debug("Expression scan failed", zap.String("exp", exp), zap.Error(err))
		variables, _, _ = ScanExpression(exp, parameters, false)
	}

	for _, name := range variables {
		if _, found := connector.LookupNestedMap(strings.Split(name, "."), parameters); found {
			continue
		}
		if alreadyOptional, seen := optionalByName[name]; seen {
			optionalByName[name] = alreadyOptional && optional
		} else {
			optionalByName[name] = optional
		}
	}
}
