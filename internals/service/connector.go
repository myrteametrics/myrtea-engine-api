package service

import "github.com/myrteametrics/myrtea-engine-api/v5/internals/models"

type ConnectorService struct {
	Name string
}

func (c *ConnectorService) GetStatus() models.ServiceStatus {
	return models.ServiceStatus{IsRunning: true}
}

func (c *ConnectorService) Reload(component string) error {
	return nil
}
