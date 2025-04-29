package calendar

import (
	"bytes"
	"encoding/json"
)

// PeriodStatus indicate if the status of a date in a period
type PeriodStatus int

const (
	// InPeriod period status
	InPeriod PeriodStatus = iota + 1
	// OutOfPeriod period status
	OutOfPeriod
	// NoInfo period status
	NoInfo
)

func (s PeriodStatus) String() string {
	if level, ok := periodStatusToString[s]; ok {
		return level
	}
	return ""
}

// ToPeriodStatus get the PeriodStatus from is string representation
func ToPeriodStatus(s string) PeriodStatus {
	if level, ok := periodStatusToID[s]; ok {
		return level
	}
	return 0
}

var periodStatusToString = map[PeriodStatus]string{
	InPeriod:    "inPeriod",
	OutOfPeriod: "outOfPeriod",
	NoInfo:      "noInfo",
}

var periodStatusToID = map[string]PeriodStatus{
	"inPeriod":    InPeriod,
	"outOfPeriod": OutOfPeriod,
	"noInfo":      NoInfo,
}

// MarshalJSON marshals the enum as a quoted json string
func (s PeriodStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(periodStatusToString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *PeriodStatus) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Info' in this case.
	*s = periodStatusToID[j]
	return nil
}
