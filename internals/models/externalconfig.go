package models

type ExternalConfig struct {
	Id   int64       `json:"id"`
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}
