package models

type VariablesConfig struct {
	Id    int64  `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Scope string `json:"scope"`
}
