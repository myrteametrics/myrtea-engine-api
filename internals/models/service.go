package models

type ServiceDefinition struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type ServiceStatus struct {
	IsRunning bool `json:"running"`
}

type Service interface {
	GetStatus() ServiceStatus
	Reload(component string) error
}
