package model

import (
	"bytes"
	"encoding/json"
)

// SortOptionOrder indicate if the sorting is ASC or DESC
type SortOptionOrder int

const (
	// Asc sort order
	Asc SortOptionOrder = iota + 1
	// Desc sort order
	Desc
)

func (s SortOptionOrder) String() string {
	if level, ok := sortOptionOrderToString[s]; ok {
		return level
	}
	return ""
}

// ToSortOptionOrder get the SortOptionOrder from is string representation
func ToSortOptionOrder(s string) SortOptionOrder {
	if level, ok := sortOptionOrderToID[s]; ok {
		return level
	}
	return 0
}

var sortOptionOrderToString = map[SortOptionOrder]string{
	Asc:  "asc",
	Desc: "desc",
}

var sortOptionOrderToID = map[string]SortOptionOrder{
	"asc":  Asc,
	"desc": Desc,
}

// MarshalJSON marshals the enum as a quoted json string
func (s SortOptionOrder) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(sortOptionOrderToString[s])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (s *SortOptionOrder) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Info' in this case.
	*s = sortOptionOrderToID[j]
	return nil
}
