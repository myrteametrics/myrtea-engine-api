package metadata

// MetaData struct to represent a situation metadata
type MetaData struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	RuleID      int64       `json:"ruleId"`
	RuleVersion int64       `json:"ruleVersion"`
	CaseName    string      `json:"caseName"`
}
