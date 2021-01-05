package search

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"go.uber.org/zap"
)

// Query is a struct used to represent a query
type Query struct {
	SituationID         int64         `json:"situationId"`
	SituationInstanceID int64         `json:"situationInstanceId"`
	Time                time.Time     `json:"time"`
	Start               time.Time     `json:"start"`
	End                 time.Time     `json:"end"`
	Range               time.Duration `json:"range"`
	Facts               interface{}   `json:"facts"`
	ExpressionFacts     interface{}   `json:"expressionFacts"`
	MetaData            interface{}   `json:"metaDatas"`
	Parameters          interface{}   `json:"parameters"`
	DownSampling        DownSampling  `json:"downSampling"`
}

//DownSampling downscale definition
type DownSampling struct {
	Granularity        time.Duration `json:"granularity"`
	GranularitySpecial string        `json:"granularitySpecial"`
	Operation          string        `json:"operation"`
}

// ErrDatabase wraps a database error
type ErrDatabase struct {
	message string
}

// NewErrDatabase returns a new ErrDatabase
func NewErrDatabase(message string) ErrDatabase {
	return ErrDatabase{message: message}
}

func (e ErrDatabase) Error() string {
	return e.message
}

// ErrResourceNotFound wraps a ressource not found error
type ErrResourceNotFound struct {
	resourceType string
}

// NewErrResourceNotFound returns a new ErrResourceNotFound
func NewErrResourceNotFound(resourceType string) ErrResourceNotFound {
	return ErrResourceNotFound{resourceType: resourceType}
}

func (e ErrResourceNotFound) Error() string {
	return fmt.Sprintf("Resource of type '%s' not found", e.resourceType)
}

// Execute executes the query and return a QueryResult or an error
func (q Query) Execute() (QueryResult, error) {

	s, found, err := situation.R().Get(q.SituationID, groups.GetTokenAllGroups())
	if err != nil {
		zap.L().Error("Cannot retrieve situation", zap.Int64("situationID", q.SituationID), zap.Error(err))
		return nil, NewErrDatabase(err.Error())
	}
	if !found {
		zap.L().Warn("Situation does not exists", zap.Int64("situationID", q.SituationID))
		return nil, NewErrResourceNotFound("situation")
	}
	if !s.IsTemplate && q.SituationInstanceID != 0 {
		zap.L().Warn("Situation does not exists", zap.Int64("situationID", q.SituationID))
		return nil, NewErrResourceNotFound("situation")
	}

	return R().GetSituationHistoryRecords(s, q.SituationInstanceID, q.Time, q.Start, q.End, q.Facts, q.ExpressionFacts, q.MetaData, q.Parameters, q.DownSampling)
}

// UnmarshalJSON unmarshals a quoted json string to a valid DownSampling struct
func (d *DownSampling) UnmarshalJSON(data []byte) error {
	type Alias DownSampling
	aux := &struct {
		Granularity string `json:"granularity"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Granularity == "year" || aux.Granularity == "quarter" || aux.Granularity == "month" || aux.Granularity == "day" ||
		aux.Granularity == "hour" || aux.Granularity == "minute" || aux.Granularity == "second" {
		d.GranularitySpecial = aux.Granularity
	} else {
		g, err := time.ParseDuration(strings.ToLower(aux.Granularity))
		if err != nil {
			return err
		}
		d.Granularity = g
	}

	d.Operation = strings.ToLower(d.Operation)
	if d.Operation != "first" && d.Operation != "latest" && d.Operation != "sum" && d.Operation != "max" && d.Operation != "min" && d.Operation != "avg" {
		return errors.New("Unknown downSampling operation")
	}

	return nil
}

// UnmarshalJSON unmarshals a quoted json string to a valid Query struct
func (q *Query) UnmarshalJSON(data []byte) error {
	type Alias Query
	aux := &struct {
		Range string `json:"range"`
		*Alias
	}{
		Alias: (*Alias)(q),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(aux.Range) > 0 {
		r, err := time.ParseDuration(strings.ToLower(aux.Range))
		if err != nil {
			return err
		}
		q.Range = r
	}

	timeIsZero, startIsZero, endIsZero := q.Time.IsZero(), q.Start.IsZero(), q.End.IsZero()

	if !timeIsZero && q.DownSampling.Granularity != 0 {
		return errors.New("The 'time' parameter is not compatible with downSampling")
	}

	if !timeIsZero && (!startIsZero || !endIsZero || q.Range != 0) {
		return errors.New("The 'time' parameter is not compatible with the parameters 'start', 'end' and 'range'")
	}

	if timeIsZero && q.Range == 0 && (startIsZero || endIsZero) {
		return errors.New("Missing 'start' or 'end' parameter")
	}

	if q.Range != 0 && (startIsZero && endIsZero) {
		return errors.New("To use the 'range' parameter the parameter 'start' or 'end' should be settled")
	}

	if q.Range != 0 && (!startIsZero && !endIsZero) {
		return errors.New("To use the 'range' parameter only one of the parameters 'start' and 'end' should be settled")
	}

	if q.Range != 0 && startIsZero {
		q.Start = q.End.Add(-q.Range)
	} else if q.Range != 0 && endIsZero {
		q.End = q.Start.Add(q.Range)
	}

	if startIsZero && q.Range == 0 && q.End.After(q.Start) {
		return errors.New("The 'end' dateTime should be after the start dateTime")
	}

	err := validateSourceFilterParameter(&q.Facts, "facts")
	if err != nil {
		return err
	}
	err = validateSourceFilterParameter(&q.MetaData, "metaDatas")
	if err != nil {
		return err
	}
	err = validateSourceFilterParameter(&q.Parameters, "parameters")
	if err != nil {
		return err
	}
	return nil
}

func validateSourceFilterParameter(value *interface{}, name string) error {
	if *value == nil {
		*value = true
	} else {
		switch v := (*value).(type) {
		case bool, string:
		case []interface{}:
			val := make([]string, 0)
			for _, elem := range v {
				v, ok := elem.(string)
				if !ok {
					return fmt.Errorf("Unexpected value for '%s' parameter, expecting bool, string or []string", name)
				}
				val = append(val, v)
			}
			*value = val
		default:
			return fmt.Errorf("Unexpected value for '%s' parameter, expecting bool, string or []string", name)
		}
	}
	return nil
}

func parseDurationPart(value string, unit time.Duration) time.Duration {
	if len(value) == 0 {
		return 0
	}
	if parsed, err := strconv.ParseFloat(value[:len(value)-1], 64); err == nil {
		return time.Duration(float64(unit) * parsed)
	}
	return 0
}
