package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"time"
)

// FactValue the current value of a fact
type FactValue interface {
	String() string
	SetCurrent(bool)
	GetType() string
	GetDeepness() int32
}

//*****************************************************************************

// SingleValue represents a fact with a single value
type SingleValue struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	IsCurrent bool        `json:"isCurrent"`
}

func (factValue *SingleValue) String() string {
	return fmt.Sprintf("Value: %f  Key: %v  IsCurrent: %v", factValue.Value, factValue.Key, factValue.IsCurrent)
}

// SetCurrent set the IsCurrent boolean
func (factValue *SingleValue) SetCurrent(isCurrent bool) {
	factValue.IsCurrent = isCurrent
}

// GetType get fact value type
func (factValue *SingleValue) GetType() string {
	return "single"
}

// GetDeepness get deepness
func (factValue *SingleValue) GetDeepness() int32 {
	return 1
}

//*****************************************************************************

// ObjectValue is a factValue implementation dedicated to object facts
type ObjectValue struct {
	Attributes map[string]interface{} `json:"attributes"`
}

func (factValue *ObjectValue) String() string {
	b, _ := json.Marshal(factValue.Attributes)
	return string(b)
}

// SetCurrent set the IsCurrent boolean
func (factValue *ObjectValue) SetCurrent(current bool) {}

// GetType get fact value type
func (factValue *ObjectValue) GetType() string {
	return "object"
}

// GetDeepness get deepness
func (factValue *ObjectValue) GetDeepness() int32 {
	return 0
}

// NotSupportedValue value for not supported history facts
type NotSupportedValue struct {
	IsCurrent bool
}

func (factValue *NotSupportedValue) String() string {
	return "Not supported Fact History"
}

// SetCurrent set the IsCurrent boolean
func (factValue *NotSupportedValue) SetCurrent(current bool) {
	factValue.IsCurrent = current
}

// GetType get fact value type
func (factValue *NotSupportedValue) GetType() string {
	return "not_supported"
}

// GetDeepness get deepness
func (factValue *NotSupportedValue) GetDeepness() int32 {
	return 0
}

//*****************************************************************************

// FrontFactHistory represents the current fact value and its history
type FrontFactHistory struct {
	ID           int64                   `json:"id"`
	Name         string                  `json:"name"`
	Type         string                  `json:"type"`
	Deepness     int32                   `json:"deepness"`
	CurrentValue FactValue               `json:"currentValue"`
	History      map[time.Time]FactValue `json:"history"`
}

type FactHitsRequest struct {
	FactId              int64                  `json:"factId"`
	Nhit                int                    `json:"nhit,omitempty"`
	Offset              int                    `json:"offset,omitempty"`
	SituationId         int64                  `json:"situationId,omitempty"`
	SituationInstanceId int64                  `json:"situationInstanceId,omitempty"`
	Debug               bool                   `json:"debug,omitempty"`
	FactParameters      map[string]interface{} `json:"factParameters,omitempty"`
	HitsOnly            bool                   `json:"hitsOnly,omitempty"` // If true, forces f.Intent.Operator = engine.Select
}

// Validate checks if the FactHitsRequest is valid
func (r *FactHitsRequest) Validate() error {

	if r.FactId <= 0 {
		return errors.New("factId parameter is required and must be a positive integer")
	}

	if r.Nhit < 0 {
		return errors.New("nhit cannot be negative")
	}

	if r.Offset < 0 {
		return errors.New("offset cannot be negative")
	}

	for key, value := range r.FactParameters {
		const maxSliceSize = 500

		strVal, ok := value.(string)
		if ok {
			parsed, err := expression.Process(expression.LangEval, strVal, map[string]interface{}{})
			if err != nil {
				return fmt.Errorf("parameters: the value of the key %s could not be evaluated: %s", key, err.Error())
			}

			switch parsedVal := parsed.(type) {
			case []interface{}:
				if len(parsedVal) > maxSliceSize {
					return fmt.Errorf("parameters: the slice for key %s exceeds the maximum size of %d elements", key, maxSliceSize)
				}
			case []string:
				if len(parsedVal) > maxSliceSize {
					return fmt.Errorf("parameters: the string slice for key %s exceeds the maximum size of %d elements", key, maxSliceSize)
				}
			case []int:
				if len(parsedVal) > maxSliceSize {
					return fmt.Errorf("parameters: the int slice for key %s exceeds the maximum size of %d elements", key, maxSliceSize)
				}
			case []float64:
				if len(parsedVal) > maxSliceSize {
					return fmt.Errorf("parameters: the float slice for key %s exceeds the maximum size of %d elements", key, maxSliceSize)
				}
			case map[string]interface{}:
				if len(parsedVal) > maxSliceSize {
					return fmt.Errorf("parameters: the map for key %s exceeds the maximum size of %d elements", key, maxSliceSize)
				}
			}

			r.FactParameters[key] = parsed
			continue
		}

		switch val := value.(type) {
		case []interface{}:
			if len(val) > maxSliceSize {
				return fmt.Errorf("parameters: the slice for key %s exceeds the maximum size of %d elements", key, maxSliceSize)
			}
		case []string:
			if len(val) > maxSliceSize {
				return fmt.Errorf("parameters: the string slice for key %s exceeds the maximum size of %d elements", key, maxSliceSize)
			}
		case []int:
			if len(val) > maxSliceSize {
				return fmt.Errorf("parameters: the int slice for key %s exceeds the maximum size of %d elements", key, maxSliceSize)
			}
		case []float64:
			if len(val) > maxSliceSize {
				return fmt.Errorf("parameters: the float slice for key %s exceeds the maximum size of %d elements", key, maxSliceSize)
			}
		case map[string]interface{}:
			if len(val) > maxSliceSize {
				return fmt.Errorf("parameters: the map for key %s exceeds the maximum size of %d elements", key, maxSliceSize)
			}
		}

	}

	return nil
}
