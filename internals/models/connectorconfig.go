package models

type ConnectorConfig struct {
	Id          int64       `json:"id"`
	Name        string      `json:"name"`
	ConnectorId string      `json:"connector_id"`
	Current     interface{} `json:"current"`
}
