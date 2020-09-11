package models

import (
	"fmt"
	"time"
)

//FactValue the current value of a fact
type FactValue interface {
	String() string
	SetCurrent(bool)
	GetType() string
	GetDeepness() int32
}

//*****************************************************************************

//SingleValue represents a fact with a single value
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

//NotSupportedValue value for not supported history facts
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

//FrontFactHistory represents the current fact value and its history
type FrontFactHistory struct {
	ID           int64                   `json:"id"`
	Name         string                  `json:"name"`
	Type         string                  `json:"type"`
	Deepness     int32                   `json:"deepness"`
	CurrentValue FactValue               `json:"currentValue"`
	History      map[time.Time]FactValue `json:"history"`
}
