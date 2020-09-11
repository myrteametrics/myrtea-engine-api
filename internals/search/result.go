package search

import (
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v4/plugins/baseline"
)

// QueryResult is a struct used to represent a query result
type QueryResult []SituationHistoryRecords

// SituationHistoryRecords struct used to group situation history records by datetime
type SituationHistoryRecords struct {
	DateTime   time.Time                `json:"dateTime"`
	Situations []SituationHistoryRecord `json:"situations"`
}

//SituationHistoryRecord struct used to represent a situation history record
type SituationHistoryRecord struct {
	SituationID           int64                  `json:"situationId"`
	SituationName         string                 `json:"situationName"`
	SituationInstanceID   int64                  `json:"situationInstanceId"`
	SituationInstanceName string                 `json:"situationInstanceName"`
	Parameters            map[string]interface{} `json:"parameters,omitempty"`
	ExpressionFacts       map[string]interface{} `json:"expressionFacts,omitempty"`
	MetaData              map[string]interface{} `json:"metaDatas,omitempty"`
	Facts                 []FactHistoryRecord    `json:"facts,omitempty"`
	DateTime              time.Time              `json:"dateTime"`
}

//FactHistoryRecord struct to represent a fact history record
type FactHistoryRecord struct {
	DateTime  time.Time                         `json:"dateTime"`
	FactID    int64                             `json:"factId"`
	FactName  string                            `json:"factName"`
	Value     interface{}                       `json:"value,omitempty"`
	DocCount  interface{}                       `json:"docCount,omitempty"`
	Buckets   map[string][]*reader.Item         `json:"buckets,omitempty"`
	Baselines map[string]baseline.BaselineValue `json:"baselines,omitempty"`
}
