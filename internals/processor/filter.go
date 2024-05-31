package processor

import (
	"strings"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	sdk_models "github.com/myrteametrics/myrtea-sdk/v5/models"
	"go.uber.org/zap"
)

// ReceiveObjects ...
func ReceiveObjects(factObjectName string, documents []sdk_models.Document) error {
	factObject, found, err := fact.R().GetByName(factObjectName)
	if err != nil {
		return err
	}
	if !found {
		zap.L().Error("Fact not found", zap.String("name", factObjectName))
		return nil
	}
	if !factObject.IsObject {
		zap.L().Warn("Fact is not an object fact", zap.String("name", factObjectName))
		return nil
	}

	objects := make([]map[string]interface{}, 0)
	for _, document := range documents {
		if !objectFilterKeep(factObject, document) {
			continue
		}

		source := document.Source
		source["id"] = document.ID

		// Not working ATM (Field are flatten + array theorically not supported)
		// source = filterSource(factObject, source)
		objects = append(objects, source)
	}

	return evaluateFactObjects(factObject, objects)
}

func filterSource(f engine.Fact, source map[string]interface{}) map[string]interface{} {
	if f.Intent.Term == "*" {
		return source
	}

	filteredSource := make(map[string]interface{}, 0)
	for _, path := range strings.Split(f.Intent.Term, ",") {
		field, found := findField(source, path)
		if !found {
			// err
			return nil
		}
		filteredSource[path] = field

	}
	return filteredSource
}

func objectFilterKeep(f engine.Fact, document sdk_models.Document) bool {
	return applyCondition(f.Condition, document.Source)
}

func applyCondition(c engine.ConditionFragment, source map[string]interface{}) bool {
	if c == nil {
		return true
	}

	switch frag := c.(type) {
	case *engine.BooleanFragment:
		results := make([]bool, 0)
		for _, subc := range frag.Fragments {
			results = append(results, applyCondition(subc, source))
		}
		switch frag.Operator {
		case engine.And:
			return checkAll(results)
		case engine.Or:
			return checkAny(results)
		case engine.Not:
			return !checkAny(results)
		default:
			zap.L().Warn("Operator unknown", zap.Any("frag", frag))
		}

	case *engine.LeafConditionFragment:
		switch frag.Operator {
		case engine.Exists:
			return checkExists(source, frag.Field)
		case engine.For:
			return checkFor(source, frag.Field, frag.Value)
		case engine.From:
			return checkRange(source, frag.Field, frag.Value, -1)
		case engine.To:
			return checkRange(source, frag.Field, -1, frag.Value2)
		case engine.Between:
			return checkRange(source, frag.Field, frag.Value, frag.Value2)
		case engine.Script:
			zap.L().Warn("Script not supported")
		}
	default:
		zap.L().Warn("Condition type unknown", zap.Any("frag", frag))
	}

	return false
}

func findField(source interface{}, term string) (interface{}, bool) {
	termParts := strings.Split(term, ".")
	subSource := source
	for _, p := range termParts {
		s, ok := subSource.(map[string]interface{})
		if !ok {
			return nil, false
		}
		subSource = s[p]
	}
	_, ok := subSource.(map[string]interface{})
	if ok {
		return nil, false
	}
	return subSource, true
}

func checkExists(source map[string]interface{}, term string) bool {
	_, found := findField(source, term)
	if found {
		return true
	}
	return false
}

func checkFor(source map[string]interface{}, term string, value interface{}) bool {
	field, found := findField(source, term)
	if !found {
		return false
	}
	if field == value {
		return true
	}
	return false
}

func checkRange(source map[string]interface{}, term string, value interface{}, value2 interface{}) bool {
	// TODO: Implements range support (to, from, between)
	// Beware of values types (int,int32,int64 ? date with specific format ? etc.)
	// Beware of multi-types comparison int64 >= float64 (etc.)
	zap.L().Warn("Range condition are currently not supported and are therefore ignored")
	return true
}

func checkAny(a []bool) bool {
	for _, b := range a {
		if b {
			return true
		}
	}
	return false
}

func checkAll(a []bool) bool {
	for _, b := range a {
		if !b {
			return false
		}
	}
	return true
}
