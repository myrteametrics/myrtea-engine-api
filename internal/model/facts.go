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

type FactHitsReq struct {
	Params
	FactId              int64 `json:"factId"`
	Nhit                int   `json:"nhit,omitempty"`
	Offset              int   `json:"offset,omitempty"`
	SituationId         int64 `json:"situationId,omitempty"`
	SituationInstanceId int64 `json:"situationInstanceId,omitempty"`
	Debug               *bool `json:"debug,omitempty"`
	HitsOnly            bool  `json:"hitsOnly,omitempty"` // If true, forces f.Intent.Operator = engine.Select
}

// Validate checks if the FactHitsReq request is valid and transforms some parameters
func (r *FactHitsReq) Process() error {
	if err := r.validateFields(); err != nil {
		return err
	}
	return r.Params.Process()
}

// validateBasicFields validates the basic fields of the request
func (r *FactHitsReq) validateFields() error {
	if r.FactId <= 0 {
		return errors.New("factId parameter is required and must be a positive integer")
	}

	if r.Nhit < 0 {
		return errors.New("nhit cannot be negative")
	}

	if r.Offset < 0 {
		return errors.New("offset cannot be negative")
	}

	if r.Debug == nil {
		debug := true
		r.Debug = &debug
	}
	return nil
}

type Params struct {
	FactParams map[string]interface{} `json:"factParameters,omitempty"`
}

// Validate validates and transforms all fact parameters
func (p *Params) Process() error {
	if p.FactParams == nil {
		p.FactParams = make(map[string]interface{})
	}

	for k, v := range p.FactParams {
		if err := p.transform(k, v); err != nil {
			return err
		}
	}
	return nil
}

// transform validates and transforms a single parameter
func (p *Params) transform(key string, val interface{}) error {
	if s, ok := val.(string); ok {
		return p.transformStr(key, s)
	}
	return p.checkSize(key, val)
}

// transformStr handles string parameter transformation
func (p *Params) transformStr(key, str string) error {
	// Skip RFC3339 dates
	if _, err := time.Parse(time.RFC3339, str); err == nil {
		return nil
	}

	// Evaluate expression
	parsed, err := expression.Process(expression.LangEval, str, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("param %s eval failed: %w", key, err)
	}

	if err := p.checkSize(key, parsed); err != nil {
		return err
	}

	// Update with evaluated value
	p.FactParams[key] = parsed
	return nil
}

// checkSize validates collection size limits
func (p *Params) checkSize(key string, val interface{}) error {
	const maxSize = 500

	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		if v.Len() > maxSize {
			return fmt.Errorf("param %s exceeds max size %d", key, maxSize)
		}
	}
	return nil
}
