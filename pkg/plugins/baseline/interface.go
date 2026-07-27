package baseline

import (
	"time"
)

// Baseline is the interface that we're exposing as a plugin.
type BaselineService interface {
	// GetBaselineValue(id int64, factID int64, situationID int64, situationInstanceID int64, time time.Time) (BaselineValue, error)
	GetBaselineValues(id int64, factID int64, situationID int64, situationInstanceID int64, ti time.Time) (map[string]BaselineValue, error)
	BuildBaselineValues(baselineID int64) error
	GetMatrixProfileResults(situationID int64, situationInstanceID int64, ti time.Time) (map[string]MatrixProfileResult, error)
}

type BaselineValue struct {
	Time       time.Time `json:"time,omitempty"`
	Value      float64   `json:"value,omitempty"`
	ValueLower float64   `json:"valueLower,omitempty"`
	ValueUpper float64   `json:"valueUpper,omitempty"`
	Avg        float64   `json:"avg,omitempty"`
	Std        float64   `json:"std,omitempty"`
	Median     float64   `json:"median,omitempty"`
}

const (
	// StatusUnavailable is reported instead of a best match status when the matrix profile
	// could not be computed at all (no reference yet, history too short, ...).
	StatusUnavailable = "unavailable"
	// ScoreUnavailable is the sentinel score reported alongside StatusUnavailable. Real scores
	// are clamped to [0, 100], so a negative value cannot be mistaken for a genuine one.
	//
	// Business rules comparing a score to a low threshold must guard on the status, since
	// `matchingScore < 50` is true for an unavailable result:
	//
	//	_baseline_mp.x.status != "unavailable" && _baseline_mp.x.matchingScore < 50
	ScoreUnavailable = -1.0
)

// MatrixProfileResult carries the status and matching percentage computed by the
// matrix profile pipeline for a single matrix profile definition, keyed by definition name.
//
// None of the fields carry omitempty on purpose: these results are addressed by path in
// business rule conditions (`_baseline_mp.<name>.status`), and a key missing from the JSON
// makes the whole condition evaluate to false with no trace, because the rule engine
// discards resolution errors (see ruleeng.Case.Evaluate).
type MatrixProfileResult struct {
	Status                   string    `json:"status"`
	MatchingScore            float64   `json:"matchingScore"`
	MatchingScoreTheoretical float64   `json:"matchingScoreTheoretical"`
	BestMatchTime            time.Time `json:"bestMatchTime"`
}
