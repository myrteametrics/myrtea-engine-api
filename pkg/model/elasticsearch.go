package model

type ElasticSearchConfig struct {
	Id              int64    `json:"id"`
	Name            string   `json:"name"`
	URLs            []string `json:"urls"`
	Default         bool     `json:"default"`
	ExportActivated bool     `json:"exportActivated"`
}
