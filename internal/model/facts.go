package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"reflect"
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

// Validate checks if the FactHitsRequest request is valid and transforms some parameters
func (r *FactHitsRequest) ValidateParseParam() error {
	if err := r.validateBasicFields(); err != nil {
		return err
	}

	return r.validateAndTransformFactParameters()
}

// validateBasicFields validates the basic fields of the request
func (r *FactHitsRequest) validateBasicFields() error {
	if r.FactId <= 0 {
		return errors.New("factId parameter is required and must be a positive integer")
	}

	if r.Nhit < 0 {
		return errors.New("nhit cannot be negative")
	}

	if r.Offset < 0 {
		return errors.New("offset cannot be negative")
	}

	return nil
}

// validateAndTransformFactParameters validates and transforms all fact parameters
func (r *FactHitsRequest) validateAndTransformFactParameters() error {
	for key, value := range r.FactParameters {
		if err := r.validateAndTransformParameter(key, value); err != nil {
			return err
		}
	}
	return nil
}

// validateAndTransformParameter validates and transforms a specific parameter
func (r *FactHitsRequest) validateAndTransformParameter(key string, value interface{}) error {
	if strVal, ok := value.(string); ok {
		return r.validateAndTransformStringParameter(key, strVal)
	}

	return r.checkCollectionSize(key, value)
}

// validateAndTransformStringParameter handles, validates and transforms a string parameter
func (r *FactHitsRequest) validateAndTransformStringParameter(key, strVal string) error {
	// Ignore if it's a valid date in RFC3339 format
	if _, err := time.Parse(time.RFC3339, strVal); err == nil {
		return nil
	}

	// Evaluate expression
	parsed, err := expression.Process(expression.LangEval, strVal, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("parameters: the value of the key %s could not be evaluated: %s", key, err.Error())
	}

	if err := r.checkCollectionSize(key, parsed); err != nil {
		return err
	}

	// Update parameter with evaluated value (transformation)
	r.FactParameters[key] = parsed
	return nil
}

// checkCollectionSize checks if the size of a collection exceeds the maximum limit
func (r *FactHitsRequest) checkCollectionSize(key string, value interface{}) error {
	const maxSliceSize = 500

	v := reflect.ValueOf(value)
	kind := v.Kind()

	if kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
		if v.Len() > maxSliceSize {
			return fmt.Errorf("parameters: the collection for key %s exceeds the maximum size of %d elements",
				key, maxSliceSize)
		}
	}

	return nil
}
