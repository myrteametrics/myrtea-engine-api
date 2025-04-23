package models

type ConnectorConfig struct {
	Id          int64       `json:"id"`
	Name        string      `json:"name"`
	ConnectorId string      `json:"connectorId"`
	Current     interface{} `json:"current"`
}
