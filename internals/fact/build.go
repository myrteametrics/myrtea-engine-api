package fact

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact/lexer"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	parsec "github.com/prataprc/goparsec"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// BuildFactsFromFile This functions creates the facts from the provided file
func BuildFactsFromFile(path string, file string) (map[string]*engine.Fact, []error) {
	conf := viper.New()
	conf.SetConfigType("yaml")
	conf.AddConfigPath(path)
	conf.SetConfigName(file)

	if err := conf.ReadInConfig(); err != nil {
		zap.L().Error(fmt.Sprintf("initializeConfig.ReadInConfig: %s", err))
		return nil, []error{err}
	}

	facts := make(map[string]*engine.Fact, 0)
	var errs []error

	factsRaw := conf.GetStringMap("facts")
	for rawFactKey, rawFactValue := range factsRaw {
		zap.L().Info("Reflecting fact", zap.String("fact", rawFactKey))
		if reflect.ValueOf(rawFactValue).Kind() != reflect.Map {
			zap.L().Info("ERROR: Reflect fact data is not a map", zap.String("fact", rawFactKey))
			continue
		}

		fact, err := ParseFact(rawFactKey, rawFactValue.(map[string]interface{}))
		if err != nil {
			errs = append(errs, errors.New(rawFactKey+": "+err.Error()))
		}

		if fact != nil {
			facts[rawFactKey] = fact
		}
	}
	return facts, errs
}

// ParseFact parse a fact configuration map and return a new instance of fact
func ParseFact(factName string, factData map[string]interface{}) (*engine.Fact, error) {
	var source string
	var model string
	var intentOperator string
	var dimensions []*engine.DimensionFragment
	var restitutions []engine.Restitution
	var comment string
	var term string
	var condition engine.ConditionFragment

	for factFieldKey, factFieldValue := range factData {

		switch factFieldKey {
		case "source":
			if reflect.ValueOf(factFieldValue).Kind() == reflect.String {
				source = factFieldValue.(string)
			} else {
				return nil, errors.New("source is not a string")
			}

		case "model":
			if reflect.ValueOf(factFieldValue).Kind() == reflect.String {
				model = factFieldValue.(string)
			} else {
				return nil, errors.New("model is not a string")
			}

		case "operator":
			if reflect.ValueOf(factFieldValue).Kind() == reflect.String {
				intentOperator = factFieldValue.(string)
			} else {
				return nil, errors.New("intent operator is not a string")
			}

		case "term":
			if reflect.ValueOf(factFieldValue).Kind() == reflect.String {
				term = factFieldValue.(string)
			} else {
				return nil, errors.New("term is not a string")
			}

		case "dimensions":
			if reflect.ValueOf(factFieldValue).Kind() == reflect.Slice {
				factDimensions := factFieldValue.([]interface{})
				for _, iDimension := range factDimensions {
					if reflect.ValueOf(iDimension).Kind() == reflect.String {
						dim := iDimension.(string)
						dimParts := strings.Split(dim, " ")
						operator := strings.ToLower(dimParts[0])
						dimension, err := engine.GetDimensionFragment(operator)
						if err != nil {
							return nil, err
						}
						if operator == engine.By.String() && len(dimParts) == 2 {
							dimension.Term = dimParts[1]
							dimensions = append(dimensions, dimension)
						} else if operator == engine.DateHistogram.String() && len(dimParts) == 3 {
							dimension.Term = dimParts[1]
							dimension.DateInterval = dimParts[2]
							dimensions = append(dimensions, dimension)
						} else {
							return nil, errors.New("dimension cannot be splitted between operator and term")
						}
					}
				}
			} else {
				return nil, errors.New("dimension is not a slice")
			}

		case "filters":
			if reflect.ValueOf(factFieldValue).Kind() == reflect.Map {
				filters := factFieldValue.(map[string]interface{})

				if len(filters) > 1 {
					return nil, errors.New("more than one conditions group on root")
				}
				if len(filters) > 0 {
					fragsCondition, err := parseFilters(filters)
					if err != nil {
						return nil, err
					}
					if len(fragsCondition) > 1 {
						return nil, errors.New("more than one root condition")
					}
					condition = fragsCondition[0]
				}
			} else {
				return nil, errors.New("filters is not a map")
			}

		case "comment":
			if reflect.ValueOf(factFieldValue).Kind() == reflect.String {
				comment = factFieldValue.(string)
			} else {
				return nil, errors.New("comment is not a string")
			}

		default:
			return nil, errors.New("unknown field " + factFieldKey)
		}
	}

	if source != "" {
		fact := engine.Fact{
			Name:           factName,
			Model:          model,
			AdvancedSource: source,
		}

		return &fact, nil
	}

	if term == "" {
		return nil, errors.New("intent Term not found")
	}
	intentFragment, err := engine.GetIntentFragment(intentOperator)
	if err != nil {
		return nil, errors.New("intent Operator not found : " + intentOperator)
	}
	if term[:2] == "${" {
		intentFragment.Script = true
		script := term[2 : len(term)-1]
		intentFragment.Term = script
	} else {
		intentFragment.Term = term
	}

	fact := engine.Fact{
		Name:        factName,
		Model:       model,
		Intent:      intentFragment,
		Dimensions:  dimensions,
		Condition:   condition,
		Restitution: restitutions,
		Comment:     comment,
	}

	return &fact, nil
}

