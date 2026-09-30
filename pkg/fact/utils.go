package fact

import (
	"encoding/json"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/baseline"

	"go.uber.org/zap"
)

// MatrixProfileKnowledgeKey is the key under which matrix profile results are exposed to the
// rule engine. Business rule conditions address them as
// `_baseline_mp.<definitionName>.<field>`, for instance:
//
//	_baseline_mp.trafic.status != "unavailable" && _baseline_mp.trafic.matchingScore < 50
//
// The leading underscore keeps it out of the way of fact and situation parameter names, which
// share the same flat namespace.
const MatrixProfileKnowledgeKey = "_baseline_mp"

// GetMatrixProfileResults fetches the matrix profile results (status + matching score) for
// every matrix profile definition linked to the given situation instance, shaped for
// injection in the rule engine knowledge base under MatrixProfileKnowledgeKey.
//
// expressionFacts carries the current tick's evaluated expression facts, keyed by name: a
// "fact_expression" definition needs its current point before situation_history_v5 holds it,
// since the row for this tick is only written once its metadata is known.
//
// Returns nil when the plugin is not loaded or the call fails: rule conditions referencing an
// absent key evaluate to false, which is the intended degraded behaviour. Note that a
// definition whose computation failed is not absent — the plugin reports it explicitly with
// baseline.StatusUnavailable, so the failure stays visible to rules.
func GetMatrixProfileResults(situationID int64, situationInstanceID int64, ti time.Time, expressionFacts map[string]interface{}) map[string]interface{} {
	pluginBaseline, err := baseline.P()
	if err != nil {
		return nil
	}

	currentExpressionFacts := make(map[string]float64, len(expressionFacts))
	for k, v := range expressionFacts {
		if f, ok := v.(float64); ok {
			currentExpressionFacts[k] = f
		}
	}

	values, err := pluginBaseline.BaselineService.GetMatrixProfileResults(situationID, situationInstanceID, ti, currentExpressionFacts)
	if err != nil {
		zap.L().Error("Cannot fetch matrix profile results",
			zap.Int64("situationID", situationID),
			zap.Int64("situationInstanceID", situationInstanceID), zap.Error(err))
		return nil
	}
	if len(values) == 0 {
		return nil
	}

	// Expose the results under their JSON names, the way fact results are exposed by
	// reader.Item.ToAbstractMap. Rule conditions are written against those names
	// (`matchingScore`, not `MatchingScore`), and gval resolves struct fields by their Go
	// name, so the results have to be turned into plain maps first.
	raw, err := json.Marshal(values)
	if err != nil {
		zap.L().Error("Cannot marshal matrix profile results", zap.Error(err))
		return nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		zap.L().Error("Cannot unmarshal matrix profile results", zap.Error(err))
		return nil
	}
	return out
}
