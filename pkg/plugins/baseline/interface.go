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

// MatrixProfileResult carries the status and matching percentage computed by the
// matrix profile pipeline for a single matrix profile definition, keyed by definition name.
type MatrixProfileResult struct {
	Status                   string    `json:"status,omitempty"`
	MatchingScore            float64   `json:"matchingScore,omitempty"`
	MatchingScoreTheoretical float64   `json:"matchingScoreTheoretical,omitempty"`
	BestMatchTime            time.Time `json:"bestMatchTime,omitempty"`
}
