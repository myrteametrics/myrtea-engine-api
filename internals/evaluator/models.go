package evaluator

import (
	"time"

	"github.com/myrteametrics/myrtea-sdk/v4/ruleeng"
)

// SituationToEvaluate is used to notify a rule engine instance that a situation must me evaluated
type SituationToEvaluate struct {
	ID                 int64
	TS                 time.Time
	TemplateInstanceID int64
	Facts              []int64
	Parameters         map[string]string
}

// EvaluatedSituation represents the evaluation of a situation
type EvaluatedSituation struct {
	ID                 int64
	TS                 time.Time
	TemplateInstanceID int64
	Agenda             []ruleeng.Action
}