func parseFilters(filters map[string]interface{}) ([]engine.ConditionFragment, error) {

	fragments := make([]engine.ConditionFragment, 0)

	for key, value := range filters {
		switch key {
		case "conditions":
			if reflect.ValueOf(value).Kind() != reflect.Slice {
				return nil, errors.New("conditions value is not a slice")
			}
			leafFrags, err := getConditions(value.([]interface{}))
			if err != nil {
				return nil, err
			}
			fragments = append(fragments, leafFrags...)

		case "and":
			if reflect.ValueOf(value).Kind() != reflect.Map {
				return nil, errors.New("and value is not a map or has no children")
			}
			frags, err := parseFilters(value.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			and, err := engine.GetBooleanFragment(engine.And.String())
			if err != nil {
				return nil, err
			}
			and.Fragments = frags
			fragments = append(fragments, and)

		case "or":
			if reflect.ValueOf(value).Kind() != reflect.Map {
				return nil, errors.New("or value is not a map or has no children")
			}
			frags, err := parseFilters(value.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			or, err := engine.GetBooleanFragment(engine.Or.String())
			if err != nil {
				return nil, err
			}
			or.Fragments = frags
			fragments = append(fragments, or)

		case "not":
			if reflect.ValueOf(value).Kind() != reflect.Map {
				return nil, errors.New("not value is not a map or has no children")
			}
			frags, err := parseFilters(value.(map[string]interface{}))
			if err != nil {
				return nil, err
			}
			not, err := engine.GetBooleanFragment(engine.Not.String())
			if err != nil {
				return nil, err
			}
			not.Fragments = frags
			fragments = append(fragments, not)

		default:
			return nil, errors.New("unknown config key " + key)
		}
	}

	return fragments, nil
}

func getConditions(conditionsStr []interface{}) ([]engine.ConditionFragment, error) {

	lexerC := lexer.L()
	leafConditions := make([]engine.ConditionFragment, 0)

	if len(conditionsStr) > 0 {
		for _, conditionStr := range conditionsStr {
			if reflect.ValueOf(conditionStr).Kind() != reflect.String {
				return nil, errors.New("condition is not a string")
			}
			nodes, _ := lexerC.Ast.Parsewith(lexerC.Parser, parsec.NewScanner([]byte(conditionStr.(string))))
			if nodes == nil {
				return nil, errors.New("cannot parse expression : " + conditionStr.(string))
			}
			leafCondition, err := astToLeafCondition(nodes)
			if err != nil {
				return nil, errors.New("cannot convert condition AST in condition fragment")
			}
			leafConditions = append(leafConditions, leafCondition)
		}
	}

	return leafConditions, nil
}

func astToLeafCondition(condition parsec.Queryable) (*engine.LeafConditionFragment, error) {

	var frag engine.LeafConditionFragment

	filterParts := condition.GetChildren()
	switch condition.GetName() {
	case "EXISTS":
		filterFrag, err := engine.GetLeafConditionFragment(engine.Exists.String())
		if err != nil {
			return nil, errors.New("fragment does not exists")
		}
		field := filterParts[1].GetValue()
		filterFrag.Field = field

		return filterFrag, nil

	case "BETWEEN":
		filterFrag, err := engine.GetLeafConditionFragment(engine.Between.String())
		if err != nil {
			return nil, errors.New("fragment does not exists")
		}
		field := filterParts[2].GetValue()
		value := filterParts[0].GetValue()
		value2 := filterParts[4].GetValue()
		filterFrag.Field = field
		filterFrag.Value = value
		filterFrag.Value2 = value2

		return filterFrag, nil

	case "COMPARE":
		filterFrag, err := engine.GetLeafConditionFragment(engine.For.String())
		if err != nil {
			return nil, errors.New("fragment does not exists")
		}
		field := filterParts[0].GetValue()
		value := filterParts[2].GetValue()
		filterFrag.Field = field
		filterFrag.Value = value

		return filterFrag, nil

	case "SCRIPT":
		filterFrag, err := engine.GetLeafConditionFragment(engine.Script.String())
		if err != nil {
			return nil, errors.New("fragment does not exists")
		}
		script := filterParts[1].GetValue()
		filterFrag.Field = script

		return filterFrag, nil

	default:
		zap.L().Info("WARN: Unknown condition type from lexer", zap.String("name", condition.GetName()))

	}
	return &frag, nil
}
