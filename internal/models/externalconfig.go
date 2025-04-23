package models

import "time"

type ExternalConfig struct {
	Id             int64       `json:"id"`
	Name           string      `json:"name"`
	Data           interface{} `json:"data"`
	CreatedAt      time.Time   `json:"created_at"`
	CurrentVersion bool        `json:"current_version"`
}
