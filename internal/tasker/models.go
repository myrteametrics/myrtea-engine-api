package tasker

import (
	"time"

	"github.com/myrteametrics/myrtea-sdk/v5/ruleeng"
)

// TaskBatch batch of action to be performed
type TaskBatch struct {
	Context map[string]interface{}
	Agenda  []ruleeng.Action
}

// ContextData struct to represent the data related to the context in an action perform
type ContextData struct {
	RuleID                      int64
	RuleVersion                 int64
	CaseName                    string
	SituationID                 int64
	TemplateInstanceID          int64
	TS                          time.Time
	HistorySituationFlattenData map[string]interface{}
	SituationHistoryID          int64
}

func BuildContextData(inputs ...map[string]interface{}) ContextData {

	ctx := ContextData{}
	for _, input := range inputs {
		for key, value := range input {
			switch key {
			case "ruleID":
				ctx.RuleID = value.(int64)
			case "ruleVersion":
				ctx.RuleVersion = value.(int64)
			case "caseName":
				ctx.CaseName = value.(string)
			case "situationID":
				ctx.SituationID = value.(int64)
			case "templateInstanceID":
				ctx.TemplateInstanceID = value.(int64)
			case "ts":
				ctx.TS = value.(time.Time)
			case "historySituationFlattenData":
				ctx.HistorySituationFlattenData = value.(map[string]interface{})
			case "situationHistoryID":
				ctx.SituationHistoryID = value.(int64)
			}
		}
	}
	return ctx
}
