package models

type ExternalConfig struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}
