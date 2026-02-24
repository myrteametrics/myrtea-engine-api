package search

import (
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/baseline"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/reader"
)

// QueryResult is a struct used to represent a query result
type QueryResult []SituationHistoryRecords

// SituationHistoryRecords struct used to group situation history records by datetime
type SituationHistoryRecords struct {
	DateTime   time.Time                `json:"dateTime"`
	Situations []SituationHistoryRecord `json:"situations"`
}

// SituationHistoryRecord struct used to represent a situation history record
type SituationHistoryRecord struct {
	SituationID           int64                           `json:"situationId"`
	SituationName         string                          `json:"situationName"`
	SituationInstanceID   int64                           `json:"situationInstanceId"`
	SituationInstanceName string                          `json:"situationInstanceName"`
	Calendar              *SituationHistoryCalendarRecord `json:"calendar,omitempty"`
	// InCalendarPeriod indicates whether the historical record's timestamp falls within the calendar period.
	InCalendarPeriod *bool `json:"inCalendarPeriod,omitempty"`
	// IsCurrentlyOutsideCalendarPeriod indicates, at the time of retrieval (real-time), whether
	// the current moment is outside the calendar period. Unlike InCalendarPeriod, this is not
	// derived from the historical record's timestamp but from time.Now() at the moment of the request.
	IsCurrentlyOutsideCalendarPeriod *bool                  `json:"isCurrentlyOutsideCalendarPeriod,omitempty"`
	Parameters                       map[string]interface{} `json:"parameters,omitempty"`
	ExpressionFacts                  map[string]interface{} `json:"expressionFacts,omitempty"`
	MetaData                         map[string]interface{} `json:"metaDatas,omitempty"`
	Facts                            []FactHistoryRecord    `json:"facts,omitempty"`
	DateTime                         time.Time              `json:"dateTime"`
}

// FactHistoryRecord struct to represent a fact history record
type FactHistoryRecord struct {
	DateTime  time.Time                         `json:"dateTime"`
	FactID    int64                             `json:"factId"`
	FactName  string                            `json:"factName"`
	Value     interface{}                       `json:"value,omitempty"`
	DocCount  interface{}                       `json:"docCount,omitempty"`
	Buckets   map[string][]*reader.Item         `json:"buckets,omitempty"`
	Baselines map[string]baseline.BaselineValue `json:"baselines,omitempty"`
}

type SituationHistoryCalendarRecord struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Timezone    string `json:"timezone"`
}
